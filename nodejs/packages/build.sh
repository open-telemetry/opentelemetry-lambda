#!/bin/sh
set -e

mkdir -p build
docker build --progress plain -t aws-otel-lambda-nodejs-layer ../
docker run --rm -v "$(pwd)/build:/out" aws-otel-lambda-nodejs-layer
