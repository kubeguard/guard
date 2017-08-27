#!/usr/bin/env bash

pushd $GOPATH/src/github.com/appscode/kad/hack/gendocs
go run main.go
popd
