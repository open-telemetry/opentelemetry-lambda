#!/bin/sh
set -e

GOARCH=${GOARCH-amd64}

mkdir -p build
GOOS=linux go build -o ./build/bootstrap .
cd build
zip bootstrap.zip bootstrap
