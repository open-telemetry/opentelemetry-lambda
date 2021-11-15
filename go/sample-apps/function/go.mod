module github.com/open-telemetry/opentelemetry-lambda/go/sample-apps/function

go 1.16

require (
	github.com/aws/aws-lambda-go v1.27.0
	github.com/aws/aws-sdk-go-v2/config v1.4.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.11.0
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda v0.27.0
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda/xrayconfig v0.27.0
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws v0.27.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.27.0
	go.opentelemetry.io/contrib/propagators/aws v1.2.0
	go.opentelemetry.io/otel v1.2.0
)
