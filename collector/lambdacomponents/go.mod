module github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents

go 1.18

require (
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusremotewriteexporter v0.68.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/sigv4authextension v0.68.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.68.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.68.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor v0.68.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.68.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor v0.68.0
	go.opentelemetry.io/collector v0.68.0
	go.opentelemetry.io/collector/component v0.68.0
	go.opentelemetry.io/collector/exporter/loggingexporter v0.68.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.68.0
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.68.0
	go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.68.0
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.68.0
	go.uber.org/multierr v1.9.0
)

require (
	github.com/alecthomas/participle/v2 v2.0.0-beta.5 // indirect
	github.com/antonmedv/expr v1.9.0 // indirect
	github.com/aws/aws-sdk-go-v2 v1.17.3 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.18.6 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.6 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.21 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.27 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.21 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.28 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.21 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.11.27 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.13.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.17.7 // indirect
	github.com/aws/smithy-go v1.13.5 // indirect
	github.com/cenkalti/backoff/v4 v4.2.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/iancoleman/strcase v0.2.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.15.13 // indirect
	github.com/knadh/koanf v1.4.4 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/go-grpc-compression v1.1.17 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.68.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter v0.68.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl v0.68.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/resourcetotelemetry v0.68.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus v0.68.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheusremotewrite v0.68.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/common v0.39.0 // indirect
	github.com/prometheus/prometheus v0.40.7 // indirect
	github.com/rs/cors v1.8.2 // indirect
	github.com/shirou/gopsutil/v3 v3.22.10 // indirect
	github.com/tidwall/gjson v1.10.2 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tidwall/tinylru v1.1.0 // indirect
	github.com/tidwall/wal v1.1.7 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/collector/confmap v0.68.0 // indirect
	go.opentelemetry.io/collector/consumer v0.68.0 // indirect
	go.opentelemetry.io/collector/featuregate v0.68.0 // indirect
	go.opentelemetry.io/collector/pdata v1.0.0-rc2 // indirect
	go.opentelemetry.io/collector/semconv v0.68.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.37.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.37.0 // indirect
	go.opentelemetry.io/otel v1.11.2 // indirect
	go.opentelemetry.io/otel/metric v0.34.0 // indirect
	go.opentelemetry.io/otel/trace v1.11.2 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/exp v0.0.0-20221205204356-47842c84f3db // indirect
	golang.org/x/net v0.4.0 // indirect
	golang.org/x/sys v0.3.0 // indirect
	golang.org/x/text v0.5.0 // indirect
	google.golang.org/genproto v0.0.0-20221027153422-115e99e71e1c // indirect
	google.golang.org/grpc v1.51.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

// ambiguous import: found package cloud.google.com/go/compute/metadata in multiple modules:
//        cloud.google.com/go
//        cloud.google.com/go/compute
// Force cloud.google.com/go to be at least v0.107.0, so that the metadata is not present.
replace cloud.google.com/go => cloud.google.com/go v0.107.0
