#!/bin/sh

mkdir -p build
go mod tidy
GOOS=linux GOARCH=amd64 go build -o ./build/bootstrap .
cd build || exit
zip bootstrap.zip bootstrap
