#!/bin/sh

mkdir -p build
docker build -t aws-otel-lambda-python-layer otel
docker run -it --rm -v "$(pwd)/build:/out" aws-otel-lambda-python-layer
