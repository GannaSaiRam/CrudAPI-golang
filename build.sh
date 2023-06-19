# Docker down before creating new
docker compose down

export ROOT_FOLDER="$(pwd)"
echo $ROOT_FOLDER


source ${ROOT_FOLDER}/.env
source ${ROOT_FOLDER}/gopath.sh


pushd $ROOT_FOLDER/grpcapi
protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative *.proto
popd
# Build MongoServer
go build -o ${ROOT_FOLDER}/${IMAGE_LOCATION}/${BIN_MONGOSERVER} ${ROOT_FOLDER}/mongoserver/main/main.go
# Build UserServer
go build -o ${ROOT_FOLDER}/${IMAGE_LOCATION}/${BIN_USERSERVER} ${ROOT_FOLDER}/userserver/main/main.go
# Build ProductServer
go build -o ${ROOT_FOLDER}/${IMAGE_LOCATION}/${BIN_PRODUCTSERVER} ${ROOT_FOLDER}/productserver/main/main.go


# create docker images using bin files
pushd ./containers/mongoserver
cp ${ROOT_FOLDER}/${IMAGE_LOCATION}/${BIN_MONGOSERVER} .
docker build -t ${IMAGENAME_MONGOSERVER} --build-arg port=${MONGOSERVER_PORT} --build-arg image=${BIN_MONGOSERVER} .
rm -f ${BIN_MONGOSERVER}
popd

pushd ./containers/userserver
cp ${ROOT_FOLDER}/${IMAGE_LOCATION}/${BIN_USERSERVER} .
docker build -t ${IMAGENAME_USERSERVER} --build-arg port=${USERSERVER_PORT} --build-arg image=${BIN_USERSERVER} .
rm -f ${BIN_USERSERVER}
popd

pushd ./containers/productserver
cp ${ROOT_FOLDER}/${IMAGE_LOCATION}/${BIN_PRODUCTSERVER} .
docker build -t ${IMAGENAME_PRODUCTSERVER} --build-arg port=${PRODUCTSERVER_PORT} --build-arg image=${BIN_PRODUCTSERVER} .
rm -f ${BIN_PRODUCTSERVER}
popd

# Containerize
docker compose up -d