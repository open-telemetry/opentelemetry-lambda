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
	"fmt"
	"math/rand"
	"os"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/receiver"
	semconv "go.opentelemetry.io/collector/semconv/v1.5.0"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-lambda/collector/internal/telemetryapi"
)

type telemetryAPIReceiver struct {
	logger                *zap.Logger
	nextConsumer          consumer.Traces
	lastPlatformStartTime string
	lastPlatformEndTime   string
	resource              pcommon.Resource
}

func (r *telemetryAPIReceiver) Start(ctx context.Context, host component.Host) error {
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

// HandleEvent handles events received by the telemetryapi.
// Logging or printing besides the error cases below is not recommended if you have subscribed to
// receive extension logs. Otherwise, logging here will cause Telemetry API to send new logs for
// the printed lines which may create an infinite loop.
func (r *telemetryAPIReceiver) HandleEvent(ctx context.Context, event telemetryapi.Event) error {
	r.logger.Debug(fmt.Sprintf("Event: %s", event.Type), zap.Any("event", event))
	switch event.Type {
	// Function initialization started.
	case telemetryapi.PlatformInitStart:
		r.logger.Info(fmt.Sprintf("Init start: %s", r.lastPlatformStartTime), zap.Any("event", event))
		r.lastPlatformStartTime = event.Time
		// Function initialization completed.
	case telemetryapi.PlatformInitRuntimeDone:
		r.logger.Info(fmt.Sprintf("Init end: %s", r.lastPlatformEndTime), zap.Any("event", event))
		r.lastPlatformEndTime = event.Time
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

	if len(r.lastPlatformStartTime) > 0 && len(r.lastPlatformEndTime) > 0 {
		if td, err := r.createPlatformInitSpan(r.lastPlatformStartTime, r.lastPlatformEndTime); err == nil {
			err := r.nextConsumer.ConsumeTraces(context.Background(), td)
			if err != nil {
				r.logger.Error("error receiving traces", zap.Error(err))
				return err
			}
			r.lastPlatformEndTime = ""
			r.lastPlatformStartTime = ""
		}
	}
	return nil
}

func (r *telemetryAPIReceiver) createPlatformInitSpan(start, end string) (ptrace.Traces, error) {
	traceData := ptrace.NewTraces()
	rs := traceData.ResourceSpans().AppendEmpty()
	r.resource.CopyTo(rs.Resource())

	ss := rs.ScopeSpans().AppendEmpty()
	ss.Scope().SetName("github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapi")
	span := ss.Spans().AppendEmpty()
	span.SetTraceID(newTraceID())
	span.SetSpanID(newSpanID())
	span.SetName("platform.initRuntimeDone")
	span.SetKind(ptrace.SpanKindInternal)
	span.Attributes().PutBool(semconv.AttributeFaaSColdstart, true)
	layout := "2006-01-02T15:04:05.000Z"
	startTime, err := time.Parse(layout, start)
	if err != nil {
		return ptrace.Traces{}, err
	}
	span.SetStartTimestamp(pcommon.NewTimestampFromTime(startTime))
	endTime, err := time.Parse(layout, end)
	if err != nil {
		return ptrace.Traces{}, err
	}
	span.SetEndTimestamp(pcommon.NewTimestampFromTime(endTime))
	return traceData, nil
}

func newTelemetryAPIReceiver(
	cfg *Config,
	next consumer.Traces,
	set receiver.CreateSettings,
) (*telemetryAPIReceiver, error) {
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
	receiver := telemetryAPIReceiver{
		logger:       set.Logger,
		nextConsumer: next,
		resource:     r,
	}
	telemetryapi.RegisterHandler(&receiver)
	return &receiver, nil
}
