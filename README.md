# CrudAPI-golang
- By this we are supporting, jwt authentication based which is used for authentication/authorization.
- Two types of information are passed into mongoDB(destination of dataflow).
    1. Users data using rest api call.
    2. Products information using grpc call via cli.

### Initialization
End-to-end flow is written in **./build.sh** file.
1. It initially stops existing containers locally.
2. Go files are created as *"grpcapi"* module using .proto files in *grpcapi* folder
3. We build binary images for each of the microservice we provide in this.
    - mongoserver
    - userserver
    - productserver
4. Using these images, we build docker images using "ubuntu" as base image.
5. We containerize them.

## Mongoserver
This is the intermediate server provided for the services need to do to contact directly to mongo.

In this service we provide, direct contact with mongodb via which we do crud operations in it.
- Get kafka events from the topics from each service and perform insert or update or delete operations.
- Get rest operation call, and return/update the results.

## Userserver
This is running on port 8000 as rest service.
```
POST: localhost:8000/login
	{
	  "user": "username1",
	  "password": "password1"
	}

POST: localhost:8000/create
	{
	  "users": 
	    {
	      "phonenumber": "user",
	      "age": 10,
	      "height": 112.5,
	      "address": {
	        "state": "Telangana",
	        "country": "India"
	      }
	    }
	}

PUT localhost:8000/update/<uid>
	{
	  "username": "user",
	  "phonenumber": "41465645",
	  "age": 22,
	  "address": {
	    "city": "Hyderabad"
	  }
	}

GET localhost:8000/get/<uid>
```
To initialize these services, user must login to it. These user credentials are stored in mongodb of **LoginUser** variable valued collection. In which data is stored as
```
{
  "user": "username1",
  "password": "password1"
}
```
After successful login, this transaction creates a jwt token with in "access_key" field. For each transation, the token is verified and confirms if the token is valid and approves the transaction. Each will be stored in **UserCollection** variable valued collection.

## Productserver
Likewise userserver, this talks with mongoserver via grpc calls instead of rest calls.

The port this runs on is 7007.
```
	go run cliclient/main/main.go --serv_addr localhost:7007 -login -loginuser username1 -loginpass password1
	# (Use the jwtkey generated and put as value in a -access flag)
	go run cliclient/main/main.go -insert -addressline1 addr1 -addressline2 addr2 -city c -state s -country count -zipcode z --serv_addr localhost:7007 -product "{}" -access <token>
	go run cliclient/main/main.go --serv_addr localhost:7007 -update -addressline1 ad1 -product '{"category":"cat", "object": "obj1"}' -uid <uid> -access <token>
	go run cliclient/main/main.go --serv_addr localhost:7007 -delete -uid <uid> -access <token>
```
The created jwt token is put in as a string inside the next trancation for key _-access_.
Every input data will be stored in **ProductCollection** variable valued collection in mongodb.