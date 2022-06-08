module github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents

go 1.17

require (
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusremotewriteexporter v0.51.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awsprometheusremotewriteexporter v0.51.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.51.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.51.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor v0.51.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.51.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor v0.51.0
	go.opentelemetry.io/collector v0.51.0
	go.uber.org/multierr v1.8.0
)
