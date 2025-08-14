#!/bin/sh
set -e

mkdir -p build

# Honor NO_CACHE and optional platform for Apple Silicon cross-builds
BUILD_FLAGS="--progress plain --build-arg RUBY_VERSIONS=\"${KEEP_RUBY_GEM_VERSIONS:-3.2.0,3.3.0,3.4.0}\""
if [ -n "${NO_CACHE:-}" ]; then BUILD_FLAGS="$BUILD_FLAGS --no-cache"; fi
if [ -n "${DOCKER_DEFAULT_PLATFORM:-}" ]; then BUILD_FLAGS="$BUILD_FLAGS --platform ${DOCKER_DEFAULT_PLATFORM}"; fi

eval docker build $BUILD_FLAGS -t aws-otel-lambda-ruby-layer otel
docker run --rm -v "$(pwd)/build:/out" aws-otel-lambda-ruby-layer
