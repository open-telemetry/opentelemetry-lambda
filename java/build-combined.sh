#!/bin/bash

# Production-ready script to build a combined Java extension layer.
# This script combines our custom collector with the Java instrumentation
# built directly from the source code in this repository.

set -euo pipefail

# --- Script Setup ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$SCRIPT_DIR/build"
WORKSPACE_DIR="$BUILD_DIR/workspace"
# Collector is a sibling directory of `java/`
COLLECTOR_DIR="$SCRIPT_DIR/../collector"
# Navigate to the Java source directory (where gradlew lives)
JAVA_SRC_DIR="$SCRIPT_DIR"
ARCHITECTURE="${ARCHITECTURE:-amd64}"

# Pre-flight checks
require_cmd() { command -v "$1" >/dev/null 2>&1 || { echo "Error: '$1' is required but not installed." >&2; exit 1; }; }
require_cmd unzip
require_cmd zip

if [[ ! -e "$JAVA_SRC_DIR/gradlew" ]]; then
  echo "Error: gradlew not found at $JAVA_SRC_DIR/gradlew" >&2
  exit 1
fi
chmod +x "$JAVA_SRC_DIR/gradlew" || true

echo "Building combined Java extension layer (Arch: $ARCHITECTURE)..."

# 1. Clean and prepare the build environment
echo "--> Cleaning up previous build artifacts..."
rm -rf "$BUILD_DIR"
mkdir -p "$WORKSPACE_DIR"

# 2. Build the Java instrumentation layers from source
echo "--> Building Java instrumentation layers from source..."
# The parentheses run this in a subshell, so we don't have to cd back.
(
    cd "$JAVA_SRC_DIR"
    # Use gradle to build the agent and wrapper layers
    ./gradlew :layer-javaagent:build :layer-wrapper:build
)
echo "Java instrumentation build successful."

# 3. Extract the newly built layers into the workspace
echo "--> Extracting instrumentation layers..."
AGENT_ZIP="$JAVA_SRC_DIR/layer-javaagent/build/distributions/opentelemetry-javaagent-layer.zip"
WRAPPER_ZIP="$JAVA_SRC_DIR/layer-wrapper/build/distributions/opentelemetry-javawrapper-layer.zip"

if [[ ! -f "$AGENT_ZIP" ]]; then
  echo "Error: Expected artifact not found: $AGENT_ZIP" >&2
  exit 1
fi
if [[ ! -f "$WRAPPER_ZIP" ]]; then
  echo "Error: Expected artifact not found: $WRAPPER_ZIP" >&2
  exit 1
fi

unzip -oq -d "$WORKSPACE_DIR" "$AGENT_ZIP"
unzip -oq -d "$WORKSPACE_DIR" "$WRAPPER_ZIP"


# 4. Build the custom Go OTel Collector
echo "--> Building custom Go OTel Collector..."
(
    cd "$COLLECTOR_DIR"
    make build GOARCH="$ARCHITECTURE"
)
echo "Collector build successful."

# 5. Add the collector to the combined layer
echo "--> Adding collector to the combined layer..."
mkdir -p "$WORKSPACE_DIR/extensions"
mkdir -p "$WORKSPACE_DIR/collector-config"
cp "$COLLECTOR_DIR/build/extensions"/* "$WORKSPACE_DIR/extensions/"
cp "$COLLECTOR_DIR/config.yaml" "$WORKSPACE_DIR/collector-config/"

# 6. Create the final layer package
echo "--> Creating final layer .zip package..."
(
    cd "$WORKSPACE_DIR"
    zip -qr "$BUILD_DIR/otel-java-extension-layer-${ARCHITECTURE}.zip" .
)

echo ""
echo "✅ Combined Java extension layer created successfully!"
echo "   Location: $BUILD_DIR/otel-java-extension-layer-${ARCHITECTURE}.zip"

exit 0