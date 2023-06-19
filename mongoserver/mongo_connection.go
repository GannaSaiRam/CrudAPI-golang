package mongoserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const ()

type Connection struct {
	Client *mongo.Client
	Ctx    context.Context
}

type Addr struct {
	Addressline1 string `json:"addressLine1" bson:"addressLine1"`
	Addressline2 string `json:"addressLine2" bson:"addressLine2"`
	City         string `json:"city" bson:"city"`
	State        string `json:"state" bson:"state"`
	ZipCode      string `json:"zipcode" bson:"zipcode"`
	Country      string `json:"country" bson:"country"`
}

type User struct {
	UID         string  `json:"uid" bson:"uid"`
	Username    string  `json:"username" bson:"username"`
	PhoneNumber string  `json:"phonenumber" bson:"phonenumber"`
	Age         int     `json:"age" bson:"age"`
	Height      float64 `json:"height" bson:"height"`
	Email       string  `json:"email" bson:"email"`
	Address     Addr    `json:"address" bson:"address"`
}

var (
	MongoConnection *Connection
	MongoUri        string
)

func connect(uri string) (*mongo.Client, error) {
	var (
		cancel context.CancelFunc
		ctx    context.Context
	)
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return mongo.Connect(ctx, options.Client().ApplyURI(uri))
}

func clone(mongoconn *Connection) *Connection {
	return &Connection{
		Client: mongoconn.Client,
		Ctx:    mongoconn.Ctx,
	}
}
func NewConnection(mongoUri string) *Connection {
	MongoUri = mongoUri
	var (
		err    error
		client *mongo.Client
	)
	if MongoConnection != nil {
		return clone(MongoConnection)
	}
	client, err = connect(MongoUri)
	if err != nil {
		log.Fatalf("MongoDB connection failed: %v", err)
	}
	MongoConnection = &Connection{Client: client}
	return MongoConnection
}

func Close(client *mongo.Client, ctx context.Context) {
	var (
		err error
	)
	if err = client.Disconnect(ctx); err != nil {
		log.Println(err)
		log.Fatalf("Client connection closed failure: %v", err)
	}
}

func ping(client *mongo.Client, ctx context.Context) (err error) {
	err = client.Ping(ctx, readpref.Primary())
	if err == nil {
		log.Println("Successfully connected")
	}
	return err
}

func BsonObj(message []byte) (bdoc User, err error) {
	fmt.Println(string(message))
	err = json.Unmarshal(message, &bdoc)
	return bdoc, err
}

func MongoConn(mongoUri string) context.CancelFunc {
	var (
		err      error
		mongoCtx context.Context
		cancel   context.CancelFunc
	)
	NewConnection(mongoUri)
	mongoCtx, cancel = context.WithCancel(context.TODO())
	MongoConnection.Ctx = mongoCtx
	if err = ping(MongoConnection.Client, mongoCtx); err != nil {
		log.Fatal("Error in connecting to mongodb")
	}
	return cancel
}
