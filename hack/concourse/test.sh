#!/bin/sh

set -x -e

CURRENT_DIR=$(pwd)
mkdir -p $GOPATH/src/github.com/appscode
cp -r guard $GOPATH/src/github.com/appscode
cd $GOPATH/src/github.com/appscode/guard
go build ./...
go test -v ./...
