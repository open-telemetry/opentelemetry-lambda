module github.com/open-telemetry/opentelemetry-lambda/go/sample-apps/function

go 1.16

require (
	github.com/aws/aws-lambda-go v1.24.0
	github.com/aws/aws-sdk-go-v2/config v1.4.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.11.0
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda v0.21.0
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws v0.21.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.21.0
	go.opentelemetry.io/otel v1.0.0-RC1
)

replace (
	go.opentelemetry.io/contrib/detectors/aws/lambda => github.com/garrettwegan/opentelemetry-go-contrib/detectors/aws/lambda latest
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda => github.com/garrettwegan/opentelemetry-go-contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda latest
)
