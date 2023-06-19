package userserver

import (
	"context"
	"crudapi/consts"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

var (
	ProducerConn *kafka.Writer
)

func Producer(kafka_addr string) {
	var (
		err                 error
		dialer              *kafka.Dialer
		conn                *kafka.Conn
		customerTopicConfig kafka.TopicConfig
		stop                context.CancelFunc
		ctx                 context.Context
		bckCtx              = context.Background()
	)
	dialer = &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}
	conn, err = dialer.DialContext(bckCtx, "tcp", strings.Split(kafka_addr, ",")[0])
	if err != nil {
		log.Fatalf("error dialing kafka: %v", err)
	}

	customerTopicConfig = kafka.TopicConfig{
		Topic: consts.TopicUser, NumPartitions: 1, ReplicationFactor: 1,
	}

	err = conn.CreateTopics(customerTopicConfig)
	if err != nil {
		log.Fatalf("error creating topic: %v", err)
	}

	ctx, stop = signal.NotifyContext(
		bckCtx,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGTSTP,
	)

	// If signal is thrown,
	go func() {
		<-ctx.Done()
		stop()
		log.Println("Shutdown signal received")
		conn.Close()
		ProducerConn.Close()
		bckCtx.Done()
	}()
	ProducerConn = kafka.NewWriter(kafka.WriterConfig{
		Brokers: strings.Split(kafka_addr, ","),
		Topic:   consts.TopicUser,
		Dialer:  dialer,
	})
}

func WriteMessage(message []byte, key string) (err error) {
	err = ProducerConn.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(key),
		Value: message,
	})
	if err != nil {
		log.Printf("Error writing %v", message)
	}
	return
}
