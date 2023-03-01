#!/bin/sh
set -e

## Valid platform values are "amd64" and "arm64"
platform=${1:-amd64}

mkdir -p build
docker buildx build --platform linux/${platform} --build-arg PLATFORM=${platform} --load -t aws-otel-lambda-python-layer  otel
docker create --name aws-otel-lambda-python-container-${platform} --platform linux/${platform} aws-otel-lambda-python-layer
docker cp aws-otel-lambda-python-container-${platform}:/build/layer.zip ./build
