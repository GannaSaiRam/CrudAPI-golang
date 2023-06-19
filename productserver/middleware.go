package productserver

import (
	"context"
	"fmt"
	"log"
	"path"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var jwtKey = []byte("Secret Key")

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
