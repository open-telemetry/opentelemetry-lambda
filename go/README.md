# OpenTelemetry Lambda Go

Layer for running Go applications on AWS Lambda with OpenTelemetry.

## Provided SDK

[OpenTelemetry Lambda SDK for Go]() includes tracing APIs to instrument Lambda handler. Follow the instructions on [user guide]() to manually instrument the Lambda handler.
For other instrumentations, such as http, you'll need to include the corresponding library instrumentation from the [instrumentation project](https://github.com/open-telemetry/opentelemetry-go) and modify your code to use it in your function.

## Provided Layer

[OpenTelemetry Lambda Layer for Collector]() includes OpenTelemetry Collector for Lambda components. Follow [user guide]() to apply this layer to your Lambda handler that's already been instrumented with OpenTelemetry Lambda Go SDK to enable end-to-end tracing.

## Sample application

The [sample application]() shows the manual instrumentations of OpenTelemetry Lambda .NET SDK on a Lambda handler that triggers downstream requests to AWS S3 and HTTP.
