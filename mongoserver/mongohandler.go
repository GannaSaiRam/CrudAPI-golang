package mongoserver

import (
	"crudapi/consts"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Credentials struct {
	Username string `bson:"user" json:"user"`
	Password string `bson:"password" json:"password"`
}

func UserLogin(w http.ResponseWriter, r *http.Request) {
	var (
		err       error
		username  string
		col       *mongo.Collection
		credBytes []byte
		filter    primitive.D
	)
	username = mux.Vars(r)["username"]
	filter = bson.D{{Key: "user", Value: username}}
	col = MongoConnection.Client.Database(consts.DATABASE).Collection(consts.LoginUser)
	var cred Credentials
	err = col.FindOne(MongoConnection.Ctx, filter).Decode(&cred)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("No user found with the username"))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error while finding user: %v", err)))
		return
	}
	credBytes, err = json.Marshal(cred)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Error marshalling user credential: %v", err)))
		return
	}
	w.Write(credBytes)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	var (
		err       error
		userInfo  User
		uid       string
		col       *mongo.Collection
		userBytes []byte
		filter    primitive.D
	)
	uid = mux.Vars(r)["uid"]
	filter = bson.D{{Key: "uid", Value: uid}}
	col = MongoConnection.Client.Database(consts.DATABASE).Collection(consts.UserCollection)
	err = col.FindOne(MongoConnection.Ctx, filter).Decode(&userInfo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("No user found with the username"))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error while finding user: %v", err)))
		return
	}
	userBytes, err = json.Marshal(userInfo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Error marshalling user credential: %v", err)))
		return
	}
	w.Write(userBytes)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		userInfo      User
		uid           string
		col           *mongo.Collection
		replace_Error *mongo.SingleResult
		filter        primitive.D
	)

	uid = mux.Vars(r)["uid"]
	filter = bson.D{{Key: "uid", Value: uid}}
	col = MongoConnection.Client.Database(consts.DATABASE).Collection(consts.UserCollection)
	err = col.FindOne(MongoConnection.Ctx, filter).Decode(&userInfo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("No user found with the username"))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error while finding user: %v", err)))
		return
	}
	var bodyUser User
	if err = json.NewDecoder(r.Body).Decode(&bodyUser); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Error marshalling user credential: %v", err)))
		return
	}
	bodyUser.UID = uid
	if bodyUser.Username != "" {
		userInfo.Username = bodyUser.Username
	}
	if bodyUser.Age != 0 {
		userInfo.Age = bodyUser.Age
	}
	if bodyUser.Email != "" {
		userInfo.Email = bodyUser.Email
	}
	if bodyUser.PhoneNumber != "" {
		userInfo.PhoneNumber = bodyUser.PhoneNumber
	}
	if bodyUser.Height != 0 {
		userInfo.Height = bodyUser.Height
	}
	if bodyUser.Address.Addressline1 != "" {
		userInfo.Address.Addressline1 = bodyUser.Address.Addressline1
	}
	if bodyUser.Address.Addressline2 != "" {
		userInfo.Address.Addressline2 = bodyUser.Address.Addressline2
	}
	if bodyUser.Address.City != "" {
		userInfo.Address.City = bodyUser.Address.City
	}
	if bodyUser.Address.State != "" {
		userInfo.Address.State = bodyUser.Address.State
	}
	if bodyUser.Address.ZipCode != "" {
		userInfo.Address.ZipCode = bodyUser.Address.ZipCode
	}
	if bodyUser.Address.Country != "" {
		userInfo.Address.Country = bodyUser.Address.Country
	}
	replace_Error = col.FindOneAndReplace(
		MongoConnection.Ctx, filter, userInfo,
	)
	if replace_Error.Err() != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	w.Write([]byte("Updated user"))
}
