package userserver

import (
	"bytes"
	"context"
	"crudapi/grpcapi"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"path"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type helloServer struct {
	grpcapi.UserServiceServer
}

func generate(credentials Credentials) (string, error) {
	var (
		expirationTime time.Time
		claims         *Claims
		token          *jwt.Token
	)
	expirationTime = time.Now().Add(time.Minute * 10)
	claims = &Claims{
		Username: credentials.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func verify(tokenString string) (claims *Claims, err error) {
	var (
		token *jwt.Token
		ok    bool
	)
	token, err = jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(t *jwt.Token) (interface{}, error) {
			var ok bool
			if _, ok = t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("method doesn't match")
			}
			return jwtKey, nil
		},
	)
	if err != nil {
		return claims, fmt.Errorf("invalid token: %v", err)
	}
	claims, ok = token.Claims.(*Claims)
	if !ok {
		return claims, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

func (h *helloServer) Login(ctx context.Context, loginReq *grpcapi.LoginReq) (out *grpcapi.LoginRes, err error) {
	var (
		credentials       = Credentials{loginReq.GetUsername(), loginReq.GetPassword()}
		passwordFromMongo string
		token             string
	)
	if loginReq.Username == "" {
		return out, status.Errorf(codes.Unauthenticated, "Username is empty")
	}
	if loginReq.Password == "" {
		return out, status.Errorf(codes.Unauthenticated, "Password is empty")
	}
	passwordFromMongo, err = getCredsFromMongo(nil, credentials)
	if err != nil {
		return
	}
	if passwordFromMongo != credentials.Password {
		return out, status.Errorf(codes.Unauthenticated, "Password mismatch")
	}
	token, err = generate(credentials)
	if err != nil {
		return out, status.Errorf(codes.Internal, "Can't generate access token")
	}
	return &grpcapi.LoginRes{AccessToken: token}, err
}

func (h *helloServer) InsertUser(ctx context.Context, message *grpcapi.UserMessage) (*grpcapi.UidParam, error) {
	var (
		err          error
		messageBytes []byte
		userMessage  User
	)
	userMessage = User{
		UID:         message.Uid,
		Username:    message.Username,
		PhoneNumber: message.PhoneNumber,
		Age:         int(message.Age),
		Height:      float64(message.Height),
		Email:       message.Email,
		Address: Addr{
			Addressline1: message.Address.Addressline1,
			Addressline2: message.Address.Addressline2,
			City:         message.Address.City,
			State:        message.Address.State,
			Country:      message.Address.Country,
			ZipCode:      message.Address.ZipCode,
		},
	}
	messageBytes, err = json.Marshal(userMessage)
	if err != nil {
		return &grpcapi.UidParam{}, err
	}
	err = WriteMessage(messageBytes, userMessage.UID)
	if err != nil {
		return &grpcapi.UidParam{}, err
	}
	return &grpcapi.UidParam{Uid: userMessage.UID}, err
}

func (h *helloServer) UpdateUser(ctx context.Context, message *grpcapi.UserMessage) (*grpcapi.NoParam, error) {

	var (
		err        error
		uid        string
		buf        bytes.Buffer
		requestURL string
		req        *http.Request
		res        *http.Response
	)

	// _, err = checkValid(w, r)
	// if err != nil {
	// 	return
	// }
	uid = message.Uid
	err = json.NewEncoder(&buf).Encode(message)
	if err != nil {
		log.Fatal(err)
	}

	requestURL = fmt.Sprintf("http://%v/update/%v", MongoServerAddress, uid)
	req, err = http.NewRequest(http.MethodPut, requestURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("creating request error: %v", err)
	}
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client: error making http request: %s", err)
	}
	log.Printf("%+v", res)

	return &grpcapi.NoParam{}, nil
}

func (h *helloServer) GetUser(ctx context.Context, message *grpcapi.UidParam) (*grpcapi.UserMessage, error) {
	var (
		err  error
		user *User
		uid  string
	)
	// _, err = checkValid(w, r)
	// if err != nil {
	// 	return
	// }
	uid = message.Uid
	user, err = getUserInfoFromMongo(nil, uid)
	if err != nil {
		return &grpcapi.UserMessage{}, err
	}
	return &grpcapi.UserMessage{
		Uid:         user.UID,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
		Age:         int64(user.Age),
		Height:      float32(user.Height),
		Email:       user.Email,
		Address: &grpcapi.Address{
			Addressline1: user.Address.Addressline1,
			Addressline2: user.Address.Addressline2,
			City:         user.Address.City,
			State:        user.Address.State,
			Country:      user.Address.Country,
			ZipCode:      user.Address.ZipCode,
		},
	}, nil
}

func midlewareFunction(
	ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	var (
		md     metadata.MD
		ok     bool
		errMsg error
		values []string
		bearer string
		claims *Claims
	)

	log.Println("Triggerred: ", info.FullMethod)
	if path.Base(info.FullMethod) != "Login" {
		md, ok = metadata.FromIncomingContext(ctx)
		if !ok {
			errMsg = fmt.Errorf("metadata is not provided")
			log.Printf("Error: %v", errMsg)
			return nil, errMsg
		}
		values = md["access_key"]
		if len(values) == 0 {
			errMsg = fmt.Errorf("authorization token is not provided")
			log.Printf("Error: %v", errMsg)
			return nil, errMsg
		}
		bearer = values[0]
		claims, err = verify(bearer)
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}
		if claims.ExpiresAt < time.Now().Unix() {
			log.Printf("Token Expired, Need to relogin")
			return nil, fmt.Errorf("token expired, please relogin... ")
		}
	}

	return handler(ctx, req)
}

func GrpcUserServerStart(grpcPort string) {
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
	grpcapi.RegisterUserServiceServer(grpcServer, &helloServer{})
	if grpcServer.Serve(listen) == nil {
		log.Fatalf("Failed to start grpc server: %v", err)
	}
}
