#!/bin/bash

# Build Go extension layer (collector-only)
# Go uses manual instrumentation. This script builds only the custom collector
# and packages it into a Lambda layer zip.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$SCRIPT_DIR/build"
COLLECTOR_DIR="$SCRIPT_DIR/../collector"
ARCHITECTURE="${ARCHITECTURE:-amd64}"

# Pre-flight checks
require_cmd() { command -v "$1" >/dev/null 2>&1 || { echo "Error: '$1' is required but not installed." >&2; exit 1; }; }
require_cmd unzip
require_cmd zip

echo "Building combined Go extension layer..."

# Clean and create directories
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR/combined-layer"

echo "Step 1: Building collector..."
# Build the collector
cd "$COLLECTOR_DIR"
make build GOARCH="$ARCHITECTURE"
cd "$SCRIPT_DIR"

# Copy collector files to combined layer
echo "Copying collector to combined layer..."
mkdir -p "$BUILD_DIR/combined-layer/extensions"
mkdir -p "$BUILD_DIR/combined-layer/collector-config"
cp "$COLLECTOR_DIR/build/extensions"/* "$BUILD_DIR/combined-layer/extensions/"
cp "$COLLECTOR_DIR/config"* "$BUILD_DIR/combined-layer/collector-config/"


echo "Step 2: Creating combined layer package..."
# Package so that zip root maps directly to /opt (do NOT include an extra top-level opt/)
cd "$BUILD_DIR/combined-layer"

# Create version info file at the layer root (becomes /opt/build-info.txt)
echo "Combined layer built on $(date)" > build-info.txt
echo "Architecture: $ARCHITECTURE" >> build-info.txt
echo "Collector version: $(cat $COLLECTOR_DIR/VERSION 2>/dev/null || echo 'unknown')" >> build-info.txt
echo "Note: Go uses manual instrumentation - this layer provides the collector for Go applications" >> build-info.txt

# Zip the contents of combined-layer so that extensions/ -> /opt/extensions and collector-config/ -> /opt/collector-config
zip -qr ../otel-go-extension-layer.zip .
cd "$SCRIPT_DIR"

echo "Combined Go extension layer created: $BUILD_DIR/otel-go-extension-layer.zip"
echo "Layer contents:"
unzip -l "$BUILD_DIR/otel-go-extension-layer.zip" | head -20

echo "Build completed successfully!"