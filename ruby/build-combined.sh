#!/bin/bash

# Build combined Ruby extension layer
# This script builds a combined layer that includes:
# 1. The Ruby instrumentation layer built from local sources in this repo
# 2. The custom Go OpenTelemetry Collector

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$SCRIPT_DIR/build"
COLLECTOR_DIR="$SCRIPT_DIR/../collector"
ARCHITECTURE="${ARCHITECTURE:-amd64}"

# Pre-flight checks
require_cmd() { command -v "$1" >/dev/null 2>&1 || { echo "Error: '$1' is required but not installed." >&2; exit 1; }; }
require_cmd unzip
require_cmd zip
require_cmd docker

echo "Building combined Ruby extension layer..."

# Clean and create directories
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR/combined-layer"

echo "Step 1: Building Ruby instrumentation layer from local source..."
# Build the local Ruby layer
cd "$SCRIPT_DIR/src"
# Ensure a fresh docker build to pick up Gemfile changes (e.g., google-protobuf)
docker rmi -f aws-otel-lambda-ruby-layer >/dev/null 2>&1 || true
./build.sh
cd "$SCRIPT_DIR"

# Extract the current layer
cd "$BUILD_DIR/combined-layer"
unzip -oq "$SCRIPT_DIR/src/build/opentelemetry-ruby-layer.zip" 2>/dev/null || {
    echo "Warning: Could not extract Ruby layer, checking for alternate name..."
    unzip -oq "$SCRIPT_DIR/src/build"/*.zip 2>/dev/null || {
        echo "Error: No Ruby layer zip file found"
        exit 1
    }
}
cd "$SCRIPT_DIR"

echo "Step 2: Building collector..."
# Build the collector
cd "$COLLECTOR_DIR"
make build GOARCH="$ARCHITECTURE"
cd "$SCRIPT_DIR"

# Copy collector files to combined layer
echo "Copying collector to combined layer..."
mkdir -p "$BUILD_DIR/combined-layer/extensions"
mkdir -p "$BUILD_DIR/combined-layer/collector-config"
cp "$COLLECTOR_DIR/build/extensions"/* "$BUILD_DIR/combined-layer/extensions/"
cp "$COLLECTOR_DIR/config.yaml" "$BUILD_DIR/combined-layer/collector-config/"

echo "Step 3: Optional: slimming Ruby gems (set KEEP_RUBY_GEM_VERSIONS=3.4.0,3.3.0 to keep specific versions)..."
if [ -n "${KEEP_RUBY_GEM_VERSIONS:-}" ]; then
  IFS=',' read -r -a keep_list <<< "$KEEP_RUBY_GEM_VERSIONS"
  find "$BUILD_DIR/combined-layer/ruby/gems" -maxdepth 1 -type d -name '3.*' | while read -r dir; do
    base=$(basename "$dir")
    base_mm=$(echo "$base" | cut -d. -f1-2)
    keep=false
    for v in "${keep_list[@]}"; do
      v_mm=$(echo "$v" | cut -d. -f1-2)
      if [ "$base" = "$v" ] || [ "$base_mm" = "$v_mm" ]; then keep=true; break; fi
    done
    if [ "$keep" = false ]; then
      echo "Pruning Ruby gems version $base"
      rm -rf "$dir"
    fi
  done
fi

echo "Step 4: Creating combined layer package..."
cd "$BUILD_DIR/combined-layer"

# Create build metadata at layer root (root of zip maps to /opt)
echo "Combined layer built on $(date)" > build-info.txt
echo "Architecture: $ARCHITECTURE" >> build-info.txt
echo "Collector version: $(cat $COLLECTOR_DIR/VERSION 2>/dev/null || echo 'unknown')" >> build-info.txt

# Ensure handler is executable
chmod +x otel-handler || true

# Package so that zip root maps directly to /opt
zip -qr ../otel-ruby-extension-layer.zip .
cd "$SCRIPT_DIR"

echo "Combined Ruby extension layer created: $BUILD_DIR/otel-ruby-extension-layer.zip"
echo "Layer contents:"
unzip -l "$BUILD_DIR/otel-ruby-extension-layer.zip" | head -20

echo "Build completed successfully!"