module github.com/open-telemetry/opentelemetry-lambda/collector

go 1.20

replace github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents => ./lambdacomponents

replace github.com/open-telemetry/opentelemetry-lambda/collector/lambdalifecycle => ./lambdalifecycle

replace github.com/open-telemetry/opentelemetry-lambda/collector/processor/coldstartprocessor => ./processor/coldstartprocessor

replace github.com/open-telemetry/opentelemetry-lambda/collector/processor/decoupleprocessor => ./processor/decoupleprocessor

replace github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver => ./receiver/telemetryapireceiver

// fixes ambiguous import error: found package cloud.google.com/go/compute/metadata in multiple modules:
//        cloud.google.com/go
//        cloud.google.com/go/compute
// Force cloud.google.com/go to be at least v0.107.0, so that the metadata is not present.
replace cloud.google.com/go => cloud.google.com/go v0.107.0

require (
	github.com/golang-collections/go-datastructures v0.0.0-20150211160725-59788d5eb259
	github.com/open-telemetry/opentelemetry-collector-contrib/confmap/provider/s3provider v0.91.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.91.0
	github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents v0.91.0
	github.com/open-telemetry/opentelemetry-lambda/collector/lambdalifecycle v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.8.4
<<<<<<< HEAD
	go.opentelemetry.io/collector/component v0.91.0
	go.opentelemetry.io/collector/confmap v0.91.0
	go.opentelemetry.io/collector/otelcol v0.91.0
=======
	go.opentelemetry.io/collector/component v0.88.0
	go.opentelemetry.io/collector/confmap v0.88.0
	go.opentelemetry.io/collector/otelcol v0.88.0
>>>>>>> b497143 (Revert)
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.26.0
)

require (
	cloud.google.com/go/compute v1.23.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.4-0.20230617002413-005d2dfb6b68 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.4.2 // indirect
<<<<<<< HEAD
<<<<<<< HEAD
	github.com/alecthomas/participle/v2 v2.1.1 // indirect
	github.com/aws/aws-sdk-go-v2 v1.23.5 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.5.3 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.25.12 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.16.10 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.14.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.2.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.5.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.2.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.10.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.2.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.10.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.16.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.47.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.18.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.21.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.26.3 // indirect
	github.com/aws/smithy-go v1.18.1 // indirect
=======
=======
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.20.0 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/Showmax/go-fqdn v1.0.0 // indirect
>>>>>>> d53b4b7 (Updated)
	github.com/alecthomas/participle/v2 v2.1.0 // indirect
	github.com/antonmedv/expr v1.15.3 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/aws/aws-sdk-go v1.45.26 // indirect
	github.com/aws/aws-sdk-go-v2 v1.21.2 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.19.0 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.43 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.43 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.37 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.45 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.5.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.37 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.9.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.19.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.15.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.17.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.23.2 // indirect
	github.com/aws/smithy-go v1.15.0 // indirect
>>>>>>> 3776821 (Revert)
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
<<<<<<< HEAD
<<<<<<< HEAD
	github.com/expr-lang/expr v1.15.6 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
=======
=======
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/docker v24.0.6+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emicklei/go-restful/v3 v3.10.2 // indirect
	github.com/fatih/color v1.15.0 // indirect
>>>>>>> d53b4b7 (Updated)
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
>>>>>>> 3776821 (Revert)
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/jsonpointer v0.20.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
<<<<<<< HEAD
<<<<<<< HEAD
	github.com/google/uuid v1.4.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.18.1 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
=======
	github.com/google/uuid v1.3.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.18.0 // indirect
>>>>>>> 3776821 (Revert)
=======
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.18.0 // indirect
	github.com/hashicorp/consul/api v1.25.1 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.5.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
>>>>>>> d53b4b7 (Updated)
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.4 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/providers/confmap v0.1.0 // indirect
	github.com/knadh/koanf/v2 v2.0.1 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
<<<<<<< HEAD
<<<<<<< HEAD
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
=======
=======
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
>>>>>>> d53b4b7 (Updated)
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
>>>>>>> 3776821 (Revert)
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.1-0.20220423185008-bf980b35cac4 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/go-grpc-compression v1.2.2 // indirect
<<<<<<< HEAD
<<<<<<< HEAD
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusremotewriteexporter v0.91.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/sigv4authextension v0.91.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.91.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter v0.91.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl v0.91.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatautil v0.91.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/resourcetotelemetry v0.91.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus v0.91.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheusremotewrite v0.91.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.91.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.91.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor v0.91.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.91.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor v0.91.0 // indirect
	github.com/open-telemetry/opentelemetry-lambda/collector/processor/coldstartprocessor v0.91.0 // indirect
	github.com/open-telemetry/opentelemetry-lambda/collector/processor/decoupleprocessor v0.0.0-00010101000000-000000000000 // indirect
	github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver v0.91.0 // indirect
=======
=======
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
>>>>>>> d53b4b7 (Updated)
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusremotewriteexporter v0.88.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/sigv4authextension v0.88.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/ecsutil v0.88.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.88.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.88.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter v0.88.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.88.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/metadataproviders v0.88.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl v0.88.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatautil v0.88.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/resourcetotelemetry v0.88.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus v0.88.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheusremotewrite v0.88.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.88.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.88.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor v0.88.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.88.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.88.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor v0.88.0 // indirect
	github.com/open-telemetry/opentelemetry-lambda/collector/processor/coldstartprocessor v0.88.0 // indirect
	github.com/open-telemetry/opentelemetry-lambda/collector/processor/decoupleprocessor v0.0.0-00010101000000-000000000000 // indirect
	github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver v0.88.0 // indirect
<<<<<<< HEAD
>>>>>>> 3776821 (Revert)
=======
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc4 // indirect
	github.com/openshift/api v3.9.0+incompatible // indirect
	github.com/openshift/client-go v0.0.0-20210521082421-73d9475a9142 // indirect
	github.com/pkg/errors v0.9.1 // indirect
>>>>>>> d53b4b7 (Updated)
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/client_golang v1.17.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/prometheus/prometheus v0.48.0 // indirect
	github.com/prometheus/statsd_exporter v0.22.7 // indirect
	github.com/rs/cors v1.10.1 // indirect
	github.com/shirou/gopsutil/v3 v3.23.11 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
<<<<<<< HEAD
<<<<<<< HEAD
	github.com/solarwinds/opentelemetry-collector-contrib/extension/solarwindsapmsettingsextension v0.0.0-20231103200316-8f39128f4706 // indirect
	github.com/solarwindscloud/apm-proto v0.0.0-20230912021952-7c5e94624446 // indirect
	github.com/spf13/cobra v1.8.0 // indirect
=======
=======
	github.com/solarwinds/opentelemetry-collector-contrib/extension/solarwindsapmsettingsextension v0.0.0-20231129225102-9b993ae9a355 // indirect
	github.com/solarwindscloud/apm-proto v0.0.0-20230912021952-7c5e94624446 // indirect
>>>>>>> d53b4b7 (Updated)
	github.com/spf13/cobra v1.7.0 // indirect
>>>>>>> 3776821 (Revert)
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/tidwall/gjson v1.10.2 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tidwall/tinylru v1.1.0 // indirect
	github.com/tidwall/wal v1.1.7 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	go.opencensus.io v0.24.0 // indirect
<<<<<<< HEAD
	go.opentelemetry.io/collector v0.91.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.91.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v0.91.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.91.0 // indirect
	go.opentelemetry.io/collector/config/confighttp v0.91.0 // indirect
	go.opentelemetry.io/collector/config/confignet v0.91.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v0.91.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.91.0 // indirect
	go.opentelemetry.io/collector/config/configtls v0.91.0 // indirect
	go.opentelemetry.io/collector/config/internal v0.91.0 // indirect
	go.opentelemetry.io/collector/connector v0.91.0 // indirect
	go.opentelemetry.io/collector/consumer v0.91.0 // indirect
	go.opentelemetry.io/collector/exporter v0.91.0 // indirect
	go.opentelemetry.io/collector/exporter/loggingexporter v0.91.0 // indirect
	go.opentelemetry.io/collector/exporter/otlpexporter v0.91.0 // indirect
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.91.0 // indirect
	go.opentelemetry.io/collector/extension v0.91.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.91.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.0.0 // indirect
	go.opentelemetry.io/collector/pdata v1.0.0 // indirect
	go.opentelemetry.io/collector/processor v0.91.0 // indirect
	go.opentelemetry.io/collector/processor/batchprocessor v0.91.0 // indirect
	go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.91.0 // indirect
	go.opentelemetry.io/collector/receiver v0.91.0 // indirect
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.91.0 // indirect
	go.opentelemetry.io/collector/semconv v0.91.0 // indirect
	go.opentelemetry.io/collector/service v0.91.0 // indirect
	go.opentelemetry.io/contrib/config v0.1.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.46.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.46.1 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.21.1 // indirect
	go.opentelemetry.io/otel v1.21.0 // indirect
	go.opentelemetry.io/otel/bridge/opencensus v0.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v0.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.21.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.21.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.21.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.44.1-0.20231201153405-6027c1ae76f2 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.44.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.21.0 // indirect
	go.opentelemetry.io/otel/metric v1.21.0 // indirect
	go.opentelemetry.io/otel/sdk v1.21.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.21.0 // indirect
	go.opentelemetry.io/otel/trace v1.21.0 // indirect
=======
	go.opentelemetry.io/collector v0.88.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.88.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v0.88.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.88.0 // indirect
	go.opentelemetry.io/collector/config/confighttp v0.88.0 // indirect
	go.opentelemetry.io/collector/config/confignet v0.88.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v0.88.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.88.0 // indirect
	go.opentelemetry.io/collector/config/configtls v0.88.0 // indirect
	go.opentelemetry.io/collector/config/internal v0.88.0 // indirect
	go.opentelemetry.io/collector/connector v0.88.0 // indirect
	go.opentelemetry.io/collector/consumer v0.88.0 // indirect
	go.opentelemetry.io/collector/exporter v0.88.0 // indirect
	go.opentelemetry.io/collector/exporter/loggingexporter v0.88.0 // indirect
	go.opentelemetry.io/collector/exporter/otlpexporter v0.88.0 // indirect
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.88.0 // indirect
	go.opentelemetry.io/collector/extension v0.88.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.88.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.0.0-rcv0017 // indirect
	go.opentelemetry.io/collector/pdata v1.0.0-rcv0017 // indirect
	go.opentelemetry.io/collector/processor v0.88.0 // indirect
	go.opentelemetry.io/collector/processor/batchprocessor v0.88.0 // indirect
	go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.88.0 // indirect
	go.opentelemetry.io/collector/receiver v0.88.0 // indirect
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.88.0 // indirect
	go.opentelemetry.io/collector/semconv v0.88.0 // indirect
	go.opentelemetry.io/collector/service v0.88.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.45.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.45.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.20.0 // indirect
	go.opentelemetry.io/otel v1.19.0 // indirect
	go.opentelemetry.io/otel/bridge/opencensus v0.42.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.42.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.42.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v0.42.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.19.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.19.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.19.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.42.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.42.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.19.0 // indirect
	go.opentelemetry.io/otel/metric v1.19.0 // indirect
	go.opentelemetry.io/otel/sdk v1.19.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.19.0 // indirect
	go.opentelemetry.io/otel/trace v1.19.0 // indirect
>>>>>>> b497143 (Revert)
	go.opentelemetry.io/proto/otlp v1.0.0 // indirect
<<<<<<< HEAD
	golang.org/x/exp v0.0.0-20231127185646-65229373498e // indirect
=======
	golang.org/x/exp v0.0.0-20230713183714-613f0c0eb8a1 // indirect
<<<<<<< HEAD
<<<<<<< HEAD
>>>>>>> 3776821 (Revert)
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
=======
=======
	golang.org/x/mod v0.13.0 // indirect
>>>>>>> d53b4b7 (Updated)
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/oauth2 v0.13.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/term v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
<<<<<<< HEAD
>>>>>>> b497143 (Revert)
	gonum.org/v1/gonum v0.14.0 // indirect
<<<<<<< HEAD
	google.golang.org/genproto/googleapis/api v0.0.0-20231106174013-bbf56f31fb17 // indirect
=======
=======
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.14.0 // indirect
	gonum.org/v1/gonum v0.14.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
>>>>>>> d53b4b7 (Updated)
	google.golang.org/genproto/googleapis/api v0.0.0-20230822172742-b8732ec3820d // indirect
<<<<<<< HEAD
>>>>>>> 3776821 (Revert)
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231127180814-3a041ad873d4 // indirect
=======
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231012201019-e917dd12ba7a // indirect
>>>>>>> b497143 (Revert)
	google.golang.org/grpc v1.59.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
<<<<<<< HEAD
<<<<<<< HEAD
	k8s.io/api v0.28.4 // indirect
	k8s.io/apimachinery v0.28.4 // indirect
	k8s.io/client-go v0.28.4 // indirect
=======
	k8s.io/api v0.28.2 // indirect
	k8s.io/apimachinery v0.28.2 // indirect
	k8s.io/client-go v0.28.2 // indirect
>>>>>>> d53b4b7 (Updated)
	k8s.io/klog/v2 v2.100.1 // indirect
	k8s.io/kube-openapi v0.0.0-20230717233707-2695361300d9 // indirect
	k8s.io/utils v0.0.0-20230711102312-30195339c3c7 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.3.0 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
<<<<<<< HEAD
=======
>>>>>>> 3776821 (Revert)
=======
>>>>>>> d53b4b7 (Updated)
)
