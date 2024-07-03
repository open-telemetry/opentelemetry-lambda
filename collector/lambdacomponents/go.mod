module github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents

go 1.21.0

toolchain go1.22.2

require (
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusremotewriteexporter v0.102.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/sigv4authextension v0.102.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.102.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.102.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor v0.102.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.102.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.102.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor v0.102.0
	github.com/open-telemetry/opentelemetry-lambda/collector/extension/solarwindsapmsettingsextension v0.0.0-00010101000000-000000000000
	github.com/open-telemetry/opentelemetry-lambda/collector/processor/coldstartprocessor v0.98.0
	github.com/open-telemetry/opentelemetry-lambda/collector/processor/decoupleprocessor v0.0.0-00010101000000-000000000000
	github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver v0.98.0
	go.opentelemetry.io/collector/exporter v0.102.1
	go.opentelemetry.io/collector/exporter/loggingexporter v0.102.1
	go.opentelemetry.io/collector/exporter/otlpexporter v0.102.1
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.102.1
	go.opentelemetry.io/collector/extension v0.102.1
	go.opentelemetry.io/collector/otelcol v0.102.1
	go.opentelemetry.io/collector/processor v0.102.1
	go.opentelemetry.io/collector/processor/batchprocessor v0.102.1
	go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.102.1
	go.opentelemetry.io/collector/receiver v0.102.1
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.102.1
	go.uber.org/multierr v1.11.0
)

require (
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.23.0 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/Showmax/go-fqdn v1.0.0 // indirect
	github.com/alecthomas/participle/v2 v2.1.1 // indirect
	github.com/aws/aws-sdk-go-v2 v1.27.2 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.27.18 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.18 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.20.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.24.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.28.12 // indirect
	github.com/aws/smithy-go v1.20.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/expr-lang/expr v1.16.9 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/jsonpointer v0.20.2 // indirect
	github.com/go-openapi/jsonreference v0.20.4 // indirect
	github.com/go-openapi/swag v0.22.9 // indirect
	github.com/go-viper/mapstructure/v2 v2.0.0-alpha.1 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-collections/go-datastructures v0.0.0-20150211160725-59788d5eb259 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/providers/confmap v0.1.0 // indirect
	github.com/knadh/koanf/v2 v2.1.1 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.1-0.20231216201459-8508981c8b6c // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.102.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter v0.102.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl v0.102.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatautil v0.102.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/resourcetotelemetry v0.102.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sampling v0.102.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus v0.102.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheusremotewrite v0.102.0 // indirect
	github.com/open-telemetry/opentelemetry-lambda/collector v0.98.0 // indirect
	github.com/open-telemetry/opentelemetry-lambda/collector/lambdalifecycle v0.0.0-00010101000000-000000000000 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/openshift/api v3.9.0+incompatible // indirect
	github.com/openshift/client-go v0.0.0-20210521082421-73d9475a9142 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/client_golang v1.19.1 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.53.0 // indirect
	github.com/prometheus/procfs v0.15.0 // indirect
	github.com/prometheus/prometheus v0.51.2-0.20240405174432-b4a973753c6e // indirect
	github.com/rs/cors v1.10.1 // indirect
	github.com/shirou/gopsutil/v3 v3.24.4 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/solarwindscloud/apm-proto v1.0.3 // indirect
	github.com/spf13/cobra v1.8.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	github.com/tidwall/gjson v1.10.2 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tidwall/tinylru v1.1.0 // indirect
	github.com/tidwall/wal v1.1.7 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/collector v0.102.1 // indirect
	go.opentelemetry.io/collector/component v0.102.1 // indirect
	go.opentelemetry.io/collector/config/configauth v0.102.1 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.9.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.102.1 // indirect
	go.opentelemetry.io/collector/config/confighttp v0.102.1 // indirect
	go.opentelemetry.io/collector/config/confignet v0.102.1 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.9.0 // indirect
	go.opentelemetry.io/collector/config/configretry v0.102.1 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.102.1 // indirect
	go.opentelemetry.io/collector/config/configtls v0.102.1 // indirect
	go.opentelemetry.io/collector/config/internal v0.102.1 // indirect
	go.opentelemetry.io/collector/confmap v0.102.1 // indirect
	go.opentelemetry.io/collector/confmap/converter/expandconverter v0.102.1 // indirect
	go.opentelemetry.io/collector/confmap/provider/envprovider v0.102.1 // indirect
	go.opentelemetry.io/collector/confmap/provider/fileprovider v0.102.1 // indirect
	go.opentelemetry.io/collector/confmap/provider/httpprovider v0.102.1 // indirect
	go.opentelemetry.io/collector/confmap/provider/httpsprovider v0.102.1 // indirect
	go.opentelemetry.io/collector/confmap/provider/yamlprovider v0.102.1 // indirect
	go.opentelemetry.io/collector/connector v0.102.1 // indirect
	go.opentelemetry.io/collector/consumer v0.102.1 // indirect
	go.opentelemetry.io/collector/extension/auth v0.102.1 // indirect
	go.opentelemetry.io/collector/featuregate v1.9.0 // indirect
	go.opentelemetry.io/collector/pdata v1.9.0 // indirect
	go.opentelemetry.io/collector/semconv v0.102.1 // indirect
	go.opentelemetry.io/collector/service v0.102.1 // indirect
	go.opentelemetry.io/contrib/config v0.7.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.52.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.52.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.27.0 // indirect
	go.opentelemetry.io/otel v1.27.0 // indirect
	go.opentelemetry.io/otel/bridge/opencensus v1.27.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.27.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.27.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.27.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.27.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.27.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.49.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.27.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.27.0 // indirect
	go.opentelemetry.io/otel/metric v1.27.0 // indirect
	go.opentelemetry.io/otel/sdk v1.27.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.27.0 // indirect
	go.opentelemetry.io/otel/trace v1.27.0 // indirect
	go.opentelemetry.io/proto/otlp v1.2.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	gonum.org/v1/gonum v0.15.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240520151616-dc85e6b867a5 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240520151616-dc85e6b867a5 // indirect
	google.golang.org/grpc v1.64.0 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.29.3 // indirect
	k8s.io/apimachinery v0.29.3 // indirect
	k8s.io/client-go v0.29.3 // indirect
	k8s.io/klog/v2 v2.120.1 // indirect
	k8s.io/kube-openapi v0.0.0-20231010175941-2dd684a91f00 // indirect
	k8s.io/utils v0.0.0-20230726121419-3b25d923346b // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

// ambiguous import: found package cloud.google.com/go/compute/metadata in multiple modules:
//        cloud.google.com/go
//        cloud.google.com/go/compute
// Force cloud.google.com/go to be at least v0.107.0, so that the metadata is not present.
replace cloud.google.com/go => cloud.google.com/go v0.107.0

replace github.com/open-telemetry/opentelemetry-lambda/collector => ../

replace github.com/open-telemetry/opentelemetry-lambda/collector/lambdalifecycle => ../lambdalifecycle

replace github.com/open-telemetry/opentelemetry-lambda/collector/processor/coldstartprocessor => ../processor/coldstartprocessor

replace github.com/open-telemetry/opentelemetry-lambda/collector/processor/decoupleprocessor => ../processor/decoupleprocessor

replace github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver => ../receiver/telemetryapireceiver

replace github.com/open-telemetry/opentelemetry-lambda/collector/extension/solarwindsapmsettingsextension => ../extension/solarwindsapmsettingsextension
