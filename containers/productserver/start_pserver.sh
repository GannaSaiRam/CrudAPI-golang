#!/bin/bash

/$BIN_PRODUCTSERVER \
    -mongo_address=$MONGO_ADDRESS \
    -gport=$PRODUCTSERVER_GRPC \
    -kafka=$KAFKA \
    # -port=$PRODUCTSERVER_PORT \

