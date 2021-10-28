# OpenTelemetry Lambda Go

Layer for running Go applications on AWS Lambda with OpenTelemetry.

## Provided SDK

[OpenTelemetry Lambda SDK for Go](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/github.com/aws/aws-lambda-go/otellambda) includes tracing APIs to instrument Lambda handler. Follow the instructions on [user guide](https://aws-otel.github.io/docs/getting-started/lambda/lambda-go) to manually instrument the Lambda handler.
For other instrumentations, such as http, you'll need to include the corresponding library instrumentation from the [instrumentation project](https://github.com/open-telemetry/opentelemetry-go) and modify your code to use it in your function.

## Sample application

The [sample application]() shows the manual instrumentations of OpenTelemetry Lambda Go SDK on a Lambda handler that triggers downstream requests to AWS S3 and HTTP.
