#!/bin/bash

# Build combined Node.js extension layer
# This script builds a production-ready combined layer that includes:
# 1. The official OpenTelemetry Node.js instrumentation layer (pinned version)
# 2. The custom Go OpenTelemetry Collector

set -euo pipefail

# Configuration
# Pin the upstream layer version for deterministic builds
UPSTREAM_LAYER_VERSION="layer-nodejs/0.15.0"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$SCRIPT_DIR/build"
WORKSPACE_DIR="$BUILD_DIR/workspace"
COLLECTOR_DIR="$SCRIPT_DIR/../../../collector"
INSTRUMENTATION_MANAGER="$SCRIPT_DIR/../../../utils/instrumentation-layer-manager.sh"
ARCHITECTURE="${ARCHITECTURE:-amd64}"

echo "Building combined Node.js extension layer (pinned to upstream version $UPSTREAM_LAYER_VERSION)..."

# Clean and create directories
rm -rf "$BUILD_DIR"
mkdir -p "$WORKSPACE_DIR"

echo "Step 1: Downloading official OpenTelemetry Node.js instrumentation layer..."
# Download the pinned upstream instrumentation layer and capture the output
DOWNLOAD_RESULT=$("$INSTRUMENTATION_MANAGER" download nodejs "$BUILD_DIR/temp" "$ARCHITECTURE" "$UPSTREAM_LAYER_VERSION" 2>&1)
DOWNLOAD_EXIT_CODE=$?

echo "$DOWNLOAD_RESULT" # Display the download output for verification

if [ $DOWNLOAD_EXIT_CODE -ne 0 ]; then
    echo "ERROR: Failed to download upstream Node.js instrumentation layer version $UPSTREAM_LAYER_VERSION"
    echo "This is a critical error for production builds. Exiting."
    exit 1
fi

# Extract instrumentation layer directly to workspace
if [ ! -d "$BUILD_DIR/temp/instrumentation" ]; then
    echo "ERROR: Downloaded instrumentation layer is missing expected structure"
    exit 1
fi
echo "Extracting Node.js instrumentation layer to workspace..."
cp -r "$BUILD_DIR/temp/instrumentation"/* "$WORKSPACE_DIR/"

echo "Step 2: Building custom OpenTelemetry Collector..."
# Build the collector
cd "$COLLECTOR_DIR"
if ! make build GOARCH="$ARCHITECTURE"; then
    echo "ERROR: Failed to build collector"
    exit 1
fi
cd "$SCRIPT_DIR"

echo "Step 3: Adding collector to combined layer..."
# Copy collector files to workspace
mkdir -p "$WORKSPACE_DIR/extensions"
mkdir -p "$WORKSPACE_DIR/collector-config"
cp "$COLLECTOR_DIR/build/extensions"/* "$WORKSPACE_DIR/extensions/"
cp "$COLLECTOR_DIR/config.yaml" "$WORKSPACE_DIR/collector-config/"

echo "Step 4: Creating build metadata..."
# Extract the exact release tag from the download output
ACTUAL_DOWNLOAD_TAG=$(echo "$DOWNLOAD_RESULT" | grep "Release tag:" | awk '{print $3}')
if [ -z "$ACTUAL_DOWNLOAD_TAG" ]; then
    ACTUAL_DOWNLOAD_TAG="unknown (check build log for details)"
fi

# Add build info to workspace root
cat > "$WORKSPACE_DIR/build-info.txt" << EOF
Combined Node.js extension layer
Built on: $(date -u +"%Y-%m-%d %H:%M:%S UTC")
Architecture: $ARCHITECTURE
Requested Upstream Node.js layer version: $UPSTREAM_LAYER_VERSION
Actual Downloaded Upstream Tag: $ACTUAL_DOWNLOAD_TAG
Collector version: $(cat "$COLLECTOR_DIR/VERSION" 2>/dev/null || echo 'unknown')
EOF

echo "Step 5: Creating final layer package..."
# Package the combined layer (workspace becomes /opt at runtime)
cd "$WORKSPACE_DIR"
zip -r ../otel-nodejs-extension-layer.zip .
cd "$SCRIPT_DIR"

# Clean up temporary files
rm -rf "$BUILD_DIR/temp"

echo "✅ Combined Node.js extension layer created: $BUILD_DIR/otel-nodejs-extension-layer.zip"
echo ""
echo "Layer contents preview:"
unzip -l "$BUILD_DIR/otel-nodejs-extension-layer.zip" | head -20
echo ""
echo "Build completed successfully!"