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

package telemetryapireceiver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver/internal/telemetryapi"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/receiver"
	semconv "go.opentelemetry.io/collector/semconv/v1.25.0"
	"go.uber.org/zap"
)

const scopeName = "github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver"

type invocationState struct {
	start time.Time
}

type telemetryAPIReceiver struct {
	config *Config
	logger *zap.Logger

	nextLogs    consumer.Logs
	nextTraces  consumer.Traces
	nextMetrics consumer.Metrics

	httpServer *http.Server
	resource   pcommon.Resource

	// State management for init and invoke phases
	initStartTime time.Time
	invocations   map[string]invocationState
}

func newTelemetryAPIReceiver(
	cfg *Config,
	set receiver.Settings,
) (*telemetryAPIReceiver, error) {
	envResourceMap := map[string]string{
		"AWS_LAMBDA_FUNCTION_MEMORY_SIZE": semconv.AttributeFaaSMaxMemory,
		"AWS_LAMBDA_FUNCTION_VERSION":     semconv.AttributeFaaSVersion,
		"AWS_REGION":                      semconv.AttributeFaaSInvokedRegion,
	}
	r := pcommon.NewResource()
	r.Attributes().PutStr(semconv.AttributeCloudProvider, semconv.AttributeCloudProviderAWS)
	if val, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); ok {
		r.Attributes().PutStr(semconv.AttributeServiceName, val)
		r.Attributes().PutStr(semconv.AttributeFaaSName, val)
	} else {
		r.Attributes().PutStr(semconv.AttributeServiceName, "unknown_service")
	}

	if val, ok := os.LookupEnv("OTEL_SERVICE_NAME"); ok {
		r.Attributes().PutStr(semconv.AttributeServiceName, val)
	}

	for env, resourceAttribute := range envResourceMap {
		if val, ok := os.LookupEnv(env); ok {
			r.Attributes().PutStr(resourceAttribute, val)
		}
	}

	if envID, ok := os.LookupEnv("LOGZIO_ENV_ID"); ok {
		r.Attributes().PutStr("env_id", envID)
	}

	return &telemetryAPIReceiver{
		config:      cfg,
		logger:      set.Logger,
		resource:    r,
		invocations: make(map[string]invocationState),
	}, nil
}

// Start sets up the HTTP server and subscribes to the Telemetry API.
func (r *telemetryAPIReceiver) Start(ctx context.Context, host component.Host) error {
	address := listenOnAddress(r.config.Port)
	r.logger.Info("Starting HTTP server to listen for telemetry.", zap.String("address", address))

	mux := http.NewServeMux()
	mux.HandleFunc("/", r.httpHandler)
	r.httpServer = &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		if err := r.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			r.logger.Fatal("HTTP server failed to start", zap.Error(err))
		}
	}()

	apiClient, err := telemetryapi.NewClient(r.logger)
	if err != nil {
		return fmt.Errorf("failed to create telemetry api client: %w", err)
	}

	extensionID, err := apiClient.Register(ctx, typeStr)
	if err != nil {
		return fmt.Errorf("failed to register extension: %w", err)
	}
	r.config.extensionID = extensionID

	// If the user has configured any types, subscribe to them.
	if len(r.config.Types) > 0 {
		eventTypes := make([]telemetryapi.EventType, len(r.config.Types))
		for i, s := range r.config.Types {
			eventTypes[i] = telemetryapi.EventType(s)
		}
		bufferingCfg := telemetryapi.BufferingCfg{
			MaxItems:  r.config.MaxItems,
			MaxBytes:  r.config.MaxBytes,
			TimeoutMS: r.config.TimeoutMS,
		}
		destinationCfg := telemetryapi.Destination{
			Protocol: telemetryapi.ProtocolHTTP,
			URI:      fmt.Sprintf("http://%s/", address),
		}

		err = apiClient.Subscribe(ctx, extensionID, eventTypes, bufferingCfg, destinationCfg)
		if err != nil {
			return fmt.Errorf("failed to subscribe to Telemetry API: %w", err)
		}
	}

	return nil
}

func (r *telemetryAPIReceiver) Shutdown(ctx context.Context) error {
	if r.httpServer != nil {
		return r.httpServer.Shutdown(ctx)
	}
	return nil
}

// httpHandler processes the incoming telemetry events.
func (r *telemetryAPIReceiver) httpHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.logger.Error("Failed to read request body", zap.Error(err))
		http.Error(w, "error reading body", http.StatusInternalServerError)
		return
	}

	var events []event
	if err := json.Unmarshal(body, &events); err != nil {
		r.logger.Error("Failed to unmarshal telemetry events", zap.Error(err))
		http.Error(w, "error unmarshalling body", http.StatusBadRequest)
		return
	}

	for _, e := range events {
		r.logger.Debug("Processing event", zap.String("type", e.Type))

		switch telemetryapi.EventType(e.Type) {
		// Tracing Events
		case telemetryapi.PlatformInitStart:
			parsedTime, err := time.Parse(time.RFC3339, e.Time)
			if err != nil {
				r.logger.Warn("Failed to parse platform.initStart timestamp, using current time as fallback.",
					zap.String("timestamp", e.Time),
					zap.Error(err))
				r.initStartTime = time.Now()
			} else {
				r.initStartTime = parsedTime
			}
		case telemetryapi.PlatformInitRuntimeDone:
			if !r.initStartTime.IsZero() {
				if traces, err := r.createInitSpan(e); err == nil {
					_ = r.nextTraces.ConsumeTraces(ctx, traces)
				}
				r.initStartTime = time.Time{} // Reset after use
			}
		case telemetryapi.PlatformStart:
			if record, ok := e.Record.(map[string]interface{}); ok {
				if reqID, ok := record["requestId"].(string); ok {
					r.invocations[reqID] = invocationState{start: e.getTime()}
				}
			}
		case telemetryapi.PlatformRuntimeDone:
			if record, ok := e.Record.(map[string]interface{}); ok {
				if reqID, ok := record["requestId"].(string); ok {
					if state, ok := r.invocations[reqID]; ok {
						if traces, err := r.createInvokeSpan(e, state); err == nil {
							_ = r.nextTraces.ConsumeTraces(ctx, traces)
						}
						delete(r.invocations, reqID) // Clean up state
					}
				}
			}

		// Metrics Event
		case telemetryapi.PlatformReport:
			if r.nextMetrics != nil {
				if metrics, err := r.createMetrics(e); err == nil {
					_ = r.nextMetrics.ConsumeMetrics(ctx, metrics)
				}
			}

		// Logs Events
		case telemetryapi.Function, telemetryapi.Extension:
			if r.nextLogs != nil {
				if logs, err := r.createLogs(e); err == nil {
					_ = r.nextLogs.ConsumeLogs(ctx, logs)
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

// --- Consumer Registration ---
func (r *telemetryAPIReceiver) registerLogsConsumer(next consumer.Logs) {
	r.nextLogs = next
}

func (r *telemetryAPIReceiver) registerTracesConsumer(next consumer.Traces) {
	r.nextTraces = next
}

func (r *telemetryAPIReceiver) registerMetricsConsumer(next consumer.Metrics) {
	r.nextMetrics = next
}

// --- Helper Functions ---
func listenOnAddress(port int) string {
	if awsLocal, ok := os.LookupEnv("AWS_SAM_LOCAL"); ok && awsLocal == "true" {
		return "127.0.0.1:" + strconv.Itoa(port)
	}
	return "sandbox.localdomain:" + strconv.Itoa(port)
}
