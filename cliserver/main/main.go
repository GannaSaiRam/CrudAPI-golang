package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"crudapi/grpcapi"

	"google.golang.org/grpc"
)

var (
	// 	// data          string
	// 	serverAddress string
	port string

// Username     string
// PhoneNumber  string
// Age          int
// Height       float64
// Email        string
// Addressline1 string
// Addressline2 string
// City         string
// State        string
// ZipCode      string
// Country      string
)

type helloServer struct {
	grpcapi.UserServiceServer
}

func init() {
	// flag.StringVar(&data, "data", "", "Used as string")
	// flag.StringVar(&serverAddress, "serv_addr", "localhost:8000", "To address which this cli is passing data to")
	flag.StringVar(&port, "cli_port", "8888", "Port on which cli is going to host")
}

func (h *helloServer) PrintMessage(ctx context.Context, message *grpcapi.UserMessage) (*grpcapi.NoParam, error) {
	log.Printf("%+v", message)
	return &grpcapi.NoParam{}, nil
}

func main() {
	flag.Parse()

	var (
		err        error
		listen     net.Listener
		grpcServer *grpc.Server
	)
	listen, err = net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatalf("Failed to start server at the port: %v", port)
	} else {
		log.Printf("Started listening to port: %v", port)
	}
	grpcServer = grpc.NewServer()
	grpcapi.RegisterUserServiceServer(grpcServer, &helloServer{})
	if grpcServer.Serve(listen) == nil {
		log.Fatalf("Failed to start grpc server: %v", err)
	}
}
