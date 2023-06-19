# TODO
# Go params
PWD = $(shell pwd)
GOCMD=go
GOBUILD=$(GOCMD) build
BINARY_SERVER_LINUX=server
BINARY_MONGOSER_LINUX=mongoserver
BINARY_SERVER_MAC=server
BINARY_MONGOSER_MAC=mongoserver
BINARY_SERVER_WINDOWS=server.exe
BINARY_MONGOSER_WINDOWS=mongoserver.exe

all: build-linux64

# Cross compilation
build-linux64:
	. $(PWD)/.env
	. $(PWD)/gopath.sh
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(PWD)/bin/bin/$(BINARY_SERVER_LINUX) -v $(PWD)/server/main/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(PWD)/bin/bin/$(BINARY_MONGOSER_LINUX) -v $(PWD)/mongoserver/main/main.go
	@echo 'Done building build-linux64!'

build-linux32:
	. $(PWD)/.env
	. $(PWD)/gopath.sh
	CGO_ENABLED=0 GOOS=linux GOARCH=386 $(GOBUILD) -o $(PWD)/bin/bin/$(BINARY_SERVER_LINUX) -v $(PWD)/server/main/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=386 $(GOBUILD) -o $(PWD)/bin/bin/$(BINARY_MONGOSER_LINUX) -v $(PWD)/mongoserver/main/main.go
	@echo 'Done building build-linux32!'

build-mac64:
	. $(PWD)/.env
	. $(PWD)/gopath.sh
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(PWD)/bin/bin/$(BINARY_SERVER_MAC) -v $(PWD)/server/main/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(PWD)/bin/bin/$(BINARY_MONGOSER_MAC) -v $(PWD)/mongoserver/main/main.go
	@echo 'Done building build-mac64!'

build-mac32:
	. $(PWD)/.env
	. $(PWD)/gopath.sh
	CGO_ENABLED=0 GOOS=darwin GOARCH=386 $(GOBUILD) -o $(PWD)/bin/bin/$(BINARY_SERVER_MAC) -v $(PWD)/server/main/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=386 $(GOBUILD) -o $(PWD)/bin/bin/$(BINARY_MONGOSER_MAC) -v $(PWD)/mongoserver/main/main.go
	@echo 'Done building build-mac32!'

build-windows64:
	. $(PWD)/.env
	. $(PWD)/gopath.sh
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(PWD)/bin/bin/$(BINARY_SERVER_WINDOWS) -v $(PWD)/server/main/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(PWD)/bin/bin/$(BINARY_MONGOSER_WINDOWS) -v $(PWD)/mongoserver/main/main.go
	@echo 'Done building build-windows64!'

build-windows32:
	. $(PWD)/.env
	. $(PWD)/gopath.sh
	CGO_ENABLED=0 GOOS=windows GOARCH=386 $(GOBUILD) -o $(PWD)/bin/bin/$(BINARY_SERVER_WINDOWS) -v $(PWD)/server/main/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=386 $(GOBUILD) -o $(PWD)/bin/bin/$(BINARY_MONGOSER_WINDOWS) -v $(PWD)/mongoserver/main/main.go
	@echo 'Done building build-windows32!'
