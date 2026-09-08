#!/bin/sh
set -e

ARCH=${ARCH:-amd64}
PLATFORM="linux/${ARCH}"

mkdir -p build

docker build --progress plain --platform "$PLATFORM" \
  -t "aws-otel-lambda-ruby-layer-${ARCH}" otel

docker run --rm --platform "$PLATFORM" \
  -v "$(pwd)/build:/out" \
  "aws-otel-lambda-ruby-layer-${ARCH}"

mv build/opentelemetry-ruby-layer.zip "build/opentelemetry-ruby-layer-${ARCH}.zip"
