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
	"strings"
	"time"

	"github.com/golang-collections/go-datastructures/queue"
	telemetryapi "github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver/internal/telemetryapi"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/receiver"
	semconv "go.opentelemetry.io/collector/semconv/v1.25.0"
	"go.uber.org/zap"
)

const defaultListenerPort = "4325"
const initialQueueSize = 5
const timeFormatLayout = "2006-01-02T15:04:05.000Z"
const scopeName = "github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapi"

type telemetryAPIReceiver struct {
	httpServer            *http.Server
	logger                *zap.Logger
	queue                 *queue.Queue // queue is a synchronous queue and is used to put the received log events to be dispatched later
	nextTraces            consumer.Traces
	nextMetrics           consumer.Metrics
	nextLogs              consumer.Logs
	lastPlatformStartTime string
	lastPlatformEndTime   string
	extensionID           string
	resource              pcommon.Resource
}

func (r *telemetryAPIReceiver) Start(ctx context.Context, host component.Host) error {
	address := listenOnAddress()
	r.logger.Info("Listening for requests", zap.String("address", address))

	mux := http.NewServeMux()
	mux.HandleFunc("/", r.httpHandler)
	r.httpServer = &http.Server{Addr: address, Handler: mux}
	go func() {
		_ = r.httpServer.ListenAndServe()
	}()

	telemetryClient := telemetryapi.NewClient(r.logger)
	_, err := telemetryClient.Subscribe(ctx, []telemetryapi.EventType{telemetryapi.Platform, telemetryapi.Function}, r.extensionID, fmt.Sprintf("http://%s/", address))
	if err != nil {
		r.logger.Info("Listening for requests", zap.String("address", address), zap.String("extensionID", r.extensionID))
		return err
	}
	return nil
}

func (r *telemetryAPIReceiver) Shutdown(ctx context.Context) error {
	return nil
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

	var slice []telemetryapi.Event
	if err := json.Unmarshal(body, &slice); err != nil {
		r.logger.Error("error unmarshalling body", zap.Error(err))
		return
	}

	// traces
	if r.nextTraces != nil {
		if traces, err := r.createTraces(slice); err == nil {
			if traces.SpanCount() > 0 {
				err := r.nextTraces.ConsumeTraces(context.Background(), traces)
				if err != nil {
					r.logger.Error("error receiving traces", zap.Error(err))
				}
			}
		}
	}

	// metrics
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

func (r *telemetryAPIReceiver) createTraces(slice []telemetryapi.Event) (ptrace.Traces, error) {
	for _, el := range slice {
		r.logger.Debug(fmt.Sprintf("Event: %s", el.Type), zap.Any("event", el))
		switch el.Type {
		// Function initialization started.
		case string(telemetryapi.PlatformInitStart):
			r.logger.Info(fmt.Sprintf("Init start: %s", r.lastPlatformStartTime), zap.Any("event", el))
			r.lastPlatformStartTime = el.Time
		// Function initialization completed.
		case string(telemetryapi.PlatformInitRuntimeDone):
			r.logger.Info(fmt.Sprintf("Init end: %s", r.lastPlatformEndTime), zap.Any("event", el))
			r.lastPlatformEndTime = el.Time
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
	if len(r.lastPlatformStartTime) > 0 && len(r.lastPlatformEndTime) > 0 {
		td, err := r.createPlatformInitSpan(r.lastPlatformStartTime, r.lastPlatformEndTime)
		if err == nil {
			r.lastPlatformEndTime = ""
			r.lastPlatformStartTime = ""
		}
		return td, err
	}
	return ptrace.Traces{}, errors.New("no traces created")
}

func (r *telemetryAPIReceiver) createMetrics(slice []telemetryapi.Event) (pmetric.Metrics, error) {
	return pmetric.Metrics{}, nil
}

func (r *telemetryAPIReceiver) createLogs(slice []telemetryapi.Event) (plog.Logs, error) {
	log := plog.NewLogs()
	resourceLog := log.ResourceLogs().AppendEmpty()
	r.resource.CopyTo(resourceLog.Resource())
	scopeLog := resourceLog.ScopeLogs().AppendEmpty()
	scopeLog.Scope().SetName(scopeName)
	for _, el := range slice {
		r.logger.Debug(fmt.Sprintf("Event: %s", el.Type), zap.Any("event", el))
		logRecord := scopeLog.LogRecords().AppendEmpty()
		logRecord.Attributes().PutStr("type", el.Type)
		if t, err := time.Parse(timeFormatLayout, el.Time); err == nil {
			logRecord.SetTimestamp(pcommon.NewTimestampFromTime(t))
			logRecord.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		} else {
			r.logger.Error("error parsing time", zap.Error(err))
			return plog.Logs{}, err
		}
		if el.Type == string(telemetryapi.Function) || el.Type == string(telemetryapi.Extension) {
			if record, ok := el.Record.(map[string]interface{}); ok {
				// in JSON format https://docs.aws.amazon.com/lambda/latest/dg/telemetry-schema-reference.html#telemetry-api-function
				if timestamp, ok := record["timestamp"].(string); ok {
					if observedTime, err := time.Parse(timeFormatLayout, timestamp); err == nil {
						logRecord.SetTimestamp(pcommon.NewTimestampFromTime(observedTime))
					} else {
						r.logger.Error("error parsing time", zap.Error(err))
						return plog.Logs{}, err
					}
				}
				if level, ok := record["level"].(string); ok {
					level = strings.ToUpper(level)
					switch level {
					case "TRACE":
						logRecord.SetSeverityNumber(1)
					case "TRACE2":
						logRecord.SetSeverityNumber(2)
					case "TRACE3":
						logRecord.SetSeverityNumber(3)
					case "TRACE4":
						logRecord.SetSeverityNumber(4)
					case "DEBUG":
						logRecord.SetSeverityNumber(5)
					case "DEBUG2":
						logRecord.SetSeverityNumber(6)
					case "DEBUG3":
						logRecord.SetSeverityNumber(7)
					case "DEBUG4":
						logRecord.SetSeverityNumber(8)
					case "INFO":
						logRecord.SetSeverityNumber(9)
					case "INFO2":
						logRecord.SetSeverityNumber(10)
					case "INFO3":
						logRecord.SetSeverityNumber(11)
					case "INFO4":
						logRecord.SetSeverityNumber(12)
					case "WARN":
						logRecord.SetSeverityNumber(13)
					case "WARN2":
						logRecord.SetSeverityNumber(14)
					case "WARN3":
						logRecord.SetSeverityNumber(15)
					case "WARN4":
						logRecord.SetSeverityNumber(16)
					case "ERROR":
						logRecord.SetSeverityNumber(17)
					case "ERROR2":
						logRecord.SetSeverityNumber(18)
					case "ERROR3":
						logRecord.SetSeverityNumber(19)
					case "ERROR4":
						logRecord.SetSeverityNumber(20)
					case "FATAL":
						logRecord.SetSeverityNumber(21)
					case "FATAL2":
						logRecord.SetSeverityNumber(22)
					case "FATAL3":
						logRecord.SetSeverityNumber(23)
					case "FATAL4":
						logRecord.SetSeverityNumber(24)
					case "CRITICAL":
						logRecord.SetSeverityNumber(21)
					case "ALL":
						logRecord.SetSeverityNumber(1)
					default:
					}
					logRecord.SetSeverityText(logRecord.SeverityNumber().String())
				}
				if requestId, ok := record["requestId"].(string); ok {
					logRecord.Attributes().PutStr(semconv.AttributeFaaSInvocationID, requestId)
				}
				if line, ok := record["message"].(string); ok {
					logRecord.Body().SetStr(line)
				}
			} else {
				// in plain text https://docs.aws.amazon.com/lambda/latest/dg/telemetry-schema-reference.html#telemetry-api-function
				if line, ok := el.Record.(string); ok {
					logRecord.Body().SetStr(line)
				}
			}
		} else {
			logRecord.SetSeverityNumber(9)
			logRecord.SetSeverityText(logRecord.SeverityNumber().String())
			if j, err := json.Marshal(el.Record); err == nil {
				logRecord.Body().SetStr(string(j))
			} else {
				r.logger.Error("error stringify record", zap.Error(err))
				return plog.Logs{}, err
			}
		}
	}
	return log, nil
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

func (r *telemetryAPIReceiver) createPlatformInitSpan(start, end string) (ptrace.Traces, error) {
	traceData := ptrace.NewTraces()
	rs := traceData.ResourceSpans().AppendEmpty()
	r.resource.CopyTo(rs.Resource())
	ss := rs.ScopeSpans().AppendEmpty()
	ss.Scope().SetName(scopeName)
	span := ss.Spans().AppendEmpty()
	span.SetTraceID(newTraceID())
	span.SetSpanID(newSpanID())
	span.SetName("platform.initRuntimeDone")
	span.SetKind(ptrace.SpanKindInternal)
	span.Attributes().PutBool(semconv.AttributeFaaSColdstart, true)
	startTime, err := time.Parse(timeFormatLayout, start)
	if err != nil {
		return ptrace.Traces{}, err
	}
	span.SetStartTimestamp(pcommon.NewTimestampFromTime(startTime))
	endTime, err := time.Parse(timeFormatLayout, end)
	if err != nil {
		return ptrace.Traces{}, err
	}
	span.SetEndTimestamp(pcommon.NewTimestampFromTime(endTime))
	return traceData, nil
}

func newTelemetryAPIReceiver(
	cfg *Config,
	set receiver.CreateSettings,
) *telemetryAPIReceiver {
	envResourceMap := map[string]string{
		"AWS_LAMBDA_FUNCTION_MEMORY_SIZE": semconv.AttributeFaaSMaxMemory,
		"AWS_LAMBDA_FUNCTION_VERSION":     semconv.AttributeFaaSVersion,
		"AWS_REGION":                      semconv.AttributeFaaSInvokedRegion,
	}
	r := pcommon.NewResource()
	r.Attributes().PutStr(semconv.AttributeFaaSInvokedProvider, semconv.AttributeFaaSInvokedProviderAWS)
	if val, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); ok {
		r.Attributes().PutStr(semconv.AttributeServiceName, val)
		r.Attributes().PutStr(semconv.AttributeFaaSName, val)
	} else {
		r.Attributes().PutStr(semconv.AttributeServiceName, "unknown_service")
	}

	for env, resourceAttribute := range envResourceMap {
		if val, ok := os.LookupEnv(env); ok {
			r.Attributes().PutStr(resourceAttribute, val)
		}
	}
	return &telemetryAPIReceiver{
		logger:      set.Logger,
		queue:       queue.New(initialQueueSize),
		extensionID: cfg.extensionID,
		resource:    r,
	}
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
