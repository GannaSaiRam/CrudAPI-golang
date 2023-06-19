#!/bin/bash

export GOPATH=$(pwd)/bin
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOPATH:$GOBIN
export PATH="$PATH:$(go env GOPATH)/bin"