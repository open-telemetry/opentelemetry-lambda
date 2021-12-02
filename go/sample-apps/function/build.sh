#!/bin/sh

GOARCH=${GOARCH-amd64}

mkdir -p build
GOOS=linux go build -o ./build/bootstrap .
cd build || exit
zip bootstrap.zip bootstrap
