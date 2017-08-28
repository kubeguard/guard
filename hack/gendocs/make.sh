#!/usr/bin/env bash

pushd $GOPATH/src/github.com/appscode/guard/hack/gendocs
go run main.go
popd
