package telemetryapireceiver // import "github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver"

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-collections/go-datastructures/queue"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/receiver"
	semconv "go.opentelemetry.io/collector/semconv/v1.25.0"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-lambda/collector/internal/telemetryapi"
)

const defaultLogsListenerPort = "4327"
const initialLogsQueueSize = 5
const timeFormatLayout = "2006-01-02T15:04:05.000Z"
const scopeName = "github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapi"

type telemetryAPILogsReceiver struct {
	httpServer   *http.Server
	logger       *zap.Logger
	queue        *queue.Queue // queue is a synchronous queue and is used to put the received log events to be dispatched later
	nextConsumer consumer.Logs
	extensionID  string
	resource     pcommon.Resource
}

func (r *telemetryAPILogsReceiver) Start(ctx context.Context, host component.Host) error {
	address := listenOnLogsAddress()
	r.logger.Info("Listening for requests", zap.String("address", address))

	mux := http.NewServeMux()
	mux.HandleFunc("/", r.httpHandler)
	r.httpServer = &http.Server{Addr: address, Handler: mux}
	go func() {
		_ = r.httpServer.ListenAndServe()
	}()

	telemetryClient := telemetryapi.NewClient(r.logger)
	_, err := telemetryClient.SubscribeEvents(ctx, []telemetryapi.EventType{telemetryapi.Platform, telemetryapi.Function}, r.extensionID, fmt.Sprintf("http://%s/", address))
	if err != nil {
		r.logger.Info("Listening for requests", zap.String("address", address), zap.String("extensionID", r.extensionID))
		return err
	}
	return nil
}

func (r *telemetryAPILogsReceiver) Shutdown(ctx context.Context) error {
	return nil
}

func (r *telemetryAPILogsReceiver) httpHandler(w http.ResponseWriter, req *http.Request) {
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

	if logs, err := r.createLogs(slice); err == nil {
		err := r.nextConsumer.ConsumeLogs(context.Background(), logs)
		if err != nil {
			r.logger.Error("error receiving logs", zap.Error(err))
		}
	}

	r.logger.Debug("logEvents received", zap.Int("count", len(slice)), zap.Int64("queue_length", r.queue.Len()))
	slice = nil
}

func (r *telemetryAPILogsReceiver) createLogs(slice []event) (plog.Logs, error) {
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
					logRecord.SetSeverityText(level)
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
					default:
					}
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
			logRecord.SetSeverityText("INFO")
			logRecord.SetSeverityNumber(9)
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

func newTelemetryAPILogsReceiver(
	cfg *Config,
	next consumer.Logs,
	set receiver.CreateSettings,
) (*telemetryAPILogsReceiver, error) {
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
	return &telemetryAPILogsReceiver{
		logger:       set.Logger,
		queue:        queue.New(initialLogsQueueSize),
		nextConsumer: next,
		extensionID:  cfg.extensionID,
		resource:     r,
	}, nil
}

func listenOnLogsAddress() string {
	envAwsLocal, ok := os.LookupEnv("AWS_SAM_LOCAL")
	var addr string
	if ok && envAwsLocal == "true" {
		addr = ":" + defaultLogsListenerPort
	} else {
		addr = "sandbox.localdomain:" + defaultLogsListenerPort
	}

	return addr
}
