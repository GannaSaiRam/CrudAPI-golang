package main

import (
	"context"
	"crudapi/mongoserver"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	port     = flag.String("port", "9000", "Mongoserver running port")
	kafka    = flag.String("kafka", "kafka:19092", "Kafka running address")
	mongoUri = flag.String("mongo_uri", "mongodb://service_mongodb", "Running mongodb address")
)

func main() {
	var (
		err    error
		router *mux.Router
	)
	flag.Parse()
	cancel := mongoserver.MongoConn(*mongoUri)
	defer cancel()
	defer mongoserver.Close(mongoserver.MongoConnection.Client, context.TODO())
	go mongoserver.UserConsume(context.TODO(), *kafka, *mongoUri)
	go mongoserver.ProdConsume(*kafka, *mongoUri)
	router = mux.NewRouter()

	router.HandleFunc("/login/{username}", mongoserver.UserLogin).Methods("GET")
	router.HandleFunc("/get/{uid}", mongoserver.GetUser).Methods("GET")
	router.HandleFunc("/update/{uid}", mongoserver.UpdateUser).Methods("PUT")
	log.Printf("Started server at %v\n", *port)
	err = http.ListenAndServe(fmt.Sprintf(":%v", *port), router)
	if err != nil {
		log.Fatal(err)
	}
}
