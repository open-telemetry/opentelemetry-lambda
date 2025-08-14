#!/bin/bash

# Build combined Node.js extension layer
# This script builds a production-ready combined layer that includes:
# 1. The official OpenTelemetry Node.js instrumentation layer (pinned version)
# 2. The custom Go OpenTelemetry Collector

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$SCRIPT_DIR/build"
WORKSPACE_DIR="$BUILD_DIR/workspace"
COLLECTOR_DIR="$SCRIPT_DIR/../../../collector"
ARCHITECTURE="${ARCHITECTURE:-amd64}"

# Pre-flight checks
require_cmd() { command -v "$1" >/dev/null 2>&1 || { echo "Error: '$1' is required but not installed." >&2; exit 1; }; }
require_cmd unzip
require_cmd zip
require_cmd npm

echo "Building combined Node.js extension layer from local sources..."

# Clean and create directories
rm -rf "$BUILD_DIR"
mkdir -p "$WORKSPACE_DIR"

echo "Step 1: Building OpenTelemetry Node.js instrumentation layer from local source..."
# Build via workspace so root devDependencies (rimraf, bestzip, etc.) are available
(
  cd "$SCRIPT_DIR/../.." && \
  npm ci --workspaces && \
  npm run build -w @opentelemetry-lambda/sdk-layer
)
LOCAL_LAYER_ZIP="$SCRIPT_DIR/build/layer.zip"
if [ ! -f "$LOCAL_LAYER_ZIP" ]; then
  echo "ERROR: Local Node.js layer artifact not found: $LOCAL_LAYER_ZIP" >&2
  exit 1
fi
echo "Extracting locally built Node.js layer to workspace..."
mkdir -p "$WORKSPACE_DIR"
unzip -oq -d "$WORKSPACE_DIR" "$LOCAL_LAYER_ZIP"

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
cat > "$WORKSPACE_DIR/build-info.txt" << EOF
Combined Node.js extension layer (built from local source)
Built on: $(date -u +"%Y-%m-%d %H:%M:%S UTC")
Architecture: $ARCHITECTURE
Layer package hash: $(shasum "$LOCAL_LAYER_ZIP" 2>/dev/null | awk '{print $1}')
Collector version: $(cat "$COLLECTOR_DIR/VERSION" 2>/dev/null || echo 'unknown')
Git commit: $(git -C "$SCRIPT_DIR/../../.." rev-parse --short HEAD 2>/dev/null || echo 'unknown')
EOF

echo "Step 5: Creating final layer package..."
# Package the combined layer (workspace becomes /opt at runtime)
cd "$WORKSPACE_DIR"
zip -qr ../otel-nodejs-extension-layer.zip .
cd "$SCRIPT_DIR"

# Clean up temporary files
:

echo "✅ Combined Node.js extension layer created: $BUILD_DIR/otel-nodejs-extension-layer.zip"
echo ""
echo "Layer contents preview:"
unzip -l "$BUILD_DIR/otel-nodejs-extension-layer.zip" | head -20
echo ""
echo "Build completed successfully!"