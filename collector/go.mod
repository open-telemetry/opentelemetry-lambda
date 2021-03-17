module github.com/open-telemetry/opentelemetry-lambda/collector

go 1.14

require (
	github.com/aws-observability/aws-otel-collector/pkg/lambdacomponents v0.6.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	go.opentelemetry.io/collector v0.20.0
	go.uber.org/zap v1.16.0
)

replace github.com/open-telemetry/opentelemetry-collector-contrib/internal/awsxray => github.com/open-telemetry/opentelemetry-collector-contrib/internal/awsxray v0.20.0
