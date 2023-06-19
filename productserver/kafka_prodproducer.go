package productserver

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"crudapi/consts"

	"github.com/segmentio/kafka-go"
)

var (
	ProducerConn *kafka.Writer
)

func ProductProducer(kafka_addr string) {
	var (
		err                  error
		dialer               *kafka.Dialer
		conn                 *kafka.Conn
		customerTopicConfigs kafka.TopicConfig
		stop                 context.CancelFunc
		ctx                  context.Context
		bckCtx               = context.Background()
	)

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
		bckCtx.Done()
		if conn != nil {
			conn.Close()
		}
		if ProducerConn != nil {
			ProducerConn.Close()
		}
	}()

	dialer = &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	conn, err = dialer.DialContext(bckCtx, "tcp", strings.Split(kafka_addr, ",")[0])
	if err != nil {
		log.Fatalf("error dialing kafka: %v", err)
	}

	customerTopicConfigs = kafka.TopicConfig{
		Topic: consts.TopicProduct, NumPartitions: 1, ReplicationFactor: 1,
	}

	err = conn.CreateTopics(customerTopicConfigs)
	if err != nil {
		log.Fatalf("error creating topics: %v", err)
	}

	ProducerConn = kafka.NewWriter(kafka.WriterConfig{
		Brokers:    strings.Split(kafka_addr, ","),
		Topic:      consts.TopicProduct,
		Dialer:     dialer,
		BatchBytes: 1048576, // 1MB(default)
	})
}

func writeMessage(conn *kafka.Writer, message []byte, key string) (err error) {
	err = conn.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(key),
		Value: message,
	})
	if err != nil {
		return fmt.Errorf("error writing %v", message)
	}
	return
}

func ProdMessage(message []byte, key string) error {
	return writeMessage(ProducerConn, message, key)
}
