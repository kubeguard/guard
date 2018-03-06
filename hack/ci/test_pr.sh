#!/bin/bash

set -x -e

CURRENT_DIR=$(pwd)
mkdir -p $GOPATH/src/github.com/appscode
cp -r pull-request $GOPATH/src/github.com/appscode/guard
cd $GOPATH/src/github.com/appscode/guard
git rev-parse HEAD
go build -v
go test -v ./... &> $CURRENT_DIR/test_result/message
