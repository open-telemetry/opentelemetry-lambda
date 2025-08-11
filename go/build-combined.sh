#!/bin/bash

# Build combined Go extension layer
# This script builds a combined layer that includes:
# 1. The custom collector (Go doesn't have auto-instrumentation, only manual instrumentation)
# 2. The upstream OpenTelemetry Go instrumentation layer (if available)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$SCRIPT_DIR/build"
COLLECTOR_DIR="$SCRIPT_DIR/../collector"
INSTRUMENTATION_MANAGER="$SCRIPT_DIR/../utils/instrumentation-layer-manager.sh"
ARCHITECTURE="${ARCHITECTURE:-amd64}"

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

echo "Step 2: Checking for upstream instrumentation layer..."
# Note: Go typically doesn't have auto-instrumentation layers like other languages
# but we'll check anyway in case upstream releases one
if "$INSTRUMENTATION_MANAGER" check go; then
    echo "Downloading upstream OpenTelemetry Go instrumentation layer..."
    TEMP_DIR="$BUILD_DIR/temp"
    mkdir -p "$TEMP_DIR"
    
    # Download the upstream instrumentation layer
    RESULT=$("$INSTRUMENTATION_MANAGER" download go "$TEMP_DIR" "$ARCHITECTURE" 2>&1) || {
        echo "Warning: Could not download upstream instrumentation layer: $RESULT"
        echo "Continuing with collector only..."
    }
    
    if [ -d "$TEMP_DIR/instrumentation" ]; then
        echo "Including upstream instrumentation layer..."
        cp -r "$TEMP_DIR/instrumentation"/* "$BUILD_DIR/combined-layer/"
        
        # Save version info
        echo "$RESULT" | grep "Release tag:" > "$BUILD_DIR/combined-layer/upstream-instrumentation-version.txt" 2>/dev/null || echo "unknown" > "$BUILD_DIR/combined-layer/upstream-instrumentation-version.txt"
        
        rm -rf "$TEMP_DIR"
        echo "Upstream instrumentation layer included."
    fi
else
    echo "No upstream instrumentation layer available for Go (expected - Go uses manual instrumentation)"
fi

echo "Step 3: Creating combined layer package..."
cd "$BUILD_DIR"

# Create proper Lambda layer directory structure with /opt/ prefix
mkdir -p lambda-layer/opt
mv combined-layer/* lambda-layer/opt/

# Create version info file in the opt directory
echo "Combined layer built on $(date)" > lambda-layer/opt/build-info.txt
echo "Architecture: $ARCHITECTURE" >> lambda-layer/opt/build-info.txt
echo "Collector version: $(cat $COLLECTOR_DIR/VERSION 2>/dev/null || echo 'unknown')" >> lambda-layer/opt/build-info.txt
echo "Note: Go uses manual instrumentation - this layer provides the collector for Go applications" >> lambda-layer/opt/build-info.txt

# Package the combined layer with correct structure
cd lambda-layer
zip -r ../otel-go-extension-layer.zip *
cd "$SCRIPT_DIR"

echo "Combined Go extension layer created: $BUILD_DIR/otel-go-extension-layer.zip"
echo "Layer contents:"
unzip -l "$BUILD_DIR/otel-go-extension-layer.zip" | head -20

echo "Build completed successfully!"