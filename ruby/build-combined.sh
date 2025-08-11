#!/bin/bash

# Build combined Ruby extension layer
# This script builds a combined layer that includes:
# 1. The custom Ruby instrumentation layer (current layer)
# 2. The custom collector
# 3. The upstream OpenTelemetry Ruby instrumentation layer (if available)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$SCRIPT_DIR/build"
COLLECTOR_DIR="$SCRIPT_DIR/../collector"
INSTRUMENTATION_MANAGER="$SCRIPT_DIR/../utils/instrumentation-layer-manager.sh"
ARCHITECTURE="${ARCHITECTURE:-amd64}"

echo "Building combined Ruby extension layer..."

# Clean and create directories
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR/combined-layer"

echo "Step 1: Building current Ruby layer..."
# Build the current Ruby layer
cd "$SCRIPT_DIR/src"
./build.sh
cd "$SCRIPT_DIR"

# Extract the current layer
cd "$BUILD_DIR/combined-layer"
unzip -q ../src/build/opentelemetry-ruby-layer.zip 2>/dev/null || {
    echo "Warning: Could not extract Ruby layer, checking for alternate name..."
    unzip -q ../src/build/*.zip 2>/dev/null || {
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
cp "$COLLECTOR_DIR/config"* "$BUILD_DIR/combined-layer/collector-config/"

echo "Step 3: Checking for upstream instrumentation layer..."
# Check if upstream OpenTelemetry instrumentation layer is available
if "$INSTRUMENTATION_MANAGER" check ruby; then
    echo "Downloading upstream OpenTelemetry Ruby instrumentation layer..."
    TEMP_DIR="$BUILD_DIR/temp"
    mkdir -p "$TEMP_DIR"
    
    # Download the upstream instrumentation layer
    RESULT=$("$INSTRUMENTATION_MANAGER" download ruby "$TEMP_DIR" "$ARCHITECTURE" 2>&1) || {
        echo "Warning: Could not download upstream instrumentation layer: $RESULT"
        echo "Continuing with custom instrumentation only..."
    }
    
    if [ -d "$TEMP_DIR/instrumentation" ]; then
        echo "Including upstream instrumentation layer..."
        mkdir -p "$BUILD_DIR/combined-layer/upstream-ruby"
        cp -r "$TEMP_DIR/instrumentation"/* "$BUILD_DIR/combined-layer/upstream-ruby/"
        
        # Save version info
        echo "$RESULT" | grep "Release tag:" > "$BUILD_DIR/combined-layer/upstream-instrumentation-version.txt" 2>/dev/null || echo "unknown" > "$BUILD_DIR/combined-layer/upstream-instrumentation-version.txt"
        
        rm -rf "$TEMP_DIR"
        echo "Upstream instrumentation layer included."
    fi
else
    echo "No upstream instrumentation layer available for Ruby"
fi

echo "Step 4: Creating combined layer package..."
cd "$BUILD_DIR"

# Create proper Lambda layer directory structure with /opt/ prefix
mkdir -p lambda-layer/opt
mv combined-layer/* lambda-layer/opt/

# Create version info file in the opt directory
echo "Combined layer built on $(date)" > lambda-layer/opt/build-info.txt
echo "Architecture: $ARCHITECTURE" >> lambda-layer/opt/build-info.txt
echo "Collector version: $(cat $COLLECTOR_DIR/VERSION 2>/dev/null || echo 'unknown')" >> lambda-layer/opt/build-info.txt

# Package the combined layer with correct structure
cd lambda-layer
zip -r ../otel-ruby-extension-layer.zip *
cd "$SCRIPT_DIR"

echo "Combined Ruby extension layer created: $BUILD_DIR/otel-ruby-extension-layer.zip"
echo "Layer contents:"
unzip -l "$BUILD_DIR/otel-ruby-extension-layer.zip" | head -20

echo "Build completed successfully!"