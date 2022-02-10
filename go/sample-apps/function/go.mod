module github.com/open-telemetry/opentelemetry-lambda/go/sample-apps/function

go 1.17

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

require (
	github.com/aws/aws-sdk-go-v2 v1.11.0 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.2.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.1.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.2.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.2.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.5.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.5.0 // indirect
	github.com/aws/smithy-go v1.9.0 // indirect
	github.com/cenkalti/backoff/v4 v4.1.1 // indirect
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	go.opentelemetry.io/contrib/detectors/aws/lambda v0.27.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.2.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.2.0 // indirect
	go.opentelemetry.io/otel/internal/metric v0.25.0 // indirect
	go.opentelemetry.io/otel/metric v0.25.0 // indirect
	go.opentelemetry.io/otel/sdk v1.2.0 // indirect
	go.opentelemetry.io/otel/trace v1.2.0 // indirect
	go.opentelemetry.io/proto/otlp v0.11.0 // indirect
	golang.org/x/net v0.0.0-20200822124328-c89045814202 // indirect
	golang.org/x/sys v0.0.0-20210423185535-09eb48e85fd7 // indirect
	golang.org/x/text v0.3.0 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/grpc v1.42.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)
