#!/bin/sh

set -x -e
echo "test"
CURRENT_DIR=$(pwd)
mkdir -p $GOPATH/src/github.com/appscode
cp -r pull-request $GOPATH/src/github.com/appscode/guard
cd $GOPATH/src/github.com/appscode/guard
echo "testing commit"
git rev-parse HEAD
go build ./...
go test -v ./...
