package userserver

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var jwtKey = []byte("secret_key")

// var users = map[string]string{
// 	"user1": "password1",
// 	"user2": "password2",
// 	"user3": "password3",
// }

var MongoServerAddress string

func Init(mongoServerAddress string) {
	MongoServerAddress = mongoServerAddress
}

type Credentials struct {
	Username string `json:"user"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type Addr struct {
	Addressline1 string `json:"addressLine1"`
	Addressline2 string `json:"addressLine2"`
	City         string `json:"city"`
	State        string `json:"state"`
	ZipCode      string `json:"zipcode"`
	Country      string `json:"country"`
}

type User struct {
	UID         string  `json:"uid"`
	Username    string  `json:"username"`
	PhoneNumber string  `json:"phonenumber"`
	Age         int     `json:"age"`
	Height      float64 `json:"height"`
	Email       string  `json:"email"`
	Address     Addr    `json:"address"`
}

type CreateUser struct {
	User User `json:"users"`
}

type CreateUsers struct {
	Users []User `json:"users"`
}

func checkValid(w http.ResponseWriter, r *http.Request) (*Claims, error) {
	var (
		err         error
		cookie      *http.Cookie
		token       *jwt.Token
		tokenString string
		claims      = &Claims{}
	)
	cookie, err = r.Cookie("token")
	if err == http.ErrNoCookie {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("No cookie present"))
		return nil, err
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Request is of not expected format: %v", err)))
		return nil, err
	}
	tokenString = cookie.Value
	token, err = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err == jwt.ErrSignatureInvalid {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Present cookie is not valid"))
		return nil, err
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Error while parsing cookie: %v", err)))
		return nil, err
	}
	if !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Token is not valid, please re-login"))
		return nil, fmt.Errorf("InvalidToken")
	}
	return claims, nil
}

func getCredsFromMongo(w http.ResponseWriter, credentials Credentials) (password string, err error) {
	var (
		requestURL string
		req        *http.Request
		res        *http.Response
		resBody    []byte
	)

	requestURL = fmt.Sprintf("http://%v/login/%v", MongoServerAddress, credentials.Username)
	log.Println(requestURL)
	req, err = http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		if w != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("client: could not create request: %v\n", err)))
		}
		return password, fmt.Errorf("client: could not create request: %v", err)
	}
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		if w != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("client: error making http request: %s\n", err)))
		}
		return password, fmt.Errorf("client: error making http request: %s", err)
	}
	resBody, err = io.ReadAll(res.Body)
	if err != nil {
		if w != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("read: %v", err)))
		}
		return password, fmt.Errorf("read: %v", err)
	}
	var cred Credentials
	err = json.Unmarshal(resBody, &cred)
	if err != nil {
		if w != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("No user with username"))
		}
		return password, fmt.Errorf("no user with username: %v", credentials.Username)
	}
	if cred.Password == "" {
		if w != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("User doesn't exist"))
		}
		return password, fmt.Errorf("User doesn't exist")
	}
	return cred.Password, nil
}

func getUserInfoFromMongo(w http.ResponseWriter, uid string) (*User, error) {
	var (
		err          error
		requestURL   string
		req          *http.Request
		res          *http.Response
		resBody      []byte
		user         User
		writeMessage string
	)
	requestURL = fmt.Sprintf("http://%v/get/%v", MongoServerAddress, uid)
	req, err = http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		writeMessage = "client: could not create request: %v"
		if w != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(writeMessage, err)))
		}
		return nil, fmt.Errorf(writeMessage, err)
	}
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		writeMessage = "client: error making http request: %s"
		if w != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(writeMessage, err)))
		}
		return nil, fmt.Errorf(writeMessage, err)
	}
	resBody, err = io.ReadAll(res.Body)
	if err != nil {
		writeMessage = "read: %v"
		if w != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(writeMessage, err)))
		}
		return nil, fmt.Errorf(writeMessage, err)
	}
	err = json.Unmarshal(resBody, &user)
	if err != nil {
		if w != nil {
			w.WriteHeader(http.StatusUnauthorized)
		}
		return nil, err
	}
	return &user, nil
}

func Login(w http.ResponseWriter, r *http.Request) {
	var (
		credentials       Credentials
		err               error
		expirationTime    time.Time
		claims            *Claims
		token             *jwt.Token
		passwordFromMongo string
	)
	if err = json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Request body is not in expected format"))
		return
	}
	if credentials.Password == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Password Empty"))
		return
	}
	passwordFromMongo, err = getCredsFromMongo(w, credentials)
	if err != nil {
		return
	}
	if passwordFromMongo != credentials.Password {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Password mismatch"))
		return
	}
	expirationTime = time.Now().Add(5 * time.Minute)
	claims = &Claims{
		Username: credentials.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	var tokenString string
	if tokenString, err = token.SignedString(jwtKey); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error signing the secret jwt string"))
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
}

func Home(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		claims *Claims
	)
	claims, err = checkValid(w, r)
	if err != nil {
		return
	}
	w.Write([]byte(fmt.Sprintf("Hello, %s", claims.Username)))
}

func Refresh(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		token          *jwt.Token
		tokenString    string
		expirationTime time.Time
		claims         *Claims
	)
	claims, err = checkValid(w, r)
	if err != nil {
		return
	}
	expirationTime = time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = expirationTime.Unix()
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error signing the secret jwt string to refresh"))
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
}

func Create(w http.ResponseWriter, r *http.Request) {
	var (
		err          error
		user         CreateUser
		messageBytes []byte
	)
	_, err = checkValid(w, r)
	if err != nil {
		return
	}
	err = json.NewDecoder(r.Body).Decode(&user)
	if err == nil {
		user.User.UID = uuid.New().String()
		fmt.Printf("Created: %+v\n", user.User)
		messageBytes, err = json.Marshal(user.User)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Kafka message body is not as expected"))
			return
		}
		// messageBytes = append([]byte("["), append(messageBytes, ']')...)
		if err = WriteMessage(messageBytes, user.User.UID); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error in writing request body to kafka"))
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf("User created: %v", user.User.UID)))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}

func Get(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		user       *User
		writeBytes []byte
		uid        string
	)
	_, err = checkValid(w, r)
	if err != nil {
		return
	}
	uid = mux.Vars(r)["uid"]
	if uid == "" {
		w.Write([]byte("User Id is null string"))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	user, err = getUserInfoFromMongo(w, uid)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("User doesn't exist for the given userid"))
		return
	}
	writeBytes, err = json.Marshal(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Body is not as expected"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(writeBytes)
}

func Update(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		uid        string
		res        *http.Response
		req        *http.Request
		requestURL string
	)

	_, err = checkValid(w, r)
	if err != nil {
		return
	}
	uid = mux.Vars(r)["uid"]
	if uid == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("User Id is null string"))
		return
	}

	requestURL = fmt.Sprintf("http://%v/update/%v", MongoServerAddress, uid)
	req, err = http.NewRequest(http.MethodPut, requestURL, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("client: could not create request: %v\n", err)))
		return
	}
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("client: error making http request: %s\n", err)))
		return
	}
	log.Printf("Error while updating request: %v", err)
	log.Printf("%+v", res)
}
