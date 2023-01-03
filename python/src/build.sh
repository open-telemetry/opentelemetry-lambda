#!/bin/sh
set -e

mkdir -p build
docker build -t aws-otel-lambda-python-layer otel
docker run --rm -v "$(pwd)/build:/out" aws-otel-lambda-python-layer
