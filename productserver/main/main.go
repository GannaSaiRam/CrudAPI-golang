package main

import (
	"crudapi/productserver"
	"flag"
)

var (
	grpcPort = flag.String("gport", "7007", "Server grpc port")
	// port        = flag.String("port", "7000", "Port of product server")
	kafka       = flag.String("kafka", "kafka:19092", "Kafka Address")
	mongoServer = flag.String("mongo_address", "localhost:9000", "Server address")
)

func main() {
	flag.Parse()
	productserver.InitMongo(*mongoServer)
	go productserver.ProductProducer(*kafka)
	// GRPC
	productserver.GrpcProductServerStart(*grpcPort)
}
