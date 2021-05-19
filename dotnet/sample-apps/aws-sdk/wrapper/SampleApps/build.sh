#!/bin/sh

mkdir -p build/dotnet
git clone https://github.com/open-telemetry/opentelemetry-dotnet-contrib.git
dotnet add AwsSdkSample/*.csproj reference opentelemetry-dotnet-contrib/src/OpenTelemetry.Contrib.Instrumentation.AWSLambda/*.csproj
dotnet publish --output "./build/dotnet" --configuration "Release" --framework "netcoreapp3.1" /p:GenerateRuntimeConfigurationFiles=true --runtime linux-x64 --self-contained false 
cd build/dotnet || exit
zip -r ../function.zip *
