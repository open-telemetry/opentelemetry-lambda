module github.com/open-telemetry/opentelemetry-lambda/collector

go 1.16

replace github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents => ./lambdacomponents

require (
	github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents v0.0.0
	go.opentelemetry.io/collector v0.29.0
)
