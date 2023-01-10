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
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/golang-collections/go-datastructures/queue"
	"github.com/open-telemetry/opentelemetry-lambda/collector/internal/telemetryapi"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv "go.opentelemetry.io/collector/semconv/v1.5.0"
	"go.uber.org/zap"
)

const defaultListenerPort = "4325"
const initialQueueSize = 5

type telemetryAPIReceiver struct {
	httpServer            *http.Server
	logger                *zap.Logger
	queue                 *queue.Queue // queue is a synchronous queue and is used to put the received log events to be dispatched later
	nextConsumer          consumer.Traces
	lastPlatformStartTime string
	lastPlatformEndTime   string
	extensionID           string
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
	_, err := telemetryClient.Subscribe(ctx, r.extensionID, fmt.Sprintf("http://%s/", address))
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

	for _, el := range slice {
		r.logger.Debug(fmt.Sprintf("Event: %s", el.Type), zap.Any("event", el))
		switch el.Type {
		// Function initialization started.
		case "platform.initStart":
			r.logger.Info(fmt.Sprintf("Init start: %s", r.lastPlatformStartTime), zap.Any("event", el))
			r.lastPlatformStartTime = el.Time
		// Function initialization completed.
		case "platform.initRuntimeDone":
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
		if td, err := r.createPlatformInitSpan(r.lastPlatformStartTime, r.lastPlatformEndTime); err == nil {
			err := r.nextConsumer.ConsumeTraces(context.Background(), td)
			if err == nil {
				r.lastPlatformEndTime = ""
				r.lastPlatformStartTime = ""
			} else {
				r.logger.Error("error receiving traces", zap.Error(err))
			}
		}
	}

	r.logger.Debug("logEvents received", zap.Int("count", len(slice)), zap.Int64("queue_length", r.queue.Len()))
	slice = nil
}

func (r *telemetryAPIReceiver) createPlatformInitSpan(start, end string) (ptrace.Traces, error) {
	traceData := ptrace.NewTraces()
	rs := traceData.ResourceSpans().AppendEmpty()
	ss := rs.ScopeSpans().AppendEmpty()
	ss.Scope().SetName("github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapi")
	span := ss.Spans().AppendEmpty()
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
	set component.ReceiverCreateSettings,
) (*telemetryAPIReceiver, error) {
	return &telemetryAPIReceiver{
		logger:       set.Logger,
		queue:        queue.New(initialQueueSize),
		nextConsumer: next,
		extensionID:  cfg.extensionID,
	}, nil
}

func listenOnAddress() string {
	envAwsLocal, ok := os.LookupEnv("AWS_SAM_LOCAL")
	var addr string
	if ok && envAwsLocal == "true" {
		addr = ":" + defaultListenerPort
	} else {
		addr = "sandbox:" + defaultListenerPort
	}

	return addr
}
