#!/bin/bash

# OpenTelemetry Lambda Instrumentation Layer Manager
# This script detects and downloads available instrumentation layers from the official OpenTelemetry Lambda releases

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OTEL_LAMBDA_REPO="open-telemetry/opentelemetry-lambda"
RELEASES_API="https://api.github.com/repos/${OTEL_LAMBDA_REPO}/releases"

# Language to instrumentation layer mapping
# Based on OpenTelemetry Lambda releases structure
# Using a simple function-based mapping for better portability
get_layer_prefix_for_language() {
    local language="$1"
    case "$language" in
        "nodejs") echo "layer-nodejs" ;;
        "python") echo "layer-python" ;;
        "javaagent") echo "layer-javaagent" ;;
        "javawrapper") echo "layer-javawrapper" ;;
        "dotnet") echo "layer-dotnet" ;;
        *) return 1 ;;
    esac
}

# Function to get the latest release tag for a specific layer
get_latest_layer_release() {
    local layer_prefix="$1"
    
    # Get all releases and filter by the layer prefix
    curl -s "${RELEASES_API}" | \
        jq -r --arg prefix "$layer_prefix" \
        '.[] | select(.tag_name | startswith($prefix + "/")) | .tag_name' | \
        head -n 1
}

# Function to get download URL for a specific layer asset
get_layer_download_url() {
    local tag_name="$1"
    local asset_pattern="$2"
    
    curl -s "${RELEASES_API}/tags/${tag_name}" | \
        jq -r --arg pattern "$asset_pattern" \
        '.assets[] | select(.name | test($pattern)) | .browser_download_url'
}

# Function to download instrumentation layer for a language
download_instrumentation_layer() {
    local language="$1"
    local output_dir="$2"
    local architecture="${3:-amd64}"
    
    # Check if language has instrumentation layer
    local layer_prefix
    if ! layer_prefix=$(get_layer_prefix_for_language "$language"); then
        echo "No instrumentation layer available for $language"
        return 1
    fi
    echo "Looking for instrumentation layer for $language (prefix: $layer_prefix)"
    
    # Get latest release tag
    local latest_tag=$(get_latest_layer_release "$layer_prefix")
    if [[ -z "$latest_tag" ]]; then
        echo "No releases found for $layer_prefix"
        return 1
    fi
    
    echo "Found latest release: $latest_tag"
    
    # Determine asset pattern based on language and architecture
    local asset_pattern
    case "$language" in
        "nodejs")
            asset_pattern="opentelemetry-nodejs.*\.zip"
            ;;
        "python")
            asset_pattern="opentelemetry-python.*\.zip"
            ;;
        "javaagent")
            asset_pattern="opentelemetry-javaagent.*\.zip"
            ;;
        "javawrapper")
            asset_pattern="opentelemetry-javawrapper.*\.zip"
            ;;
        "dotnet")
            asset_pattern="opentelemetry-dotnet.*\.zip"
            ;;
        *)
            echo "Unknown asset pattern for language: $language"
            return 1
            ;;
    esac
    
    # Get download URL
    local download_url=$(get_layer_download_url "$latest_tag" "$asset_pattern")
    if [[ -z "$download_url" ]]; then
        echo "No downloadable asset found for $latest_tag with pattern $asset_pattern"
        return 1
    fi
    
    echo "Downloading instrumentation layer from: $download_url"
    
    # Create output directory
    mkdir -p "$output_dir"
    
    # Download and extract
    local filename="${latest_tag//\//-}-instrumentation.zip"
    local filepath="$output_dir/$filename"
    
    curl -L -o "$filepath" "$download_url"
    
    # Extract to instrumentation directory
    local extract_dir="$output_dir/instrumentation"
    mkdir -p "$extract_dir"
    unzip -q "$filepath" -d "$extract_dir"
    
    echo "Instrumentation layer extracted to: $extract_dir"
    echo "Release tag: $latest_tag"
    
    # Return the extract directory path and release tag
    echo "$extract_dir|$latest_tag"
}

# Function to check if instrumentation layer is available for a language
has_instrumentation_layer() {
    local language="$1"
    get_layer_prefix_for_language "$language" > /dev/null 2>&1
}

# Function to list all available instrumentation layers
list_available_layers() {
    echo "Available instrumentation layers:"
    for language in nodejs python javaagent javawrapper dotnet; do
        if layer_prefix=$(get_layer_prefix_for_language "$language"); then
            local latest_tag=$(get_latest_layer_release "$layer_prefix")
            if [[ -n "$latest_tag" ]]; then
                echo "  $language: $latest_tag"
            else
                echo "  $language: No releases found"
            fi
        fi
    done
}

# Main function
main() {
    local command="${1:-help}"
    
    case "$command" in
        "download")
            if [[ $# -lt 3 ]]; then
                echo "Usage: $0 download <language> <output_dir> [architecture]"
                exit 1
            fi
            download_instrumentation_layer "$2" "$3" "${4:-amd64}"
            ;;
        "check")
            if [[ $# -lt 2 ]]; then
                echo "Usage: $0 check <language>"
                exit 1
            fi
            if has_instrumentation_layer "$2"; then
                echo "Instrumentation layer available for $2"
                exit 0
            else
                echo "No instrumentation layer available for $2"
                exit 1
            fi
            ;;
        "list")
            list_available_layers
            ;;
        "help"|*)
            echo "OpenTelemetry Lambda Instrumentation Layer Manager"
            echo ""
            echo "Usage: $0 <command> [options]"
            echo ""
            echo "Commands:"
            echo "  download <language> <output_dir> [architecture]  Download instrumentation layer"
            echo "  check <language>                                 Check if instrumentation layer is available"  
            echo "  list                                             List all available instrumentation layers"
            echo "  help                                             Show this help message"
            echo ""
            echo "Supported languages: nodejs python javaagent javawrapper dotnet"
            ;;
    esac
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi