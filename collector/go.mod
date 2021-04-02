module github.com/open-telemetry/opentelemetry-lambda/collector

go 1.16

replace (
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/awsxray => github.com/open-telemetry/opentelemetry-collector-contrib/internal/awsxray v0.22.0
	github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents => ./lambdacomponents
)

require (
	github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents v0.1.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	go.opentelemetry.io/collector v0.22.0
	go.uber.org/zap v1.16.0
)
