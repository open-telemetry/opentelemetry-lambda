#!/bin/sh
set -e

mkdir -p build

docker build --progress plain -t aws-otel-lambda-ruby-layer otel
docker run --rm -v "$(pwd)/build:/out" aws-otel-lambda-ruby-layer
