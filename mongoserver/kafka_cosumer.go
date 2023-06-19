package mongoserver

import (
	"context"
	"crudapi/consts"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
)

func UserConsume(ctx context.Context, kafkaAddr, mongoUri string) {
	var (
		err       error
		bckCtx    = context.Background()
		r         *kafka.Reader
		signalCtx context.Context
		stop      context.CancelFunc
		mongodb   *Connection
		col       *mongo.Collection
		bsonobj   User
		msg       kafka.Message
	)
	r = kafka.NewReader(kafka.ReaderConfig{
		Brokers: strings.Split(kafkaAddr, ","),
		Topic:   consts.TopicUser,
		GroupID: "user-consumerId",
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
		Close(mongodb.Client, mongodb.Ctx)
	}()

	col = MongoConnection.Client.Database(consts.DATABASE).Collection(consts.UserCollection)
	for {
		msg, err = r.ReadMessage(ctx)
		if err != nil {
			log.Fatal("could not read message " + err.Error())
			break
		}
		bsonobj, err = BsonObj(msg.Value)
		if err != nil {
			log.Printf("Bson obj preparation failed using kafka message: %v", err)
			continue
		}
		_, err = col.InsertOne(context.TODO(), bsonobj)
		if err != nil {
			log.Printf("Insertion failed: %v", err)
		}
	}
}
