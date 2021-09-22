module github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents

go 1.16

require (
  go.opentelemetry.io/collector v0.36.0
  github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusremotewriteexporter v0.36.0
  github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.36.0
  github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.36.0
  github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor v0.36.0
  github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.36.0
  github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor v0.36.0
)
