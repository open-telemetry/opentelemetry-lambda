module github.com/open-telemetry/opentelemetry-lambda/collector

go 1.16

replace github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents => ./lambdacomponents

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-kit/log v0.2.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents v0.0.0
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/prometheus/statsd_exporter v0.22.3 // indirect
	go.opentelemetry.io/collector v0.38.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.26.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.26.0 // indirect
	go.opentelemetry.io/contrib/zpages v0.26.0 // indirect
	golang.org/x/net v0.0.0-20211029224645-99673261e6eb // indirect
	golang.org/x/sys v0.0.0-20211029165221-6e7872819dc8 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20211029142109-e255c875f7c7 // indirect
)
