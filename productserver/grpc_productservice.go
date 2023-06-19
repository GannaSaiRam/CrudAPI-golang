package productserver

import (
	"crudapi/grpcapi"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

type Credentials struct {
	Username string `json:"user"`
	Password string `json:"password"`
}

func GrpcProductServerStart(grpcPort string) {
	var (
		err        error
		listen     net.Listener
		grpcServer *grpc.Server
	)
	listen, err = net.Listen("tcp", fmt.Sprintf(":%v", grpcPort))
	if err != nil {
		log.Fatalf("Failed to start server at the grpc port: %v", grpcPort)
	} else {
		log.Printf("Started listening to grpc port: %v", grpcPort)
	}
	grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(midlewareFunction),
	)
	grpcapi.RegisterProductServiceServer(grpcServer, &ProductService{})
	if grpcServer.Serve(listen) == nil {
		log.Fatalf("Failed to start grpc server: %v", err)
	}
}
