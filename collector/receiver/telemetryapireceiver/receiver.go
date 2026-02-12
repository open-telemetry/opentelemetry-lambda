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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-collections/go-datastructures/queue"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/receiver"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-lambda/collector/internal/telemetryapi"
)

const (
	initialQueueSize                    = 5
	scopeName                           = "github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapi"
	telemetrySuccessStatus              = "success"
	telemetryFailureStatus              = "failure"
	telemetryErrorStatus                = "error"
	telemetryTimeoutStatus              = "timeout"
	platformReportLogFmt                = "REPORT RequestId: %s Duration: %.2f ms Billed Duration: %.0f ms Memory Size: %.0f MB Max Memory Used: %.0f MB"
	platformStartLogFmt                 = "START RequestId: %s Version: %s"
	platformRuntimeDoneLogFmt           = "END RequestId: %s Version: %s"
	platformInitStartLogFmt             = "INIT_START Runtime Version: %s Runtime Version ARN: %s"
	platformInitRuntimeDoneLogFmt       = "INIT_RUNTIME_DONE Status: %s"
	platformInitReportLogFmt            = "INIT_REPORT Initialization Type: %s Phase: %s Status: %s Duration: %.2f ms"
	platformRestoreStartLogFmt          = "RESTORE_START Runtime Version: %s Runtime Version ARN: %s"
	platformRestoreRuntimeDoneLogFmt    = "RESTORE_RUNTIME_DONE Status: %s"
	platformRestoreReportLogFmt         = "RESTORE_REPORT Status: %s Duration: %.2f ms"
	platformTelemetrySubscriptionLogFmt = "TELEMETRY: %s Subscribed Types: %v"
	platformExtensionLogFmt             = "EXTENSION Name: %s State: %s Events: %v"
	platformLogsDroppedLogFmt           = "LOGS_DROPPED DroppedRecords: %.0f DroppedBytes: %.0f Reason: %s"
)

type telemetryAPIReceiver struct {
	httpServer              *http.Server
	logger                  *zap.Logger
	queue                   *queue.Queue // queue is a synchronous queue and is used to put the received log events to be dispatched later
	mu                      sync.Mutex
	nextTraces              consumer.Traces
	nextMetrics             consumer.Metrics
	nextLogs                consumer.Logs
	lastPlatformStartTime   string
	lastPlatformEndTime     string
	extensionID             string
	port                    int
	types                   []telemetryapi.EventType
	resource                pcommon.Resource
	faasFunctionVersion     string
	faasName                string
	faaSMetricBuilders      *FaaSMetricBuilders
	currentFaasInvocationID string
	logReport               bool
}

func (r *telemetryAPIReceiver) Start(ctx context.Context, host component.Host) error {
	address := listenOnAddress(r.port)
	r.logger.Info("Listening for requests", zap.String("address", address))

	mux := http.NewServeMux()
	mux.HandleFunc("/", r.httpHandler)
	r.httpServer = &http.Server{Addr: address, Handler: mux}
	go func() {
		_ = r.httpServer.ListenAndServe()
	}()

	telemetryClient := telemetryapi.NewClient(r.logger)
	if len(r.types) > 0 {
		_, err := telemetryClient.Subscribe(ctx, r.types, r.extensionID, fmt.Sprintf("http://%s/", address))
		if err != nil {
			r.logger.Info("Listening for requests", zap.String("address", address), zap.String("extensionID", r.extensionID))
			return err
		}
	}
	return nil
}

func (r *telemetryAPIReceiver) Shutdown(ctx context.Context) error {
	return nil
}

func newSpanID() pcommon.SpanID {
	sid := pcommon.SpanID{}
	_, _ = crand.Read(sid[:])
	return sid
}

func newTraceID() pcommon.TraceID {
	tid := pcommon.TraceID{}
	_, _ = crand.Read(tid[:])
	return tid
}

// httpHandler handles the requests coming from the Telemetry API.
// Everytime Telemetry API sends events, this function will read them from the response body
// and put into a synchronous queue to be dispatched later.
// Logging or printing besides the error cases below is not recommended if you have subscribed to
// receive extension logs. Otherwise, logging here will cause Telemetry API to send new logs for
// the printed lines which may create an infinite loop.
func (r *telemetryAPIReceiver) httpHandler(w http.ResponseWriter, req *http.Request) {
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

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, el := range slice {
		r.logger.Debug(fmt.Sprintf("Event: %s", el.Type), zap.Any("event", el))
		switch el.Type {
		// Function initialization started.
		case string(telemetryapi.PlatformInitStart):
			r.logger.Info(fmt.Sprintf("Init start: %s", r.lastPlatformStartTime), zap.Any("event", el))
			r.lastPlatformStartTime = el.Time

			if record, ok := el.Record.(map[string]any); ok {
				functionName, _ := record["functionName"].(string)
				if functionName != "" {
					r.faasName = functionName
				}
			}
		// Function initialization completed.
		case string(telemetryapi.PlatformInitRuntimeDone):
			r.logger.Info(fmt.Sprintf("Init end: %s", r.lastPlatformEndTime), zap.Any("event", el))
			r.lastPlatformEndTime = el.Time

			if len(r.lastPlatformStartTime) > 0 && len(r.lastPlatformEndTime) > 0 {
				if record, ok := el.Record.(map[string]any); ok {
					if td, err := r.createPlatformInitSpan(record, r.lastPlatformStartTime, r.lastPlatformEndTime); err == nil {
						err := r.nextTraces.ConsumeTraces(context.Background(), td)
						if err == nil {
							r.lastPlatformEndTime = ""
							r.lastPlatformStartTime = ""
						} else {
							r.logger.Error("error receiving traces", zap.Error(err))
						}
					}
				}
			}
		}
		// TODO: add support for additional events, see https://docs.aws.amazon.com/lambda/latest/dg/telemetry-api.html
		// A report of function initialization.
		// case "platform.initReport":
		// Function invocation started.
		// case "platform.start":
		// The runtime finished processing an event with either success or failure.
		// case "platform.runtimeDone":
		// A report of function invocation.
		// case "platform.report":
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
	// Metrics
	if r.nextMetrics != nil {
		if metrics, err := r.createMetrics(slice); err == nil {
			if metrics.MetricCount() > 0 {
				err := r.nextMetrics.ConsumeMetrics(context.Background(), metrics)
				if err != nil {
					r.logger.Error("error receiving metrics", zap.Error(err))
				}
			}
		}
	}

	// Logs
	if r.nextLogs != nil {
		if logs, err := r.createLogs(slice); err == nil {
			if logs.LogRecordCount() > 0 {
				err := r.nextLogs.ConsumeLogs(context.Background(), logs)
				if err != nil {
					r.logger.Error("error receiving logs", zap.Error(err))
				}
			}
		}
	}

	r.logger.Debug("logEvents received", zap.Int("count", len(slice)), zap.Int64("queue_length", r.queue.Len()))
	slice = nil
}

func (r *telemetryAPIReceiver) getRecordRequestId(record map[string]interface{}) string {
	if requestId, ok := record["requestId"].(string); ok {
		return requestId
	} else if r.currentFaasInvocationID != "" {
		return r.currentFaasInvocationID
	}
	return ""
}

func (r *telemetryAPIReceiver) createMetrics(slice []event) (pmetric.Metrics, error) {
	metric := pmetric.NewMetrics()
	resourceMetric := metric.ResourceMetrics().AppendEmpty()
	r.resource.CopyTo(resourceMetric.Resource())
	scopeMetric := resourceMetric.ScopeMetrics().AppendEmpty()
	scopeMetric.Scope().SetName(scopeName)
	scopeMetric.SetSchemaUrl(semconv.SchemaURL)

	for _, el := range slice {
		r.logger.Debug(fmt.Sprintf("Event: %s", el.Type), zap.Any("event", el))
		record, ok := el.Record.(map[string]any)
		if !ok {
			continue
		}
		ts, err := time.Parse(time.RFC3339, el.Time)
		if err != nil {
			continue
		}

		switch el.Type {
		case string(telemetryapi.PlatformInitStart):
			r.faaSMetricBuilders.coldstartsMetric.Add(1)
			r.faaSMetricBuilders.coldstartsMetric.AppendDataPoints(scopeMetric, pcommon.NewTimestampFromTime(ts))
		case string(telemetryapi.PlatformInitReport):
			status, _ := record["status"].(string)
			if status == telemetryFailureStatus || status == telemetryErrorStatus {
				r.faaSMetricBuilders.errorsMetric.Add(1)
				r.faaSMetricBuilders.errorsMetric.AppendDataPoints(scopeMetric, pcommon.NewTimestampFromTime(ts))
			} else if status == telemetryTimeoutStatus {
				r.faaSMetricBuilders.timeoutsMetric.Add(1)
				r.faaSMetricBuilders.timeoutsMetric.AppendDataPoints(scopeMetric, pcommon.NewTimestampFromTime(ts))
			}

			metrics, ok := record["metrics"].(map[string]any)
			if !ok {
				continue
			}

			durationMs, ok := metrics["durationMs"].(float64)
			if !ok {
				continue
			}

			r.faaSMetricBuilders.initDurationMetric.Record(durationMs / 1000.0)
			r.faaSMetricBuilders.initDurationMetric.AppendDataPoints(scopeMetric, pcommon.NewTimestampFromTime(ts))
		case string(telemetryapi.PlatformReport):
			metrics, ok := record["metrics"].(map[string]any)
			if !ok {
				continue
			}

			maxMemoryUsedMb, ok := metrics["maxMemoryUsedMB"].(float64)
			if ok {
				r.faaSMetricBuilders.memUsageMetric.Record(maxMemoryUsedMb * 1000000.0)
				r.faaSMetricBuilders.memUsageMetric.AppendDataPoints(scopeMetric, pcommon.NewTimestampFromTime(ts))
			}
		case string(telemetryapi.PlatformRuntimeDone):
			status, _ := record["status"].(string)

			if status == telemetrySuccessStatus {
				r.faaSMetricBuilders.invocationsMetric.Add(1)
				r.faaSMetricBuilders.invocationsMetric.AppendDataPoints(scopeMetric, pcommon.NewTimestampFromTime(ts))
			} else if status == telemetryFailureStatus || status == telemetryErrorStatus {
				r.faaSMetricBuilders.errorsMetric.Add(1)
				r.faaSMetricBuilders.errorsMetric.AppendDataPoints(scopeMetric, pcommon.NewTimestampFromTime(ts))
			} else if status == telemetryTimeoutStatus {
				r.faaSMetricBuilders.timeoutsMetric.Add(1)
				r.faaSMetricBuilders.timeoutsMetric.AppendDataPoints(scopeMetric, pcommon.NewTimestampFromTime(ts))
			}

			metrics, ok := record["metrics"].(map[string]any)
			if !ok {
				continue
			}

			durationMs, ok := metrics["durationMs"].(float64)
			if ok {
				r.faaSMetricBuilders.invokeDurationMetric.Record(durationMs / 1000.0)
				r.faaSMetricBuilders.invokeDurationMetric.AppendDataPoints(scopeMetric, pcommon.NewTimestampFromTime(ts))
			}
		}
	}
	return metric, nil
}

func (r *telemetryAPIReceiver) createLogs(slice []event) (plog.Logs, error) {
	log := plog.NewLogs()
	resourceLog := log.ResourceLogs().AppendEmpty()
	r.resource.CopyTo(resourceLog.Resource())
	scopeLog := resourceLog.ScopeLogs().AppendEmpty()
	scopeLog.Scope().SetName(scopeName)
	for _, el := range slice {
		if !r.logReport && el.Type == string(telemetryapi.PlatformReport) {
			continue
		}
		r.logger.Debug(fmt.Sprintf("Event: %s", el.Type), zap.Any("event", el))
		logRecord := scopeLog.LogRecords().AppendEmpty()
		logRecord.Attributes().PutStr("type", el.Type)
		if t, err := time.Parse(time.RFC3339, el.Time); err == nil {
			logRecord.SetTimestamp(pcommon.NewTimestampFromTime(t))
			logRecord.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		} else {
			r.logger.Error("error parsing time", zap.Error(err))
			return plog.Logs{}, err
		}
		if record, ok := el.Record.(map[string]interface{}); ok {
			requestId := r.getRecordRequestId(record)
			if requestId != "" {
				logRecord.Attributes().PutStr(string(semconv.FaaSInvocationIDKey), requestId)

				// If this is the first event in the invocation with a request id (i.e. the "platform.start" event),
				// set the current invocation id to this request id.
				if el.Type == string(telemetryapi.PlatformStart) {
					r.currentFaasInvocationID = requestId
				}
			}

			// in JSON format https://docs.aws.amazon.com/lambda/latest/dg/telemetry-schema-reference.html#telemetry-api-function
			if timestamp, ok := record["timestamp"].(string); ok {
				if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
					logRecord.SetTimestamp(pcommon.NewTimestampFromTime(t))
				} else {
					r.logger.Error("error parsing time", zap.Error(err))
					return plog.Logs{}, err
				}
			}
			if level, ok := record["level"].(string); ok {
				logRecord.SetSeverityNumber(severityTextToNumber(strings.ToUpper(level)))
				logRecord.SetSeverityText(logRecord.SeverityNumber().String())
			}

			if strings.HasPrefix(el.Type, platform) {
				if el.Type == string(telemetryapi.PlatformInitStart) {
					functionVersion, _ := record["functionVersion"].(string)
					if functionVersion != "" {
						r.faasFunctionVersion = functionVersion
					}
				}

				message := createPlatformMessage(requestId, r.faasFunctionVersion, el.Type, record)
				if message != "" {
					logRecord.Body().SetStr(message)
				}
			} else if line, ok := record["message"].(string); ok {
				logRecord.Body().SetStr(line)
			}
		} else {
			if r.currentFaasInvocationID != "" {
				logRecord.Attributes().PutStr(string(semconv.FaaSInvocationIDKey), r.currentFaasInvocationID)
			}
			// in plain text https://docs.aws.amazon.com/lambda/latest/dg/telemetry-schema-reference.html#telemetry-api-function
			if line, ok := el.Record.(string); ok {
				logRecord.Body().SetStr(line)
			}
		}
		if el.Type == string(telemetryapi.PlatformRuntimeDone) {
			r.currentFaasInvocationID = ""
		}
	}
	return log, nil
}

func createPlatformMessage(requestId string, functionVersion string, eventType string, record map[string]interface{}) string {
	switch eventType {
	case string(telemetryapi.PlatformStart):
		if requestId != "" && functionVersion != "" {
			return fmt.Sprintf(platformStartLogFmt, requestId, functionVersion)
		}
	case string(telemetryapi.PlatformRuntimeDone):
		if requestId != "" && functionVersion != "" {
			return fmt.Sprintf(platformRuntimeDoneLogFmt, requestId, functionVersion)
		}
	case string(telemetryapi.PlatformReport):
		return createPlatformReportMessage(requestId, record)
	case string(telemetryapi.PlatformInitStart):
		runtimeVersion, _ := record["runtimeVersion"].(string)
		runtimeVersionArn, _ := record["runtimeVersionArn"].(string)
		if runtimeVersion != "" || runtimeVersionArn != "" {
			return fmt.Sprintf(platformInitStartLogFmt, runtimeVersion, runtimeVersionArn)
		}
	case string(telemetryapi.PlatformInitRuntimeDone):
		status, _ := record["status"].(string)
		if status != "" {
			return fmt.Sprintf(platformInitRuntimeDoneLogFmt, status)
		}
	case string(telemetryapi.PlatformInitReport):
		initType, _ := record["initializationType"].(string)
		phase, _ := record["phase"].(string)
		status, _ := record["status"].(string)
		var durationMs float64
		durationOk := false
		if metrics, ok := record["metrics"].(map[string]interface{}); ok {
			durationMs, durationOk = metrics["durationMs"].(float64)
		}
		if initType != "" || phase != "" || status != "" || durationOk {
			return fmt.Sprintf(platformInitReportLogFmt, initType, phase, status, durationMs)
		}
	case string(telemetryapi.PlatformRestoreStart):
		runtimeVersion, _ := record["runtimeVersion"].(string)
		runtimeVersionArn, _ := record["runtimeVersionArn"].(string)
		if runtimeVersion != "" || runtimeVersionArn != "" {
			return fmt.Sprintf(platformRestoreStartLogFmt, runtimeVersion, runtimeVersionArn)
		}
	case string(telemetryapi.PlatformRestoreRuntimeDone):
		status, _ := record["status"].(string)
		if status != "" {
			return fmt.Sprintf(platformRestoreRuntimeDoneLogFmt, status)
		}
	case string(telemetryapi.PlatformRestoreReport):
		status, _ := record["status"].(string)
		var durationMs float64
		durationOk := false
		if metrics, ok := record["metrics"].(map[string]interface{}); ok {
			durationMs, durationOk = metrics["durationMs"].(float64)
		}
		if status != "" && durationOk {
			return fmt.Sprintf(platformRestoreReportLogFmt, status, durationMs)
		}
	case string(telemetryapi.PlatformTelemetrySubscription):
		name, _ := record["name"].(string)
		types, _ := record["types"].([]interface{})
		if name != "" {
			return fmt.Sprintf(platformTelemetrySubscriptionLogFmt, name, types)
		}
	case string(telemetryapi.PlatformExtension):
		name, _ := record["name"].(string)
		state, _ := record["state"].(string)
		events, _ := record["events"].([]interface{})
		if name != "" {
			return fmt.Sprintf(platformExtensionLogFmt, name, state, events)
		}
	case string(telemetryapi.PlatformLogsDropped):
		droppedRecords, ok := record["droppedRecords"].(float64)
		if !ok {
			return ""
		}
		droppedBytes, ok := record["droppedBytes"].(float64)
		if !ok {
			return ""
		}
		reason, _ := record["reason"].(string)
		if reason != "" {
			return fmt.Sprintf(platformLogsDroppedLogFmt, droppedRecords, droppedBytes, reason)
		}
	}
	return ""
}

func createPlatformReportMessage(requestId string, record map[string]interface{}) string {
	// gathering metrics
	metrics, ok := record["metrics"].(map[string]interface{})
	if !ok {
		return ""
	}
	var durationMs, billedDurationMs, memorySizeMB, maxMemoryUsedMB float64
	if durationMs, ok = metrics[string(telemetryapi.MetricDurationMs)].(float64); !ok {
		return ""
	}
	if billedDurationMs, ok = metrics[string(telemetryapi.MetricBilledDurationMs)].(float64); !ok {
		return ""
	}
	if memorySizeMB, ok = metrics[string(telemetryapi.MetricMemorySizeMB)].(float64); !ok {
		return ""
	}
	if maxMemoryUsedMB, ok = metrics[string(telemetryapi.MetricMaxMemoryUsedMB)].(float64); !ok {
		return ""
	}

	// optionally gather information about cold start time
	var initDurationMs float64
	if initDurationMsVal, exists := metrics[string(telemetryapi.MetricInitDurationMs)]; exists {
		if val, ok := initDurationMsVal.(float64); ok {
			initDurationMs = val
		}
	}

	message := fmt.Sprintf(
		platformReportLogFmt,
		requestId,
		durationMs,
		billedDurationMs,
		memorySizeMB,
		maxMemoryUsedMB,
	)
	if initDurationMs > 0 {
		message += fmt.Sprintf(" Init Duration: %.2f ms", initDurationMs)
	}

	return message
}

func severityTextToNumber(severityText string) plog.SeverityNumber {
	mapping := map[string]plog.SeverityNumber{
		"TRACE":    plog.SeverityNumberTrace,
		"TRACE2":   plog.SeverityNumberTrace2,
		"TRACE3":   plog.SeverityNumberTrace3,
		"TRACE4":   plog.SeverityNumberTrace4,
		"DEBUG":    plog.SeverityNumberDebug,
		"DEBUG2":   plog.SeverityNumberDebug2,
		"DEBUG3":   plog.SeverityNumberDebug3,
		"DEBUG4":   plog.SeverityNumberDebug4,
		"INFO":     plog.SeverityNumberInfo,
		"INFO2":    plog.SeverityNumberInfo2,
		"INFO3":    plog.SeverityNumberInfo3,
		"INFO4":    plog.SeverityNumberInfo4,
		"WARN":     plog.SeverityNumberWarn,
		"WARN2":    plog.SeverityNumberWarn2,
		"WARN3":    plog.SeverityNumberWarn3,
		"WARN4":    plog.SeverityNumberWarn4,
		"ERROR":    plog.SeverityNumberError,
		"ERROR2":   plog.SeverityNumberError2,
		"ERROR3":   plog.SeverityNumberError3,
		"ERROR4":   plog.SeverityNumberError4,
		"FATAL":    plog.SeverityNumberFatal,
		"FATAL2":   plog.SeverityNumberFatal2,
		"FATAL3":   plog.SeverityNumberFatal3,
		"FATAL4":   plog.SeverityNumberFatal4,
		"CRITICAL": plog.SeverityNumberFatal,
		"ALL":      plog.SeverityNumberTrace,
		"WARNING":  plog.SeverityNumberWarn,
	}
	if ans, ok := mapping[strings.ToUpper(severityText)]; ok {
		return ans
	} else {
		return plog.SeverityNumberUnspecified
	}
}

func (r *telemetryAPIReceiver) registerTracesConsumer(next consumer.Traces) {
	r.nextTraces = next
}

func (r *telemetryAPIReceiver) registerMetricsConsumer(next consumer.Metrics) {
	r.nextMetrics = next
}

func (r *telemetryAPIReceiver) registerLogsConsumer(next consumer.Logs) {
	r.nextLogs = next
}

func (r *telemetryAPIReceiver) createPlatformInitSpan(record map[string]any, start, end string) (ptrace.Traces, error) {
	traceData := ptrace.NewTraces()
	rs := traceData.ResourceSpans().AppendEmpty()
	r.resource.CopyTo(rs.Resource())

	ss := rs.ScopeSpans().AppendEmpty()
	ss.Scope().SetName(scopeName)
	span := ss.Spans().AppendEmpty()
	span.SetTraceID(newTraceID())
	span.SetSpanID(newSpanID())
	span.SetName(fmt.Sprintf("init %s", r.faasName))
	span.SetKind(ptrace.SpanKindInternal)
	span.Attributes().PutBool(string(semconv.FaaSColdstartKey), true)
	startTime, err := time.Parse(time.RFC3339, start)
	if err != nil {
		return ptrace.Traces{}, err
	}
	span.SetStartTimestamp(pcommon.NewTimestampFromTime(startTime))
	endTime, err := time.Parse(time.RFC3339, end)
	if err != nil {
		return ptrace.Traces{}, err
	}
	span.SetEndTimestamp(pcommon.NewTimestampFromTime(endTime))

	status, _ := record["status"].(string)
	if status != "" && status != "success" {
		span.Status().SetCode(ptrace.StatusCodeError)
		errorType, _ := record["errorType"].(string)
		if errorType != "" {
			span.Attributes().PutStr(string(semconv.ErrorTypeKey), errorType)
		} else {
			span.Attributes().PutStr(string(semconv.ErrorTypeKey), status)
		}
	}
	return traceData, nil
}

func getMetricsTemporality(cfg *Config) pmetric.AggregationTemporality {
	temporality := strings.ToLower(cfg.MetricsTemporality)
	if temporality == "" {
		temporality = os.Getenv("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE")
	}

	switch temporality {
	case "delta":
		return pmetric.AggregationTemporalityDelta
	case "cumulative":
		return pmetric.AggregationTemporalityCumulative
	default:
		return pmetric.AggregationTemporalityCumulative
	}
}

func newTelemetryAPIReceiver(
	cfg *Config,
	set receiver.Settings,
) (*telemetryAPIReceiver, error) {
	envResourceMap := map[string]string{
		"AWS_LAMBDA_FUNCTION_MEMORY_SIZE": string(semconv.FaaSMaxMemoryKey),
		"AWS_LAMBDA_FUNCTION_VERSION":     string(semconv.FaaSVersionKey),
		"AWS_REGION":                      string(semconv.FaaSInvokedRegionKey),
	}
	r := pcommon.NewResource()
	r.Attributes().PutStr(string(semconv.FaaSInvokedProviderKey), semconv.FaaSInvokedProviderAWS.Value.AsString())
	if val, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); ok {
		r.Attributes().PutStr(string(semconv.ServiceNameKey), val)
		r.Attributes().PutStr(string(semconv.FaaSNameKey), val)
	} else {
		r.Attributes().PutStr(string(semconv.ServiceNameKey), "unknown_service")
	}

	serviceInstanceID, ok := set.Resource.Attributes().Get(string(semconv.ServiceInstanceIDKey))
	if ok {
		r.Attributes().PutStr(string(semconv.ServiceInstanceIDKey), serviceInstanceID.Str())
	}

	if val, ok := os.LookupEnv("OTEL_SERVICE_NAME"); ok {
		r.Attributes().PutStr(string(semconv.ServiceNameKey), val)
	}

	for env, resourceAttribute := range envResourceMap {
		if val, ok := os.LookupEnv(env); ok {
			r.Attributes().PutStr(resourceAttribute, val)
		}
	}

	var subscribedTypes []telemetryapi.EventType
	for _, val := range cfg.Types {
		switch val {
		case "platform":
			subscribedTypes = append(subscribedTypes, telemetryapi.Platform)
		case "function":
			subscribedTypes = append(subscribedTypes, telemetryapi.Function)
		case "extension":
			subscribedTypes = append(subscribedTypes, telemetryapi.Extension)
		}
	}

	return &telemetryAPIReceiver{
		logger:             set.Logger,
		queue:              queue.New(initialQueueSize),
		extensionID:        cfg.extensionID,
		port:               cfg.Port,
		types:              subscribedTypes,
		resource:           r,
		faaSMetricBuilders: NewFaaSMetricBuilders(pcommon.NewTimestampFromTime(time.Now()), getMetricsTemporality(cfg)),
		logReport:          cfg.LogReport,
	}, nil
}

func listenOnAddress(port int) string {
	envAwsLocal, ok := os.LookupEnv("AWS_SAM_LOCAL")
	var addr string
	if ok && envAwsLocal == "true" {
		addr = ":" + strconv.Itoa(port)
	} else {
		addr = "sandbox.localdomain:" + strconv.Itoa(port)
	}

	return addr
}
