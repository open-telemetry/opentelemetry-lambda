module github.com/open-telemetry/opentelemetry-lambda/collector

go 1.26.9

toolchain go1.26.11

replace github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents => ./lambdacomponents

replace github.com/open-telemetry/opentelemetry-lambda/collector/lambdalifecycle => ./lambdalifecycle

replace github.com/open-telemetry/opentelemetry-lambda/collector/processor/coldstartprocessor => ./processor/coldstartprocessor

replace github.com/open-telemetry/opentelemetry-lambda/collector/processor/decoupleprocessor => ./processor/decoupleprocessor

replace github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver => ./receiver/telemetryapireceiver

// fixes ambiguous import error: found package cloud.google.com/go/compute/metadata in multiple modules:
//        cloud.google.com/go
//        cloud.google.com/go/compute
// Force cloud.google.com/go to be at least v0.107.0, so that the metadata is not present.
replace cloud.google.com/go => cloud.google.com/go v0.123.0

require (
	github.com/golang-collections/go-datastructures v0.0.0-20150211160725-59788d5eb259
	github.com/google/go-cmp v0.7.0
	github.com/open-telemetry/opentelemetry-collector-contrib/confmap/provider/s3provider v0.165.1
	github.com/open-telemetry/opentelemetry-collector-contrib/confmap/provider/secretsmanagerprovider v0.165.1
	github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents v0.98.0
	github.com/open-telemetry/opentelemetry-lambda/collector/lambdalifecycle v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.12.1
	go.opentelemetry.io/collector/component v1.71.0
	go.opentelemetry.io/collector/confmap v1.71.0
	go.opentelemetry.io/collector/confmap/provider/envprovider v1.71.0
	go.opentelemetry.io/collector/confmap/provider/fileprovider v1.71.0
	go.opentelemetry.io/collector/confmap/provider/httpprovider v1.71.0
	go.opentelemetry.io/collector/confmap/provider/httpsprovider v1.71.0
	go.opentelemetry.io/collector/confmap/provider/yamlprovider v1.71.0
	go.opentelemetry.io/collector/exporter/exportertest v0.165.0
	go.opentelemetry.io/collector/otelcol v0.165.0
	go.opentelemetry.io/collector/receiver/receivertest v0.165.0
	go.opentelemetry.io/collector/service v0.165.0
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.28.0
)

require (
	cloud.google.com/go/auth v0.24.1 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.3.0 // indirect
	cloud.google.com/go/compute v1.74.1 // indirect
	cloud.google.com/go/compute/metadata v0.10.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.23.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.14.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.12.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.9.0 // indirect
	github.com/GehirnInc/crypt v0.0.0-20230320061759-8cc1b52080c5 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.38.0 // indirect
	github.com/Masterminds/semver/v3 v3.5.0 // indirect
	github.com/Microsoft/go-winio v0.6.4 // indirect
	github.com/Showmax/go-fqdn v1.0.0 // indirect
	github.com/alecthomas/participle/v2 v2.1.4 // indirect
	github.com/alecthomas/units v0.0.0-20240927000941-0f3dac36c52b // indirect
	github.com/antchfx/xmlquery v1.5.2 // indirect
	github.com/antchfx/xpath v1.3.9 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/aws/aws-sdk-go-v2 v1.47.5 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.25 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.33.13 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.20.13 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.20.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.5.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.8.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.5.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.343.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.11.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.14.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.20.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.114.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.50.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.12.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.39.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.44.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.51.6 // indirect
	github.com/aws/smithy-go v1.28.7 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/brunoscheufler/aws-ecs-metadata-go v0.0.0-20221221133751-67e37ae746cd // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cenkalti/backoff/v7 v7.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dennwc/varint v1.0.0 // indirect
	github.com/digitalocean/go-metadata v0.0.0-20250129100319-e3650a3df44b // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-connections v0.8.2 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/ebitengine/purego v0.11.2 // indirect
	github.com/elastic/go-grok v0.3.1 // indirect
	github.com/elastic/lunes v0.2.2 // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/expr-lang/expr v1.17.8 // indirect
	github.com/fatih/color v1.19.0 // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/foxboron/go-tpm-keyfiles v0.0.0-20260902202739-8c9c2d1005f4 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/fxamacker/cbor/v2 v2.9.4 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v1.0.2 // indirect
	github.com/go-openapi/jsonreference v1.0.3 // indirect
	github.com/go-openapi/swag v0.29.2 // indirect
	github.com/go-openapi/swag/cmdutils v0.29.2 // indirect
	github.com/go-openapi/swag/conv v0.29.2 // indirect
	github.com/go-openapi/swag/fileutils v0.29.2 // indirect
	github.com/go-openapi/swag/jsonutils v0.29.2 // indirect
	github.com/go-openapi/swag/loading v0.29.2 // indirect
	github.com/go-openapi/swag/mangling v0.29.2 // indirect
	github.com/go-openapi/swag/netutils v0.29.2 // indirect
	github.com/go-openapi/swag/pools v0.29.2 // indirect
	github.com/go-openapi/swag/stringutils v0.29.2 // indirect
	github.com/go-openapi/swag/typeutils v0.29.2 // indirect
	github.com/go-openapi/swag/yamlutils v0.29.2 // indirect
	github.com/go-resty/resty/v2 v2.17.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.6 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/gnostic-models v0.7.1 // indirect
	github.com/google/go-tpm v0.9.9-0.20261001180303-e05fb8b1ca4e // indirect
	github.com/google/s2a-go v0.1.11 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.23 // indirect
	github.com/googleapis/gax-go/v2 v2.27.0 // indirect
	github.com/grafana/regexp v0.0.0-20250905093917-f7b3be9d1853 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.32.0 // indirect
	github.com/hashicorp/consul/api v1.34.4 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-metrics v0.6.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-version v1.9.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/serf v0.10.4 // indirect
	github.com/hetznercloud/hcloud-go/v2 v2.49.0 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.20.3 // indirect
	github.com/klauspost/cpuid/v2 v2.4.0 // indirect
	github.com/knadh/koanf/maps v0.1.3 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.1 // indirect
	github.com/knadh/koanf/v2 v2.4.1 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lightstep/go-expohisto v1.0.0 // indirect
	github.com/linode/go-metadata v0.3.1 // indirect
	github.com/lufia/plan9stats v0.0.0-20260802145828-341c2f0c90b5 // indirect
	github.com/magefile/mage v1.17.2 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.23 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/moby/api v1.56.1 // indirect
	github.com/moby/moby/client v0.6.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/spanmetricsconnector v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusremotewriteexporter v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/basicauthextension v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/internal/basicauth v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/internal/credentialsfile v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/internal/secretprovider v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/sigv4authextension v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/ecsutil v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/metadataproviders v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/pdatautil v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/expohisto v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl v1.4.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/xprofile v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/xottl v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatautil v0.161.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/resourcetotelemetry v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sampling v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheusremotewrite v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor v0.165.1 // indirect
	github.com/open-telemetry/opentelemetry-lambda/collector/processor/coldstartprocessor v0.98.0 // indirect
	github.com/open-telemetry/opentelemetry-lambda/collector/processor/decoupleprocessor v0.0.0-00010101000000-000000000000 // indirect
	github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver v0.98.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/openshift/api v0.0.0-20251015095338-264e80a2b6e7 // indirect
	github.com/openshift/client-go v0.0.0-20251015124057-db0dee36e235 // indirect
	github.com/pierrec/lz4/v4 v4.1.33 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20260916203055-22a1a467d9f0 // indirect
	github.com/prometheus/client_golang v1.25.0 // indirect
	github.com/prometheus/client_golang/exp v0.0.0-20261022091949-f3856e0d35a3 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.72.1 // indirect
	github.com/prometheus/otlptranslator v1.0.0 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	github.com/prometheus/prometheus v0.315.1 // indirect
	github.com/prometheus/sigv4 v0.5.0 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/scaleway/scaleway-sdk-go v1.39.0 // indirect
	github.com/shirou/gopsutil/v4 v4.26.10 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/tg123/go-htpasswd v1.2.5 // indirect
	github.com/tidwall/gjson v1.19.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/tinylru v1.2.1 // indirect
	github.com/tidwall/wal v1.2.1 // indirect
	github.com/tklauser/go-sysconf v0.4.0 // indirect
	github.com/tklauser/numcpus v0.12.0 // indirect
	github.com/twmb/murmur3 v1.2.2 // indirect
	github.com/ua-parser/uap-go v0.0.0-20260529044130-17c35e68e58c // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/xxh3 v1.1.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/collector/client v1.71.0 // indirect
	go.opentelemetry.io/collector/component/componentstatus v0.165.0 // indirect
	go.opentelemetry.io/collector/component/componenttest v0.165.0 // indirect
	go.opentelemetry.io/collector/config/configauth v1.71.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.71.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v1.71.0 // indirect
	go.opentelemetry.io/collector/config/confighttp v1.71.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v1.71.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.71.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.71.0 // indirect
	go.opentelemetry.io/collector/config/configoptional v1.71.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.71.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.165.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.71.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.165.0 // indirect
	go.opentelemetry.io/collector/connector v0.165.0 // indirect
	go.opentelemetry.io/collector/connector/connectortest v0.165.0 // indirect
	go.opentelemetry.io/collector/connector/xconnector v0.165.0 // indirect
	go.opentelemetry.io/collector/consumer v1.71.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.165.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror/xconsumererror v0.165.0 // indirect
	go.opentelemetry.io/collector/consumer/consumertest v0.165.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.165.0 // indirect
	go.opentelemetry.io/collector/exporter v1.71.0 // indirect
	go.opentelemetry.io/collector/exporter/debugexporter v0.165.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper v0.165.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper/xexporterhelper v0.165.0 // indirect
	go.opentelemetry.io/collector/exporter/otlpexporter v0.165.0 // indirect
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.165.0 // indirect
	go.opentelemetry.io/collector/exporter/xexporter v0.165.0 // indirect
	go.opentelemetry.io/collector/extension v1.71.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.71.0 // indirect
	go.opentelemetry.io/collector/extension/extensioncapabilities v0.165.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.165.0 // indirect
	go.opentelemetry.io/collector/extension/extensiontest v0.165.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.165.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.71.0 // indirect
	go.opentelemetry.io/collector/internal/componentalias v0.165.0 // indirect
	go.opentelemetry.io/collector/internal/fanoutconsumer v0.165.0 // indirect
	go.opentelemetry.io/collector/internal/memorylimiter v0.165.0 // indirect
	go.opentelemetry.io/collector/internal/sharedcomponent v0.165.0 // indirect
	go.opentelemetry.io/collector/internal/statusutil v0.165.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.165.0 // indirect
	go.opentelemetry.io/collector/pdata v1.71.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.165.0 // indirect
	go.opentelemetry.io/collector/pdata/testdata v0.165.0 // indirect
	go.opentelemetry.io/collector/pdata/xpdata v0.165.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.71.0 // indirect
	go.opentelemetry.io/collector/pipeline/xpipeline v0.165.0 // indirect
	go.opentelemetry.io/collector/processor v1.71.0 // indirect
	go.opentelemetry.io/collector/processor/batchprocessor v0.165.0 // indirect
	go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.165.0 // indirect
	go.opentelemetry.io/collector/processor/processorhelper v0.165.0 // indirect
	go.opentelemetry.io/collector/processor/processorhelper/xprocessorhelper v0.165.0 // indirect
	go.opentelemetry.io/collector/processor/processortest v0.165.0 // indirect
	go.opentelemetry.io/collector/processor/xprocessor v0.165.0 // indirect
	go.opentelemetry.io/collector/receiver v1.71.0 // indirect
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.165.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverhelper v0.165.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverhelper/xreceiverhelper v0.165.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.165.0 // indirect
	go.opentelemetry.io/collector/service/hostcapabilities v0.165.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.21.0 // indirect
	go.opentelemetry.io/contrib/detectors/aws/ec2/v2 v2.6.0 // indirect
	go.opentelemetry.io/contrib/detectors/aws/ecs v1.47.0 // indirect
	go.opentelemetry.io/contrib/detectors/aws/eks v1.47.0 // indirect
	go.opentelemetry.io/contrib/detectors/aws/elasticbeanstalk v0.19.0 // indirect
	go.opentelemetry.io/contrib/detectors/azure/azureappservice v0.19.0 // indirect
	go.opentelemetry.io/contrib/detectors/azure/azurecontainerapps v0.19.0 // indirect
	go.opentelemetry.io/contrib/detectors/azure/azurefunctions v0.19.0 // indirect
	go.opentelemetry.io/contrib/detectors/azure/azurevm v0.19.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.47.0 // indirect
	go.opentelemetry.io/contrib/detectors/hetzner v0.19.0 // indirect
	go.opentelemetry.io/contrib/detectors/ibmcloud/vpc v0.19.0 // indirect
	go.opentelemetry.io/contrib/detectors/kubeadm v0.19.0 // indirect
	go.opentelemetry.io/contrib/detectors/openshift v0.19.0 // indirect
	go.opentelemetry.io/contrib/detectors/vultr v0.19.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.72.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.72.0 // indirect
	go.opentelemetry.io/contrib/otelconf v0.27.0 // indirect
	go.opentelemetry.io/contrib/propagators/autoprop v0.72.0 // indirect
	go.opentelemetry.io/contrib/propagators/aws v1.47.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.47.0 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.47.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.47.0 // indirect
	go.opentelemetry.io/otel v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.23.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.23.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.69.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.23.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.47.0 // indirect
	go.opentelemetry.io/otel/log v1.47.0 // indirect
	go.opentelemetry.io/otel/metric v1.47.0 // indirect
	go.opentelemetry.io/otel/sdk v1.47.0 // indirect
	go.opentelemetry.io/otel/sdk/log v1.47.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.47.0 // indirect
	go.opentelemetry.io/otel/trace v1.47.0 // indirect
	go.opentelemetry.io/proto/otlp v1.11.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.6 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/exp v0.0.0-20261014143641-f0251eb4aa3e // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/oauth2 v0.37.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/term v0.47.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	golang.org/x/time v0.16.0 // indirect
	gonum.org/v1/gonum v0.17.0 // indirect
	google.golang.org/api v0.307.0 // indirect
	google.golang.org/genproto v0.0.0-20261012220534-4341edf3f941 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20261019180522-ce908d29299b // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261019180522-ce908d29299b // indirect
	google.golang.org/grpc v1.85.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.37.2 // indirect
	k8s.io/apimachinery v0.37.2 // indirect
	k8s.io/client-go v0.37.2 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	k8s.io/kube-openapi v0.0.0-20260821135717-be32def86098 // indirect
	k8s.io/utils v0.0.0-20260707023825-cf1189d6abe3 // indirect
	sigs.k8s.io/json v0.0.0-20260909141634-11ed52e25bc5 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.4.2 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)
