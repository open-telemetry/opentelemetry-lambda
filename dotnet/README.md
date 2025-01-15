# OpenTelemetry Lambda .NET

Nuget package for running .NET applications on AWS Lambda with OpenTelemetry.

## Provided SDK

[OpenTelemetry Lambda SDK for .NET](https://github.com/open-telemetry/opentelemetry-dotnet-contrib/tree/main/src/OpenTelemetry.Instrumentation.AWSLambda) includes tracing APIs to instrument Lambda handler and is provided on [Nuget](https://www.nuget.org/packages/OpenTelemetry.Instrumentation.AWSLambda/1.1.0-beta2). Follow the instructions on [user guide](https://aws-otel.github.io/docs/getting-started/lambda/lambda-dotnet#instrumentation) to manually instrument the Lambda handler.
For other instrumentations, such as http, you'll need to include the corresponding library instrumentation from the [instrumentation project](https://github.com/open-telemetry/opentelemetry-dotnet) and modify your code to initialize it in your function.

## Provided Layer

[OpenTelemetry Lambda Layer for Collector](https://aws-otel.github.io/docs/getting-started/lambda/lambda-dotnet#lambda-layer) includes OpenTelemetry Collector for Lambda components. Follow [user guide](https://aws-otel.github.io/docs/getting-started/lambda/lambda-dotnet#enable-tracing) to apply this layer to your Lambda handler that's already been instrumented with OpenTelemetry Lambda .NET SDK to enable end-to-end tracing.

## Sample application

The [sample application](https://github.com/open-telemetry/opentelemetry-lambda/blob/main/dotnet/sample-apps/aws-sdk/wrapper/SampleApps/AwsSdkSample/Function.cs) shows the manual instrumentations of OpenTelemetry Lambda .NET SDK on a Lambda handler that triggers a downstream request to AWS S3.
