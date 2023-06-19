package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"crudapi/userserver"

	"github.com/gorilla/mux"
)

var (
	// grpcPort = flag.String("gport", "8008", "Server grpc port")
	restPort    = flag.String("port", "8000", "Server rest port")
	mongoServer = flag.String("mongo_address", "localhost:9000", "Server address")
	kafka       = flag.String("kafka", "kafka:19092", "Kafka running address")
)

func main() {
	var (
		err    error
		router *mux.Router
	)
	flag.Parse()

	userserver.Init(*mongoServer)
	go userserver.Producer(*kafka)
	// GRPC
	// userserver.GrpcUserServerStart(*grpcPort)

	// REST
	router = mux.NewRouter()
	router.HandleFunc("/login", userserver.Login).Methods("POST")
	router.HandleFunc("/refresh", userserver.Refresh).Methods("GET")
	router.HandleFunc("/home", userserver.Home).Methods("GET")
	router.HandleFunc("/create", userserver.Create).Methods("POST")
	router.HandleFunc("/get/{uid}", userserver.Get).Methods("GET")
	router.HandleFunc("/update/{uid}", userserver.Update).Methods("PUT")
	log.Printf("Started server at %v\n", *restPort)
	err = http.ListenAndServe(fmt.Sprintf(":%v", *restPort), router)
	if err != nil {
		log.Fatal(err)
	}
}
