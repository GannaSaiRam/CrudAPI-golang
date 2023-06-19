package main

import (
	"context"
	"crudapi/grpcapi"
	"encoding/json"
	"flag"
	"log"
	"path"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var (
	data        string
	accessToken string

	login     bool
	loginuser string
	loginpass string

	update        bool
	insert        bool
	delete        bool
	serverAddress string
	uid           string
	// username      string
	// phoneNumber   string
	// age           int
	// height        float64
	// email         string
	addressline1 string
	addressline2 string
	city         string
	state        string
	zipCode      string
	country      string
)

func init() {
	flag.StringVar(&serverAddress, "serv_addr", "localhost:7007", "To address which this cli is passing data to")
	flag.BoolVar(&login, "login", false, "Login with username and password")
	flag.BoolVar(&insert, "insert", false, "Insert new user")
	flag.BoolVar(&update, "update", false, "Update user")
	flag.BoolVar(&delete, "delete", false, "Get the user info")

	flag.StringVar(&accessToken, "access", "", "Access token to run the commands")
	flag.StringVar(&loginuser, "loginuser", "", "Username to login")
	flag.StringVar(&loginpass, "loginpass", "", "Password to login")

	flag.StringVar(&uid, "uid", "", "userid of update/get")
	// flag.StringVar(&username, "username", "", "Username of new user")
	// flag.StringVar(&phoneNumber, "phonenumber", "", "PhoneNumber of new user")
	// flag.IntVar(&age, "age", 0, "Age of new user")
	// flag.Float64Var(&height, "height", 0.0, "Height of new user")
	// flag.StringVar(&email, "email", "", "Email of new user")
	flag.StringVar(&data, "product", "", "Information of product in json format")
	flag.StringVar(&addressline1, "addressline1", "", "Addressline1 of new user")
	flag.StringVar(&addressline2, "addressline2", "", "Addressline2 of new user")
	flag.StringVar(&city, "city", "", "City of new user")
	flag.StringVar(&state, "state", "", "State of new user")
	flag.StringVar(&zipCode, "zipcode", "", "ZipCode of new user")
	flag.StringVar(&country, "country", "", "Country of new user")
}

func middlewareClientInterceptor(ctx context.Context, method string,
	req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

	if path.Base(method) != "Login" && accessToken != "" {
		// adding jwt to interceptor
		ctx = metadata.AppendToOutgoingContext(ctx, "access_key", accessToken)
	}

	err := invoker(ctx, method, req, reply, cc, opts...)
	return err
}

func main() {
	// Client
	flag.Parse()
	var (
		err     error
		conn    grpc.ClientConnInterface
		client  grpcapi.ProductServiceClient
		newUser *grpcapi.ProductMessage
	)
	conn, err = grpc.Dial(
		serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(middlewareClientInterceptor),
	)
	if err != nil {
		log.Fatalf("Can't dial to server: %v", serverAddress)
	}
	// defer conn.Close()
	client = grpcapi.NewProductServiceClient(conn)
	if !(delete && !update && !insert && !login ||
		!delete && update && !insert && !login ||
		!delete && !update && insert && !login ||
		!delete && !update && !insert && login) {
		log.Println("Please provide any one operation is supported of insert/update/delete/login")
		return
	}
	if login {
		ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
		defer cancel()
		out, err := client.Login(ctx, &grpcapi.LoginReq{Username: loginuser, Password: loginpass})
		if err != nil {
			log.Println(err)
			return
		}
		accessToken = out.AccessToken
		log.Printf("Login Successful, got access token: %v . Use this for next runs.", out.AccessToken)
		return
	}
	newUser = &grpcapi.ProductMessage{
		// Username:    username,
		// PhoneNumber: phoneNumber,
		// Age:         int64(age),
		// Height:      float32(height),
		// Email:       email,
		ProductInfo: []byte(data),
		Source: &grpcapi.Address{
			Addressline1: addressline1,
			Addressline2: addressline2,
			City:         city,
			State:        state,
			ZipCode:      zipCode,
			Country:      country,
		},
	}
	if insert {
		if data == "" || !json.Valid([]byte(data)) {
			log.Printf("Given data is not valid json")
			return
		}
		newUser.Uid = uuid.New().String()
		newUser.Action = grpcapi.Action_INSERT
		sendProductInfo(client, newUser)
		return
	} else if update {
		if data == "" || !json.Valid([]byte(data)) {
			log.Printf("Given data is not valid json")
			return
		}
		if uid == "" {
			log.Println("uid is not present to update")
			return
		}
		newUser.Uid = uid
		newUser.Action = grpcapi.Action_UPDATE
		updateProductInfo(client, newUser)
		return
	} else if delete {
		if uid == "" {
			log.Println("uid is not present to delete info")
			return
		}
		deleteProductInfo(client, &grpcapi.UidParam{Uid: uid, Action: grpcapi.Action_DELETE})
		return
	}

}

func sendProductInfo(client grpcapi.ProductServiceClient, prod *grpcapi.ProductMessage) {
	var (
		err    error
		ctx    context.Context
		cancel context.CancelFunc
		out    *grpcapi.UidParam
	)

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	log.Printf("%+v", prod)
	out, err = client.InsertProduct(ctx, prod)
	if err != nil {
		log.Printf("Error sending new prod info: %v", err)
		return
	}
	log.Printf("Inserted new prod with uid: %v", out.Uid)
}

func updateProductInfo(client grpcapi.ProductServiceClient, prod *grpcapi.ProductMessage) {
	var (
		err    error
		ctx    context.Context
		cancel context.CancelFunc
		out    *grpcapi.UidParam
	)

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	log.Printf("%+v", prod)
	out, err = client.UpdateProduct(ctx, prod)
	if err != nil {
		log.Printf("Error upating prod(%v) info: %v", prod.Uid, err)
		return
	}
	log.Printf("Updated prod with uid: %v", out.Uid)
}

func deleteProductInfo(client grpcapi.ProductServiceClient, prod *grpcapi.UidParam) {
	var (
		err    error
		ctx    context.Context
		cancel context.CancelFunc
		out    *grpcapi.UidParam
	)

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	log.Printf("%+v", prod)
	out, err = client.DeleteProduct(ctx, prod)
	if err != nil {
		log.Printf("Error deleting prod(%v) info: %v", prod.Uid, err)
		return
	}
	log.Printf("ProdInfo with uid(%v) deleted", out.Uid)
}
