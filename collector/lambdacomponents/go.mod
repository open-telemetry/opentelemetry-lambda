module github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents

go 1.26.1

require (
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/spanmetricsconnector v0.160.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusremotewriteexporter v0.160.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/basicauthextension v0.160.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/sigv4authextension v0.160.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.160.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.160.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor v0.160.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.160.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor v0.160.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor v0.160.0
	github.com/open-telemetry/opentelemetry-lambda/collector/processor/coldstartprocessor v0.98.0
	github.com/open-telemetry/opentelemetry-lambda/collector/processor/decoupleprocessor v0.0.0-00010101000000-000000000000
	github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver v0.98.0
	go.opentelemetry.io/collector/component v1.66.0
	go.opentelemetry.io/collector/connector v0.160.0
	go.opentelemetry.io/collector/exporter v1.66.0
	go.opentelemetry.io/collector/exporter/debugexporter v0.160.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.160.0
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.160.0
	go.opentelemetry.io/collector/extension v1.66.0
	go.opentelemetry.io/collector/otelcol v0.160.0
	go.opentelemetry.io/collector/processor v1.66.0
	go.opentelemetry.io/collector/processor/batchprocessor v0.160.0
	go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.160.0
	go.opentelemetry.io/collector/receiver v1.66.0
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.160.0
	go.opentelemetry.io/collector/service v0.160.0
	go.uber.org/multierr v1.11.0
)

require (
	cloud.google.com/go/auth v0.23.1 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.23.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.14.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.12.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.8.0 // indirect
	github.com/GehirnInc/crypt v0.0.0-20230320061759-8cc1b52080c5 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.35.0 // indirect
	github.com/Masterminds/semver/v3 v3.5.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/alecthomas/participle/v2 v2.1.4 // indirect
	github.com/alecthomas/units v0.0.0-20240927000941-0f3dac36c52b // indirect
	github.com/antchfx/xmlquery v1.5.1 // indirect
	github.com/antchfx/xpath v1.3.8 // indirect
	github.com/aws/aws-sdk-go-v2 v1.45.1 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.33.1 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.20.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.19.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.5.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.5.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.14.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.35.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.40.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.47.1 // indirect
	github.com/aws/smithy-go v1.28.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/brunoscheufler/aws-ecs-metadata-go v0.0.0-20221221133751-67e37ae746cd // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cenkalti/backoff/v7 v7.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dennwc/varint v1.0.0 // indirect
	github.com/ebitengine/purego v0.10.2 // indirect
	github.com/elastic/go-grok v0.3.1 // indirect
	github.com/elastic/lunes v0.2.2 // indirect
	github.com/expr-lang/expr v1.17.8 // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/foxboron/go-tpm-keyfiles v0.0.0-20260427185012-515ba073c4c1 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/fxamacker/cbor/v2 v2.9.2 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.6 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-collections/go-datastructures v0.0.0-20150211160725-59788d5eb259 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/go-tpm v0.9.9-0.20260602212016-9f0977c7f65a // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.20 // indirect
	github.com/googleapis/gax-go/v2 v2.23.0 // indirect
	github.com/grafana/regexp v0.0.0-20250905093917-f7b3be9d1853 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.30.0 // indirect
	github.com/hashicorp/go-version v1.9.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.19.2 // indirect
	github.com/klauspost/cpuid/v2 v2.4.0 // indirect
	github.com/knadh/koanf/maps v0.1.3 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.1 // indirect
	github.com/knadh/koanf/v2 v2.3.6 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lightstep/go-expohisto v1.0.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20260802145828-341c2f0c90b5 // indirect
	github.com/magefile/mage v1.17.2 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/internal/basicauth v0.160.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/internal/credentialsfile v0.160.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.160.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter v0.160.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/pdatautil v0.160.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl v0.160.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatautil v0.160.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/resourcetotelemetry v0.160.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sampling v0.160.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus v0.160.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheusremotewrite v0.160.0 // indirect
	github.com/open-telemetry/opentelemetry-lambda/collector v0.98.0 // indirect
	github.com/open-telemetry/opentelemetry-lambda/collector/lambdalifecycle v0.0.0-00010101000000-000000000000 // indirect
	github.com/pierrec/lz4/v4 v4.1.29 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/power-devops/perfstat v0.0.0-20260805114148-88456608a4f6 // indirect
	github.com/prometheus/client_golang v1.24.1 // indirect
	github.com/prometheus/client_golang/exp v0.0.0-20260812154952-0c5dccd910c0 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.70.1 // indirect
	github.com/prometheus/otlptranslator v1.0.0 // indirect
	github.com/prometheus/procfs v0.21.1 // indirect
	github.com/prometheus/prometheus v0.314.0 // indirect
	github.com/prometheus/sigv4 v0.4.1 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/shirou/gopsutil/v4 v4.26.8 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/stretchr/testify v1.12.1 // indirect
	github.com/tg123/go-htpasswd v1.2.5 // indirect
	github.com/tidwall/gjson v1.19.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/tinylru v1.2.1 // indirect
	github.com/tidwall/wal v1.2.1 // indirect
	github.com/tklauser/go-sysconf v0.4.0 // indirect
	github.com/tklauser/numcpus v0.12.0 // indirect
	github.com/twmb/murmur3 v1.1.8 // indirect
	github.com/ua-parser/uap-go v0.0.0-20260529044130-17c35e68e58c // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/xxh3 v1.1.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/collector v0.160.0 // indirect
	go.opentelemetry.io/collector/client v1.66.0 // indirect
	go.opentelemetry.io/collector/component/componentstatus v0.160.0 // indirect
	go.opentelemetry.io/collector/component/componenttest v0.160.0 // indirect
	go.opentelemetry.io/collector/config/configauth v1.66.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.66.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v1.66.0 // indirect
	go.opentelemetry.io/collector/config/confighttp v0.160.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v1.66.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.66.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.66.0 // indirect
	go.opentelemetry.io/collector/config/configoptional v1.66.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.66.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.160.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.66.0 // indirect
	go.opentelemetry.io/collector/confmap v1.66.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.160.0 // indirect
	go.opentelemetry.io/collector/connector/connectortest v0.160.0 // indirect
	go.opentelemetry.io/collector/connector/xconnector v0.160.0 // indirect
	go.opentelemetry.io/collector/consumer v1.66.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.160.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror/xconsumererror v0.160.0 // indirect
	go.opentelemetry.io/collector/consumer/consumertest v0.160.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.160.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper v0.160.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper/xexporterhelper v0.160.0 // indirect
	go.opentelemetry.io/collector/exporter/exportertest v0.160.0 // indirect
	go.opentelemetry.io/collector/exporter/xexporter v0.160.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.66.0 // indirect
	go.opentelemetry.io/collector/extension/extensioncapabilities v0.160.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.160.0 // indirect
	go.opentelemetry.io/collector/extension/extensiontest v0.160.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.160.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.66.0 // indirect
	go.opentelemetry.io/collector/internal/componentalias v0.160.0 // indirect
	go.opentelemetry.io/collector/internal/fanoutconsumer v0.160.0 // indirect
	go.opentelemetry.io/collector/internal/memorylimiter v0.160.0 // indirect
	go.opentelemetry.io/collector/internal/sharedcomponent v0.160.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.160.0 // indirect
	go.opentelemetry.io/collector/pdata v1.66.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.160.0 // indirect
	go.opentelemetry.io/collector/pdata/testdata v0.160.0 // indirect
	go.opentelemetry.io/collector/pdata/xpdata v0.160.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.66.0 // indirect
	go.opentelemetry.io/collector/pipeline/xpipeline v0.160.0 // indirect
	go.opentelemetry.io/collector/processor/processorhelper v0.160.0 // indirect
	go.opentelemetry.io/collector/processor/processorhelper/xprocessorhelper v0.160.0 // indirect
	go.opentelemetry.io/collector/processor/processortest v0.160.0 // indirect
	go.opentelemetry.io/collector/processor/xprocessor v0.160.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverhelper v0.160.0 // indirect
	go.opentelemetry.io/collector/receiver/receivertest v0.160.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.160.0 // indirect
	go.opentelemetry.io/collector/service/hostcapabilities v0.160.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.20.0 // indirect
	go.opentelemetry.io/contrib/detectors/aws/ecs v1.45.0 // indirect
	go.opentelemetry.io/contrib/detectors/aws/eks v1.45.0 // indirect
	go.opentelemetry.io/contrib/detectors/azure/azurevm v0.17.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.45.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.70.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.70.0 // indirect
	go.opentelemetry.io/contrib/otelconf v0.25.0 // indirect
	go.opentelemetry.io/contrib/propagators/autoprop v0.70.0 // indirect
	go.opentelemetry.io/contrib/propagators/aws v1.45.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.45.0 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.45.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.45.0 // indirect
	go.opentelemetry.io/otel v1.46.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.21.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.21.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.45.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.45.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.45.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.45.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.45.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.67.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.21.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.45.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.45.0 // indirect
	go.opentelemetry.io/otel/log v0.22.0 // indirect
	go.opentelemetry.io/otel/metric v1.46.0 // indirect
	go.opentelemetry.io/otel/sdk v1.46.0 // indirect
	go.opentelemetry.io/otel/sdk/log v0.21.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.46.0 // indirect
	go.opentelemetry.io/otel/trace v1.46.0 // indirect
	go.opentelemetry.io/proto/otlp v1.11.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/zap v1.28.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.55.0 // indirect
	golang.org/x/exp v0.0.0-20260813180055-c1d0aacb2297 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/term v0.45.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	gonum.org/v1/gonum v0.17.0 // indirect
	google.golang.org/api v0.293.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260810153831-ec0a7760b754 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260810153831-ec0a7760b754 // indirect
	google.golang.org/grpc v1.83.2 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apimachinery v0.36.3 // indirect
	k8s.io/client-go v0.36.3 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	k8s.io/kube-openapi v0.0.0-20260721132016-d427ff9ee9ad // indirect
	k8s.io/utils v0.0.0-20260707023825-cf1189d6abe3 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.4.2 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)

// ambiguous import: found package cloud.google.com/go/compute/metadata in multiple modules:
//        cloud.google.com/go
//        cloud.google.com/go/compute
// Force cloud.google.com/go to be at least v0.107.0, so that the metadata is not present.
replace cloud.google.com/go => cloud.google.com/go v0.123.0

replace github.com/open-telemetry/opentelemetry-lambda/collector => ../

replace github.com/open-telemetry/opentelemetry-lambda/collector/lambdalifecycle => ../lambdalifecycle

replace github.com/open-telemetry/opentelemetry-lambda/collector/processor/coldstartprocessor => ../processor/coldstartprocessor

replace github.com/open-telemetry/opentelemetry-lambda/collector/processor/decoupleprocessor => ../processor/decoupleprocessor

replace github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver => ../receiver/telemetryapireceiver
