package mongoserver

import (
	"context"
	"crudapi/consts"
	"crudapi/grpcapi"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/protobuf/encoding/protojson"
)

type failingExecutionRetry struct {
	RetryCount int
	Function   func(*mongo.Collection, kafka.Message) error
	Param1     *mongo.Collection
	Param2     kafka.Message
}

func prodMessageParse(message []byte) (*grpcapi.ProductMessage, error) {
	var (
		err                error
		crudProductMessage grpcapi.ProductMessage
	)
	err = protojson.Unmarshal(message, &crudProductMessage)
	return &crudProductMessage, err
}

func prodUidParse(message []byte) (*grpcapi.UidParam, error) {
	var (
		err                error
		crudProductMessage grpcapi.UidParam
	)
	err = protojson.Unmarshal(message, &crudProductMessage)
	return &crudProductMessage, err
}

func ProdConsume(kafkaAddr, mongoUri string) {
	var (
		err              error
		bckCtx           = context.Background()
		r                *kafka.Reader
		signalCtx        context.Context
		stop             context.CancelFunc
		kafkaCtx         = context.TODO()
		col              *mongo.Collection
		msg              kafka.Message
		limitExecutions  = make(chan struct{}, consts.KafkaExecutionLimit)
		failedExecutions = make(chan failingExecutionRetry, consts.KafkaExecutionLimit)
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
	go func(chan failingExecutionRetry) {
		var (
			execution failingExecutionRetry
			err       error
		)
		execution = <-failedExecutions
		if execution.RetryCount < consts.MaxFailedExecutionRetry {
			if err = execution.Function(execution.Param1, execution.Param2); err != nil {
				execution.RetryCount += 1
				failedExecutions <- execution
			}
		}
	}(failedExecutions)
	col = MongoConnection.Client.Database(consts.DATABASE).Collection(consts.ProductCollection)
	for {
		msg, err = r.ReadMessage(kafkaCtx)
		if err != nil {
			log.Fatal("could not read message " + err.Error())
			break
		}
		limitExecutions <- struct{}{}
		go func() {
			if err = executeOperations(col, msg); err != nil {
				failedExecutions <- failingExecutionRetry{0, executeOperations, col, msg}
			}
			<-limitExecutions
		}()
	}
}

func updateProdInfoStringToObj(col *mongo.Collection, filter primitive.D) error {
	var (
		updateAsDict primitive.D
		err          error
	)
	updateAsDict = bson.D{{
		Key: "$set",
		Value: bson.D{{
			Key: "productInformation",
			Value: bson.D{{
				Key: "$function",
				Value: bson.M{
					"lang": "js",
					"args": []string{"$productinfo"},
					"body": "function(infoStr) { return JSON.parse(infoStr); }",
				},
			}},
		}},
	}}
	_, err = col.UpdateOne(MongoConnection.Ctx, filter, mongo.Pipeline{updateAsDict})
	if err != nil {
		log.Println("Converting productinfo to json object failed")
		return fmt.Errorf("converting productinfo to json object failed: %v", err)
	}
	return err
}

func executeOperations(col *mongo.Collection, msg kafka.Message) error {
	var (
		err      error
		err1     error
		err2     error
		bsonobj1 *grpcapi.ProductMessage
		bsonobj2 *grpcapi.UidParam
		filter   primitive.D
		update   primitive.D
	)
	log.Println(string(msg.Value))
	if err != nil {
		log.Printf("Bson obj preparation failed using kafka message: %v", err)
		return err
	}
	bsonobj1, err1 = prodMessageParse(msg.Value)
	bsonobj2, err2 = prodUidParse(msg.Value)
	if err1 != nil && err2 != nil {
		log.Printf("Error in Unmarshalling: %v, %v", err1, err2)
		return fmt.Errorf("error in Unmarshalling: %v, %v", err1, err2)
	}
	if bsonobj1 != nil && bsonobj1.GetAction() == grpcapi.Action_INSERT {
		_, err = col.InsertOne(MongoConnection.Ctx, bsonobj1)
		if err != nil {
			log.Printf("Insertion failed: %v", err)
			return fmt.Errorf("insertion failed: %v", err)
		}
		filter = bson.D{{Key: "uid", Value: bsonobj1.GetUid()}}
		return updateProdInfoStringToObj(col, filter)
	} else if bsonobj1 != nil && bsonobj1.GetAction() == grpcapi.Action_UPDATE {
		filter = bson.D{{Key: "uid", Value: bsonobj1.GetUid()}}
		update = bson.D{{
			Key: "$set",
			Value: bson.M{
				"source":      bsonobj1.GetSource(),
				"productinfo": bsonobj1.GetProductInfo(),
			},
		}}
		_, err = col.UpdateOne(MongoConnection.Ctx, filter, update)
		if err != nil {
			log.Printf("Updation failed: %v", err)
			return fmt.Errorf("updation failed: %v", err)
		}
		return updateProdInfoStringToObj(col, filter)
	} else if bsonobj2 != nil && bsonobj2.Action == grpcapi.Action_DELETE {
		filter = bson.D{{Key: "uid", Value: bsonobj2.GetUid()}}
		log.Println(filter)
		_, err = col.DeleteOne(MongoConnection.Ctx, filter)
		if err != nil {
			log.Printf("Deletion failed: %v", err)
			return fmt.Errorf("deletion failed: %v", err)
		}
	} else {
		log.Fatalf("action hasn't implemented for %v, %v", bsonobj1.Action.String(), bsonobj2.Action.String())
	}
	return nil
}
