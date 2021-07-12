#!/bin/sh

mkdir -p build/go
dotnet publish --output "./build/dotnet" --configuration "Release" --framework "netcoreapp3.1" /p:GenerateRuntimeConfigurationFiles=true --runtime linux-x64 --self-contained false
GOOS=linux GOARCH=amd64 go build -o build/go/bootstrap .
cd build/go || exit
zip bootstrap.zip bootstrap
