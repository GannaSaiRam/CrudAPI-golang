package mongoserver

import (
	"context"
	"crudapi/consts"
	"crudapi/grpcapi"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type messageGrpcApi interface {
	*grpcapi.ProductMessage | *grpcapi.UidParam
	GetUid() string
	GetAction() grpcapi.Action
	ProtoReflect() protoreflect.Message
}

func prodMessageParse(message []byte) (*grpcapi.ProductMessage, error) {
	var crudProductMessage grpcapi.ProductMessage
	err := protojson.Unmarshal(message, &crudProductMessage)
	return &crudProductMessage, err
}

func prodUidParse(message []byte) (*grpcapi.UidParam, error) {
	var crudProductMessage grpcapi.UidParam
	err := protojson.Unmarshal(message, &crudProductMessage)
	return &crudProductMessage, err
}

func ProdConsume(kafkaAddr, mongoUri string) {
	var (
		err       error
		err1      error
		err2      error
		bckCtx    = context.Background()
		r         *kafka.Reader
		signalCtx context.Context
		stop      context.CancelFunc
		col       *mongo.Collection
		msg       kafka.Message
		kafkaCtx  = context.TODO()
		bsonobj1  *grpcapi.ProductMessage
		bsonobj2  *grpcapi.UidParam
	)
	r = kafka.NewReader(kafka.ReaderConfig{
		Brokers: strings.Split(kafkaAddr, ","),
		Topic:   consts.TopicProduct,
		GroupID: "products-consumerId",
	})
	if r == nil {
		log.Fatal("Error connecting to kafka brokers")
	}

	signalCtx, stop = signal.NotifyContext(
		bckCtx,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGTSTP,
	)

	// If signal is thrown,
	go func() {
		<-signalCtx.Done()
		stop()
		log.Println("Shutdown signal received")
		r.Close()
		bckCtx.Done()
		Close(MongoConnection.Client, MongoConnection.Ctx)
	}()

	col = MongoConnection.Client.Database(consts.DATABASE).Collection(consts.ProductCollection)
	for {
		msg, err = r.ReadMessage(kafkaCtx)
		if err != nil {
			log.Fatal("could not read message " + err.Error())
			break
		}
		log.Println(string(msg.Value))
		if err != nil {
			log.Printf("Bson obj preparation failed using kafka message: %v", err)
			continue
		}
		bsonobj1, err1 = prodMessageParse(msg.Value)
		bsonobj2, err2 = prodUidParse(msg.Value)
		if err1 != nil && err2 != nil {
			log.Printf("Error in Unmarshalling: %v, %v", err1, err2)
			continue
		}
		log.Printf("%+v, %+v", bsonobj1, bsonobj2)
		if bsonobj1 != nil && bsonobj1.GetAction() == grpcapi.Action_INSERT {
			_, err := col.InsertOne(MongoConnection.Ctx, bsonobj1)
			if err != nil {
				log.Printf("Insertion failed: %v", err)
			}
		} else if bsonobj1 != nil && bsonobj1.GetAction() == grpcapi.Action_UPDATE {
			filter := bson.D{{Key: "uid", Value: bsonobj1.GetUid()}}
			update := bson.M{
				"$set": bson.M{
					"source":      bsonobj1.GetSource(),
					"productinfo": bsonobj1.GetProductInfo(),
				},
			}
			_, err := col.UpdateOne(MongoConnection.Ctx, filter, update)
			if err != nil {
				log.Printf("Updation failed: %v", err)
			}
		} else if bsonobj2 != nil && bsonobj2.Action == grpcapi.Action_DELETE {
			filter := bson.D{{Key: "uid", Value: bsonobj2.GetUid()}}
			log.Println(filter)
			_, err := col.DeleteOne(MongoConnection.Ctx, filter)
			if err != nil {
				log.Printf("Deletion failed: %v", err)
			}
		} else {
			log.Fatalf("action hasn't implemented for %v, %v", bsonobj1.Action.String(), bsonobj2.Action.String())
		}
	}
}
