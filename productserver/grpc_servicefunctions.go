package productserver

import (
	"context"
	"crudapi/grpcapi"
	"encoding/json"
	"fmt"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type ProductService struct {
	grpcapi.ProductServiceServer
}

type message interface {
	*grpcapi.ProductMessage | *grpcapi.UidParam
	GetUid() string
	ProtoReflect() protoreflect.Message
}

func commonProdCrud[T message](crudProductMessage T) error {
	var (
		err          error
		messageBytes []byte
	)
	messageBytes, err = protojson.Marshal(crudProductMessage)
	if err != nil {
		return fmt.Errorf("error in marshalling document")
	}
	if err = ProdMessage(messageBytes, crudProductMessage.GetUid()); err != nil {
		return fmt.Errorf("error loading message to kafka: %v", err)
	}
	return nil
}

func (p *ProductService) Login(ctx context.Context, loginreq *grpcapi.LoginReq) (out *grpcapi.LoginRes, err error) {
	var (
		credentials = Credentials{loginreq.GetUsername(), loginreq.GetPassword()}
		credsMongo  Credentials
		token       string
		resBody     []byte
	)
	if credentials.Username == "" {
		return out, status.Errorf(codes.Unauthenticated, "Username is empty")
	}
	if credentials.Password == "" {
		return out, status.Errorf(codes.Unauthenticated, "Password is empty")
	}
	resBody, err = GetCredsFromMongo(credentials.Username)
	if err != nil {
		return out, status.Errorf(codes.Internal, "Error retrieving mongo credentials")
	}
	err = json.Unmarshal(resBody, &credsMongo)
	log.Println(string(resBody))
	if err != nil {
		return out, status.Errorf(codes.InvalidArgument, "Arguments are not in expected format")
	}
	if credsMongo.Username == "" || credsMongo.Password == "" {
		return out, status.Errorf(codes.Unauthenticated, "User doesn't exist")
	}
	if credsMongo.Password != credentials.Password {
		return out, status.Errorf(codes.Unauthenticated, "Password mismatch")
	}
	token, err = Generate(credentials)
	if err != nil {
		return out, status.Errorf(codes.Internal, "Can't generate access token")
	}
	return &grpcapi.LoginRes{AccessToken: token}, err
}

func (p *ProductService) InsertProduct(ctx context.Context, insertProduct *grpcapi.ProductMessage) (*grpcapi.UidParam, error) {
	var err error
	if err = commonProdCrud(insertProduct); err != nil {
		return nil, err
	}
	return &grpcapi.UidParam{Uid: insertProduct.Uid}, nil
}

func (p *ProductService) UpdateProduct(ctx context.Context, updateProduct *grpcapi.ProductMessage) (*grpcapi.UidParam, error) {
	var err error
	if err = commonProdCrud(updateProduct); err != nil {
		return nil, err
	}
	return &grpcapi.UidParam{Uid: updateProduct.Uid}, nil
}

func (p *ProductService) DeleteProduct(ctx context.Context, deleteProduct *grpcapi.UidParam) (*grpcapi.UidParam, error) {
	var err error
	if err = commonProdCrud(deleteProduct); err != nil {
		return nil, err
	}
	return &grpcapi.UidParam{Uid: deleteProduct.Uid}, nil
}
