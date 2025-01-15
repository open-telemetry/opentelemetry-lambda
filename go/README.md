# OpenTelemetry Lambda Go

Examples of Go applications on AWS Lambda with OpenTelemetry.

## Provided SDK

[OpenTelemetry Lambda SDK for Go](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/github.com/aws/aws-lambda-go/otellambda) includes tracing APIs to instrument Lambda handler.
For other instrumentations, such as http, you'll need to include the corresponding library instrumentation from the [instrumentation project](https://github.com/open-telemetry/opentelemetry-go) and modify your code to use it in your function.

## Provided Layer

[OpenTelemetry Lambda Layer for Collector](https://aws-otel.github.io/docs/getting-started/lambda/lambda-go#lambda-layer) includes OpenTelemetry Collector for Lambda components. Follow [user guide](https://aws-otel.github.io/docs/getting-started/lambda/lambda-go#enable-tracing) to apply this layer to your Lambda handler that's already been instrumented with OpenTelemetry Lambda .NET SDK to enable end-to-end tracing.

## Sample application

The [sample application](https://github.com/open-telemetry/opentelemetry-lambda/tree/main/go/sample-apps/function/function.go) shows the manual instrumentations of OpenTelemetry Lambda Go SDK on a Lambda handler that triggers downstream requests to AWS S3 and HTTP.
