#!/bin/bash

/$BIN_MONGOSERVER \
    -port=$MONGOSERVER_PORT \
    -kafka=$KAFKA \
    -mongo_uri=$MONGODB_URI
