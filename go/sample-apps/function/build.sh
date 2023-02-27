#!/bin/sh
set -e

GOARCH=${GOARCH-amd64}

mkdir -p build
CGO_ENABLED=0 GOOS=linux go build -o ./build/bootstrap .
cd build
zip bootstrap.zip bootstrap
