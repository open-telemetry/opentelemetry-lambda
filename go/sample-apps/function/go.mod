module github.com/open-telemetry/opentelemetry-lambda/go/sample-apps/function

go 1.16

require (
	github.com/aws/aws-lambda-go v1.24.0
	github.com/aws/aws-sdk-go-v2/config v1.4.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.11.0
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda v0.22.0
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda/xrayconfig v0.22.0
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws v0.22.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.22.0
	go.opentelemetry.io/otel v1.0.0-RC2
)

replace (
	go.opentelemetry.io/contrib/detectors/aws/lambda => github.com/garrettwegan/opentelemetry-go-contrib/detectors/aws/lambda v0.0.0-20210730164323-986e366f4c23
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda => github.com/garrettwegan/opentelemetry-go-contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda v0.0.0-20210730164323-986e366f4c23
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda/xrayconfig => github.com/garrettwegan/opentelemetry-go-contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda/xrayconfig v0.0.0-20210730201622-eef81a9505f4
)
