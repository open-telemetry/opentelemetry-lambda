#!/bin/bash

# Test script for combined layer builds
# This script tests that all combined layer build processes work correctly

set -euo pipefail

echo "Testing Combined Layer Build System"
echo "==================================="

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMP_DIR="/tmp/otel-combined-test-$$"
ARCHITECTURE="${ARCHITECTURE:-amd64}"

# Create temporary directory
mkdir -p "$TEMP_DIR"
cd "$SCRIPT_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

test_instrumentation_manager() {
    log_info "Testing instrumentation layer manager..."
    
    # Test that the script is executable
    if [[ ! -x "utils/instrumentation-layer-manager.sh" ]]; then
        log_error "instrumentation-layer-manager.sh is not executable"
        return 1
    fi
    
    # Test help command
    utils/instrumentation-layer-manager.sh help > /dev/null
    log_info "✓ Help command works"
    
    # Test list command
    utils/instrumentation-layer-manager.sh list > /dev/null
    log_info "✓ List command works"
    
    # Test check command for known languages
    for lang in nodejs python java; do
        if utils/instrumentation-layer-manager.sh check "$lang"; then
            log_info "✓ $lang instrumentation layer is available"
        else
            log_warn "✗ $lang instrumentation layer is not available (this may be expected)"
        fi
    done
}

test_collector_build() {
    log_info "Testing collector combined layer build..."
    
    cd collector
    
    # Test that we can build the collector
    if make build GOARCH="$ARCHITECTURE" > "$TEMP_DIR/collector-build.log" 2>&1; then
        log_info "✓ Collector builds successfully"
    else
        log_error "✗ Collector build failed"
        cat "$TEMP_DIR/collector-build.log"
        cd "$SCRIPT_DIR"
        return 1
    fi
    
    # Test combined package for nodejs (if available)
    if make package-combined LANGUAGE=nodejs GOARCH="$ARCHITECTURE" > "$TEMP_DIR/collector-combined.log" 2>&1; then
        log_info "✓ Collector combined layer for nodejs builds successfully"
        
        # Check that the combined layer was created
        if [[ -f "build/otel-nodejs-extension-$ARCHITECTURE.zip" ]]; then
            log_info "✓ Combined layer zip file created"
            
            # Check layer contents
            unzip -l "build/otel-nodejs-extension-$ARCHITECTURE.zip" > "$TEMP_DIR/layer-contents.txt"
            if grep -q "extensions" "$TEMP_DIR/layer-contents.txt" && grep -q "collector-config" "$TEMP_DIR/layer-contents.txt"; then
                log_info "✓ Combined layer contains expected collector components"
            else
                log_warn "? Combined layer may be missing collector components"
            fi
        else
            log_error "✗ Combined layer zip file not created"
        fi
    else
        log_warn "✗ Collector combined layer build failed (may be expected if dependencies missing)"
        cat "$TEMP_DIR/collector-combined.log" | head -20
    fi
    
    cd "$SCRIPT_DIR"
}

test_language_builds() {
    log_info "Testing language-specific combined builds..."
    
    # Test Node.js build (requires npm)
    if command -v npm > /dev/null; then
        log_info "Testing Node.js combined build..."
        cd nodejs/packages/layer
        
        if [[ -x "build-combined.sh" ]]; then
            log_info "✓ Node.js build-combined.sh is executable"
            
            # Check that package.json has the build-combined script
            if grep -q "build-combined" package.json; then
                log_info "✓ Node.js package.json has build-combined script"
            else
                log_warn "✗ Node.js package.json missing build-combined script"
            fi
        else
            log_error "✗ Node.js build-combined.sh is not executable"
        fi
        
        cd "$SCRIPT_DIR"
    else
        log_warn "Skipping Node.js test - npm not available"
    fi
    
    # Test Python build (requires docker)
    if command -v docker > /dev/null; then
        log_info "Testing Python combined build script..."
        cd python/src
        
        if [[ -x "build-combined.sh" ]]; then
            log_info "✓ Python build-combined.sh is executable"
        else
            log_error "✗ Python build-combined.sh is not executable"
        fi
        
        cd "$SCRIPT_DIR"
    else
        log_warn "Skipping Python test - docker not available"
    fi
    
    # Test Java build (requires gradlew)
    if [[ -x "java/gradlew" ]]; then
        log_info "Testing Java combined build script..."
        cd java
        
        if [[ -x "build-combined.sh" ]]; then
            log_info "✓ Java build-combined.sh is executable"
        else
            log_error "✗ Java build-combined.sh is not executable"
        fi
        
        cd "$SCRIPT_DIR"
    else
        log_warn "Skipping Java test - gradlew not available"
    fi
    
    # Test other language build scripts exist and are executable
    for lang in ruby dotnet go; do
        if [[ -x "$lang/build-combined.sh" ]]; then
            log_info "✓ $lang build-combined.sh is executable"
        else
            log_error "✗ $lang build-combined.sh is not executable"
        fi
    done
}

test_github_workflows() {
    log_info "Testing GitHub workflow files..."
    
    # Check that combined layer workflows exist
    for workflow in nodejs python java; do
        workflow_file=".github/workflows/release-combined-layer-$workflow.yml"
        if [[ -f "$workflow_file" ]]; then
            log_info "✓ $workflow combined layer workflow exists"
            
            # Basic syntax check - ensure it's valid YAML
            if command -v yq > /dev/null; then
                if yq eval . "$workflow_file" > /dev/null 2>&1; then
                    log_info "✓ $workflow workflow has valid YAML syntax"
                else
                    log_error "✗ $workflow workflow has invalid YAML syntax"
                fi
            fi
        else
            log_error "✗ $workflow combined layer workflow missing"
        fi
    done
}

run_tests() {
    log_info "Starting combined layer build system tests..."
    
    local test_count=0
    local failed_tests=0
    
    # Run tests
    for test_func in test_instrumentation_manager test_collector_build test_language_builds test_github_workflows; do
        test_count=$((test_count + 1))
        log_info "Running $test_func..."
        
        if ! $test_func; then
            failed_tests=$((failed_tests + 1))
            log_error "Test $test_func failed"
        fi
        
        echo ""
    done
    
    # Summary
    echo "Test Summary"
    echo "============"
    echo "Total tests: $test_count"
    echo "Failed tests: $failed_tests"
    echo "Passed tests: $((test_count - failed_tests))"
    
    if [[ $failed_tests -eq 0 ]]; then
        log_info "All tests passed! ✅"
        return 0
    else
        log_error "Some tests failed! ❌"
        return 1
    fi
}

# Cleanup function
cleanup() {
    if [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}

# Set up cleanup trap
trap cleanup EXIT

# Run tests
run_tests