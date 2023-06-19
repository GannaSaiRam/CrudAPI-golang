#!/bin/bash

/$BIN_USERSERVER \
    -mongo_address=$MONGO_ADDRESS \
    -port=$USERSERVER_PORT \
    -kafka=$KAFKA \
    # -gport=$USERSERVER_GRPC \
