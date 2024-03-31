// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package telemetryapireceiver // import "github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver"

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/receiver"
	semconv "go.opentelemetry.io/collector/semconv/v1.22.0"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/open-telemetry/opentelemetry-lambda/collector/internal/telemetryapi"
)

const (
	defaultListenerPort  = "4325"
	instrumentationScope = "github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapi"
)

var (
	errUnknownMetric      = errors.New("metric is unknown")
	errMissingTimestamp   = errors.New("expected timestamp is missing")
	errMissingServiceName = errors.New("service name is missing")
)

/* ------------------------------------------ CREATION ----------------------------------------- */

type telemetryAPIReceiver struct {
	mutex      sync.Mutex
	didStartUp bool
	// SHARED
	httpServer  *http.Server
	logger      *zap.Logger
	extensionID string
	resource    pcommon.Resource
	cfg         *Config
	shutdown    chan struct{}
	// TRACES
	nextTracesConsumer  consumer.Traces
	lastRequestID       string
	lastTraceTimestamps traceTimestamps
	// METRICS
	// NOTE: We're using the OpenTelemetry SDK here as generating 'pmetric' structures entirely
	//  manually is error-prone and would duplicate plenty of code available in the SDK.
	nextMetricsConsumer   consumer.Metrics
	metricsReader         *sdkmetric.ManualReader
	metricInitDurations   metric.Float64Histogram
	metricInvokeDurations metric.Float64Histogram
	metricColdstarts      metric.Int64Counter
	metricSuccesses       metric.Int64Counter
	metricFailures        metric.Int64Counter
	metricTimeouts        metric.Int64Counter
	metricMemoryUsages    metric.Int64Histogram
	lastMetricTimestamps  metricTimestamps
}

type traceTimestamps struct {
	platformInitStartTime    *time.Time
	platformInitEndTime      *time.Time
	platformRuntimeStartTime *time.Time
	platformRuntimeEndTime   *time.Time
}

type metricTimestamps struct {
	platformInitReportTime *time.Time
	platformRuntimeEndTime *time.Time
	platformReportTime     *time.Time
}

func newTelemetryAPIReceiver(
	cfg *Config,
	set receiver.CreateSettings,
) (*telemetryAPIReceiver, error) {
	// Resource attributes follow the OTEL semantiv conventions...
	r := pcommon.NewResource()
	// Cloud Resource Attributes: https://opentelemetry.io/docs/specs/semconv/resource/cloud/
	r.Attributes().PutStr(semconv.AttributeCloudProvider, semconv.AttributeCloudProviderAWS)
	r.Attributes().PutStr(semconv.AttributeCloudPlatform, semconv.AttributeCloudPlatformAWSLambda)
	if val, ok := os.LookupEnv("AWS_REGION"); ok {
		r.Attributes().PutStr(semconv.AttributeCloudRegion, val)
	}
	// Service attributes: https://opentelemetry.io/docs/specs/semconv/resource/#service
	if val, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); ok {
		r.Attributes().PutStr(semconv.AttributeServiceName, val)
		r.Attributes().PutStr(semconv.AttributeFaaSName, val)
	} else {
		r.Attributes().PutStr(semconv.AttributeServiceName, "unknown_service")
	}
	// In order for metrics to adhere to the single-writer principle, service.instance.id MUST be set:
	// https://github.com/open-telemetry/opentelemetry-specification/blob/v1.6.1/specification/metrics/datamodel.md#single-writer
	r.Attributes().PutStr(semconv.AttributeServiceInstanceID, uuid.New().String())
	// FaaS Resource Attributes: https://opentelemetry.io/docs/specs/semconv/resource/faas/
	if val, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_VERSION"); ok {
		r.Attributes().PutStr(semconv.AttributeFaaSVersion, val)
	}
	if val, ok := os.LookupEnv("AWS_LAMBDA_LOG_STREAM_NAME"); ok {
		r.Attributes().PutStr(semconv.AttributeFaaSInstance, val)
	}
	if val, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_MEMORY_SIZE"); ok {
		if mb, err := strconv.Atoi(val); err == nil {
			r.Attributes().PutInt(semconv.AttributeFaaSMaxMemory, int64(mb)*1024*1024)
		}
	}

	// This telemetry API receiver is very minimal. We're lazily initializing most members when
	// this receiver is requested in processing pipelines.
	return &telemetryAPIReceiver{
		logger:      set.Logger,
		extensionID: cfg.extensionID,
		resource:    r,
		cfg:         cfg,
		shutdown:    make(chan struct{}),
	}, nil
}

func (r *telemetryAPIReceiver) setTracesConsumer(next consumer.Traces) {
	r.nextTracesConsumer = next
}

func (r *telemetryAPIReceiver) setMetricsConsumer(next consumer.Metrics) error {
	r.nextMetricsConsumer = next
	r.metricsReader = sdkmetric.NewManualReader()

	// Configure histogram aggregation based on configuration
	var aggregation sdkmetric.Aggregation
	if r.cfg.Metrics.UseExponentialHistograms {
		aggregation = sdkmetric.AggregationBase2ExponentialHistogram{
			MaxSize:  160,
			MaxScale: 20,
		}
	} else {
		aggregation = sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: []float64{
				0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10,
			},
		}
	}
	view := sdkmetric.NewView(
		sdkmetric.Instrument{Kind: sdkmetric.InstrumentKindHistogram},
		sdkmetric.Stream{Aggregation: aggregation},
	)

	// Initialize a meter for all metrics
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(r.metricsReader),
		sdkmetric.WithView(view),
	)
	meter := provider.Meter(instrumentationScope)

	// Build the metrics and propagate the last error.
	// NOTE: The metrics defined here follow the semantic conventions for FaaS Metrics:
	//       https://opentelemetry.io/docs/specs/semconv/faas/faas-metrics/
	var err error

	// COUNTERS
	r.metricColdstarts, err = meter.Int64Counter(
		"faas.coldstarts",
		metric.WithDescription("Number of invocation cold starts."),
		metric.WithUnit("1"),
	)
	r.metricSuccesses, err = meter.Int64Counter(
		"faas.invocations",
		metric.WithDescription("Number of successful invocations."),
		metric.WithUnit("1"),
	)
	r.metricFailures, err = meter.Int64Counter(
		"faas.errors",
		metric.WithDescription("Number of invocation errors."),
		metric.WithUnit("1"),
	)
	r.metricTimeouts, err = meter.Int64Counter(
		"faas.timeouts",
		metric.WithDescription("Number of invocation timeouts."),
		metric.WithUnit("1"),
	)

	// For all counters, we push a value of zero to properly indicate the start of the
	// counter. This is particularly important if the Lambda function is called rarely.
	r.metricColdstarts.Add(context.Background(), 0)
	r.metricSuccesses.Add(context.Background(), 0)
	r.metricFailures.Add(context.Background(), 0)
	r.metricTimeouts.Add(context.Background(), 0)

	// HISTOGRAMS
	r.metricInvokeDurations, err = meter.Float64Histogram(
		"faas.invoke_duration",
		metric.WithDescription("The duration of the function's logic execution."),
		metric.WithUnit("s"),
	)
	r.metricInitDurations, err = meter.Float64Histogram(
		"faas.init_duration",
		metric.WithDescription("The duration of the function's initialization."),
		metric.WithUnit("s"),
	)
	r.metricMemoryUsages, err = meter.Int64Histogram(
		"faas.mem_usage",
		metric.WithDescription("Max memory usage per invocation."),
		metric.WithUnit("By"),
	)
	return err
}

/* ------------------------------------ COMPONENT INTERFACE ------------------------------------ */

func (r *telemetryAPIReceiver) Start(ctx context.Context, host component.Host) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.didStartUp {
		return nil
	}

	address := listenOnAddress()
	r.logger.Info("Listening for requests", zap.String("address", address))

	mux := http.NewServeMux()
	mux.HandleFunc("/", r.httpHandler)
	r.httpServer = &http.Server{Addr: address, Handler: mux}
	go func() {
		_ = r.httpServer.ListenAndServe()
	}()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		_ = <-c
		_ = r.httpServer.Shutdown(context.Background())
		close(r.shutdown)
	}()

	telemetryClient := telemetryapi.NewClient(r.logger)
	_, err := telemetryClient.Subscribe(ctx, r.extensionID, fmt.Sprintf("http://%s/", address))
	if err != nil {
		r.logger.Info(
			"Listening for requests",
			zap.String("address", address), zap.String("extensionID", r.extensionID),
		)
		return err
	}

	r.didStartUp = true
	return nil
}

func (r *telemetryAPIReceiver) Shutdown(ctx context.Context) error {
	r.httpServer.Shutdown(ctx)
	select {
	case <-ctx.Done():
		return nil
	case <-r.shutdown:
		return nil
	}
}

/* --------------------------------------- EVENT HANDLER --------------------------------------- */

// httpHandler handles the requests coming from the Telemetry API.
// Logging or printing besides the error cases below is not recommended if you have subscribed to
// receive extension logs. Otherwise, logging here will cause Telemetry API to send new logs for
// the printed lines which may create an infinite loop.
func (r *telemetryAPIReceiver) httpHandler(w http.ResponseWriter, req *http.Request) {
	// We should not run HTTP handlers in parallel, this would cause all kinds of issues. Let's
	// just lock very coarsely here for simplicity. The TelemetryAPI should not send concurrent
	// requests anyway.
	r.mutex.Lock()
	defer r.mutex.Unlock()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.logger.Error("error reading body", zap.Error(err))
		return
	}

	var slice []event
	if err := json.Unmarshal(body, &slice); err != nil {
		r.logger.Error("error unmarshalling body", zap.Error(err))
		return
	}

	ctx := context.Background()
	for _, el := range slice {
		switch el.Type {
		// Function initialization started.
		case string(telemetryapi.PlatformInitStart):
			r.logger.Debug(fmt.Sprintf("Init start: %s", el.Time), zap.Any("event", el))
			time, err := parseTime(el.Time)
			if err != nil {
				r.logger.Error("unable to set last platform init start time", zap.Error(err))
				r.lastTraceTimestamps.platformInitStartTime = nil
			} else {
				r.lastTraceTimestamps.platformInitStartTime = &time
			}
		// Function initialization completed.
		case string(telemetryapi.PlatformInitRuntimeDone):
			r.logger.Debug(fmt.Sprintf("Init end: %s", el.Time), zap.Any("event", el))
			time, err := parseTime(el.Time)
			if err != nil {
				r.logger.Error("unable to set last platform init end time", zap.Error(err))
				r.lastTraceTimestamps.platformInitEndTime = nil
			} else {
				r.lastTraceTimestamps.platformInitEndTime = &time
			}
		// Concluding report on function initialization.
		case string(telemetryapi.PlatformInitReport):
			r.logger.Debug(fmt.Sprintf("Init report: %s", el.Time), zap.Any("event", el))
			time, err := parseTime(el.Time)
			if err != nil {
				r.logger.Error("unable to set last platform init report time", zap.Error(err))
				r.lastMetricTimestamps.platformInitReportTime = nil
			} else {
				r.lastMetricTimestamps.platformInitReportTime = &time
			}
			if r.metricsReader == nil {
				continue
			}
			if record, err := parseRecord[platformInitReportRecord](el, r.logger); err == nil {
				r.metricColdstarts.Add(ctx, 1)
				r.metricInitDurations.Record(ctx, record.Metrics.DurationMs/1000.0)
			}
		// Function invocation started.
		case string(telemetryapi.PlatformStart):
			r.logger.Debug(fmt.Sprintf("Invoke start: %s", el.Time), zap.Any("event", el))
			time, err := parseTime(el.Time)
			if err != nil {
				r.logger.Error("unable to set last platform runtime start time", zap.Error(err))
				r.lastTraceTimestamps.platformRuntimeStartTime = nil
			} else {
				r.lastTraceTimestamps.platformRuntimeStartTime = &time
			}
		// Function invocation completed.
		case string(telemetryapi.PlatformRuntimeDone):
			r.logger.Debug(fmt.Sprintf("Invoke end: %s", el.Time), zap.Any("event", el))
			time, err := parseTime(el.Time)
			if err != nil {
				r.logger.Error("unable to set last platform runtime end time", zap.Error(err))
				r.lastTraceTimestamps.platformRuntimeEndTime = nil
				r.lastMetricTimestamps.platformRuntimeEndTime = nil
			} else {
				r.lastTraceTimestamps.platformRuntimeEndTime = &time
				r.lastMetricTimestamps.platformRuntimeEndTime = &time
			}
			if record, err := parseRecord[platformRuntimeDoneRecord](el, r.logger); err == nil {
				r.lastRequestID = record.RequestID
				if r.metricsReader == nil {
					continue
				}
				r.metricInvokeDurations.Record(ctx, record.Metrics.DurationMs/1000.0)
				switch record.Status {
				case statusSuccess:
					r.metricSuccesses.Add(ctx, 1)
				case statusError, statusFailure:
					r.metricFailures.Add(ctx, 1)
				case statusTimeout:
					r.metricTimeouts.Add(ctx, 1)
				}
			}
		// Concluding report on function invocation (after runtime freeze).
		case string(telemetryapi.PlatformReport):
			r.logger.Debug(fmt.Sprintf("Invoke report: %s", el.Time), zap.Any("event", el))
			time, err := parseTime(el.Time)
			if err != nil {
				r.logger.Error("unable to set last platform report time", zap.Error(err))
				r.lastMetricTimestamps.platformReportTime = nil
			} else {
				r.lastMetricTimestamps.platformReportTime = &time
			}
			if r.metricsReader == nil {
				continue
			}
			if record, err := parseRecord[platformReport](el, r.logger); err == nil {
				r.metricMemoryUsages.Record(ctx, record.Metrics.MaxMemoryUsedMb*1024*1024)
			}
		}
		// TODO: potentially add support for additional events, see https://docs.aws.amazon.com/lambda/latest/dg/telemetry-api.html
		// Runtime restore started (reserved for future use)
		// case "platform.restoreStart":
		// Runtime restore completed (reserved for future use)
		// case "platform.restoreRuntimeDone":
		// Report of runtime restore (reserved for future use)
		// case "platform.restoreReport":
		// The extension subscribed to the Telemetry API.
		// case "platform.telemetrySubscription":
		// Lambda dropped log entries.
		// case "platform.logsDropped":
	}
	// NOTE: Forward metrics first as trace forwarding clears timestamps
	r.forwardMetrics()
	r.forwardTraces()
	slice = nil
}

func parseRecord[T any](el event, logger *zap.Logger) (T, error) {
	var record T
	if err := mapstructure.Decode(el.Record, &record); err != nil {
		logger.Error(
			fmt.Sprintf("Failed to parse %s record", el.Type),
			zap.Error(err), zap.Any("event", el),
		)
		return record, err
	}
	return record, nil
}

/* ----------------------------------------- FORWARDING ---------------------------------------- */

func (r *telemetryAPIReceiver) forwardTraces() {
	if r.lastTraceTimestamps.platformRuntimeStartTime != nil && r.lastTraceTimestamps.platformRuntimeEndTime != nil {
		if td, err := r.createPlatformRuntimeSpan(); err == nil {
			err := r.nextTracesConsumer.ConsumeTraces(context.Background(), td)
			if err == nil {
				// Clear for next invocation
				r.lastTraceTimestamps.platformInitStartTime = nil
				r.lastTraceTimestamps.platformInitEndTime = nil
				r.lastTraceTimestamps.platformRuntimeStartTime = nil
				r.lastTraceTimestamps.platformRuntimeEndTime = nil
			} else {
				r.logger.Error("error receiving traces", zap.Error(err))
			}
		}
	}
}

func (r *telemetryAPIReceiver) forwardMetrics() {
	if r.metricsReader == nil {
		// If the metrics reader is not set, no metrics consumer is set, we can stop.
		return
	}

	// Collect metrics from the metrics reader
	var resourceMetrics metricdata.ResourceMetrics
	if err := r.metricsReader.Collect(context.Background(), &resourceMetrics); err != nil {
		r.logger.Error("error collecting metrics", zap.Error(err))
		return
	}
	if len(resourceMetrics.ScopeMetrics) == 0 {
		// If there are no scope metrics, we do not need to export anything
		return
	}

	// Initialize internal metrics representation
	metricData := pmetric.NewMetrics()
	resourceMetricData := metricData.ResourceMetrics().AppendEmpty()
	r.resource.CopyTo(resourceMetricData.Resource())

	// Parse metrics from metrics reader into internal representation
	for _, scope := range resourceMetrics.ScopeMetrics {
		scopeMetrics := resourceMetricData.ScopeMetrics().AppendEmpty()
		scopeMetrics.Scope().SetName(scope.Scope.Name)
		for _, metric := range scope.Metrics {
			ts, err := r.getMetricTimestamp(metric.Name)
			if err != nil {
				r.logger.Error(
					fmt.Sprintf("failed to obtain last timestamp for metric '%s'", metric.Name),
					zap.Error(err),
				)
				continue
			}
			innerMetric := scopeMetrics.Metrics().AppendEmpty()
			if err := transformMetric(metric, innerMetric, ts); err != nil {
				r.logger.Error("error parsing collected metrics", zap.Error(err))
				return
			}
		}
	}

	// Eventually, forward the metrics to the consumer
	if err := r.nextMetricsConsumer.ConsumeMetrics(context.Background(), metricData); err != nil {
		r.logger.Error("error receiving metrics", zap.Error(err))
	}
}

/* ------------------------------------------- TRACES ------------------------------------------ */

func (r *telemetryAPIReceiver) createPlatformRuntimeSpan() (ptrace.Traces, error) {
	serviceName, ok := r.resource.Attributes().Get(semconv.AttributeServiceName)
	if !ok {
		return ptrace.Traces{}, errMissingServiceName
	}

	buildInitSpan := r.lastTraceTimestamps.platformInitStartTime != nil && r.lastTraceTimestamps.platformInitEndTime != nil

	// Build trace data
	traceData := ptrace.NewTraces()
	rs := traceData.ResourceSpans().AppendEmpty()
	r.resource.CopyTo(rs.Resource())

	ss := rs.ScopeSpans().AppendEmpty()
	ss.Scope().SetName(instrumentationScope)

	// Create root span for the entire runtime invocation
	traceID := newTraceID()
	rootSpan := ss.Spans().AppendEmpty()
	rootSpan.SetTraceID(traceID)
	rootSpan.SetSpanID(newSpanID())
	rootSpan.SetName(serviceName.Str())
	rootSpan.SetKind(ptrace.SpanKindServer)
	rootSpan.Attributes().PutBool(semconv.AttributeFaaSColdstart, buildInitSpan)
	rootSpan.Attributes().PutStr(semconv.AttributeFaaSInvocationID, r.lastRequestID)
	rootStartTime := *r.lastTraceTimestamps.platformRuntimeStartTime
	if r.lastTraceTimestamps.platformInitStartTime != nil {
		rootStartTime = *r.lastTraceTimestamps.platformInitStartTime
	}
	rootSpan.SetStartTimestamp(pcommon.NewTimestampFromTime(rootStartTime))
	rootSpan.SetEndTimestamp(pcommon.NewTimestampFromTime(*r.lastMetricTimestamps.platformRuntimeEndTime))

	// Optionally create an additional span for the init span
	if buildInitSpan {
		initSpan := ss.Spans().AppendEmpty()
		initSpan.SetTraceID(traceID)
		initSpan.SetSpanID(newSpanID())
		initSpan.SetParentSpanID(rootSpan.SpanID())
		initSpan.SetName("faas.runtimeInit")
		initSpan.SetKind(ptrace.SpanKindInternal)
		initSpan.Attributes().PutStr(semconv.AttributeFaaSInvocationID, r.lastRequestID)
		initSpan.SetStartTimestamp(pcommon.NewTimestampFromTime(*r.lastTraceTimestamps.platformInitStartTime))
		initSpan.SetEndTimestamp(pcommon.NewTimestampFromTime(*r.lastTraceTimestamps.platformInitEndTime))
	}
	return traceData, nil
}

func newSpanID() pcommon.SpanID {
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	randSource := rand.New(rand.NewSource(rngSeed))
	sid := pcommon.SpanID{}
	_, _ = randSource.Read(sid[:])
	return sid
}

func newTraceID() pcommon.TraceID {
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	randSource := rand.New(rand.NewSource(rngSeed))
	tid := pcommon.TraceID{}
	_, _ = randSource.Read(tid[:])
	return tid
}

/* ------------------------------------------ METRICS ------------------------------------------ */

func (r *telemetryAPIReceiver) getMetricTimestamp(metricName string) (time.Time, error) {
	switch metricName {
	case "faas.coldstarts", "faas.init_duration":
		if r.lastMetricTimestamps.platformInitReportTime != nil {
			return *r.lastMetricTimestamps.platformInitReportTime, nil
		}
		// If the time is not set, the faas.coldstarts counter is being zero-initialized
		return time.Now(), nil
	case "faas.invoke_duration", "faas.invocations", "faas.errors", "faas.timeouts":
		if r.lastMetricTimestamps.platformRuntimeEndTime != nil {
			return *r.lastMetricTimestamps.platformRuntimeEndTime, nil
		}
		// If the time is not set, some counter is being zero-initialized
		return time.Now(), nil
	case "faas.mem_usage":
		if r.lastMetricTimestamps.platformReportTime != nil {
			return *r.lastMetricTimestamps.platformReportTime, nil
		}
		return time.Time{}, errMissingTimestamp
	default:
		return time.Time{}, errUnknownMetric
	}
}

/* ------------------------------------------- UTILS ------------------------------------------- */

func parseTime(t string) (time.Time, error) {
	layout := "2006-01-02T15:04:05.000Z"
	return time.Parse(layout, t)
}

func listenOnAddress() string {
	envAwsLocal, ok := os.LookupEnv("AWS_SAM_LOCAL")
	var addr string
	if ok && envAwsLocal == "true" {
		addr = ":" + defaultListenerPort
	} else {
		addr = "sandbox.localdomain:" + defaultListenerPort
	}

	return addr
}
