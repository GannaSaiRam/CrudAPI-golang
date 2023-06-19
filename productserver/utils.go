package productserver

import (
	"fmt"
	"io"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var MongoServerAddress string

func InitMongo(mongoUrl string) {
	MongoServerAddress = mongoUrl
}

func Generate(credentials Credentials) (string, error) {
	var (
		expirationTime time.Time
		claims         *Claims
		token          *jwt.Token
	)
	expirationTime = time.Now().Add(time.Minute * 90)
	claims = &Claims{
		Username: credentials.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func Verify(tokenString string) (claims *Claims, err error) {
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

func GetCredsFromMongo(username string) ([]byte, error) {
	var (
		err        error
		requestURL string
		req        *http.Request
		res        *http.Response
		resBody    []byte
	)

	requestURL = fmt.Sprintf("http://%v/login/%v", MongoServerAddress, username)
	req, err = http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("client: could not create request: %v", err)
	}
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return []byte{}, fmt.Errorf("client: error making http request: %s", err)
	}
	resBody, err = io.ReadAll(res.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("read: %v", err)
	}
	return resBody, nil
}
