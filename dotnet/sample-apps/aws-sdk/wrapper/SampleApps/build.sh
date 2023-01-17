#!/bin/sh
set -e

GOARCH=${GOARCH-amd64}

# Chosen from Microsoft's documentation for Runtime Identifier (RIDs)
# See more: https://docs.microsoft.com/en-us/dotnet/core/rid-catalog#linux-rids
if [ "$GOARCH" = "amd64" ]; then
    DOTNET_LINUX_ARCH=x64
elif [ "$GOARCH" = "arm64" ]; then
    DOTNET_LINUX_ARCH=arm64
else
    echo "Invalid GOARCH value $(GOARCH) received."
    exit 2
fi

mkdir -p build/dotnet
dotnet publish \
    --output "./build/dotnet" \
    --configuration "Release" \
    --framework "net6.0" /p:GenerateRuntimeConfigurationFiles=true \
    --runtime linux-$DOTNET_LINUX_ARCH \
    --self-contained false
cd build/dotnet
zip -r ../function.zip ./*
