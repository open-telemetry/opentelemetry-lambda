// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/open-telemetry/opentelemetry-lambda/collector/internal/extensionapi"
	"github.com/open-telemetry/opentelemetry-lambda/collector/internal/telemetryapi"
)

type mockCollector struct {
	err     error
	stopped bool
}

func (c *mockCollector) Start(ctx context.Context) error {
	return c.err
}
func (c *mockCollector) Stop() error {
	c.stopped = true
	return c.err
}

func TestRun(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()
	// test with an error
	lm := manager{
		collector:       &mockCollector{err: fmt.Errorf("test start error")},
		logger:          logger,
		extensionClient: extensionapi.NewClient(logger, ""),
	}
	require.Error(t, lm.Run(ctx))
	// test with no waitgroup
	lm = manager{
		collector:       &mockCollector{},
		logger:          logger,
		extensionClient: extensionapi.NewClient(logger, ""),
	}
	require.NoError(t, lm.Run(ctx))
	// test with waitgroup counter incremented
	lm = manager{
		collector:       &mockCollector{},
		logger:          logger,
		extensionClient: extensionapi.NewClient(logger, ""),
	}
	lm.wg.Add(1)
	go func() {
		require.NoError(t, lm.Run(ctx))
	}()
	lm.wg.Done()

}

func TestProcessEvents(t *testing.T) {
	type test struct {
		name           string
		cancel         bool
		err            error
		serverResponse string
		collectorError error
	}
	testCases := []test{
		{
			name:   "processEvents with context cancelled",
			cancel: true,
		},
		{
			name: "processEvents with error from extension API",
			err:  fmt.Errorf("unexpected end of JSON input"),
		},
		{
			name:           "processEvents with shutdown event received",
			serverResponse: `{"time":"2006-01-02T15:04:05.000Z", "eventType":"SHUTDOWN", "record":{}}`,
		},
		{
			name:           "processEvents with shutdown event received and collector error",
			serverResponse: `{"time":"2006-01-02T15:04:05.000Z", "eventType":"SHUTDOWN", "record":{}}`,
			collectorError: fmt.Errorf("test shutdown error"),
			err:            fmt.Errorf("test shutdown error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			var ctx context.Context

			if tc.cancel {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			} else {
				ctx = context.Background()
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				_, err := w.Write([]byte(tc.serverResponse))
				require.NoError(t, err)
				_, err = io.ReadAll(r.Body)
				require.NoError(t, err, "failed to read request body: %v", err)
			}))
			defer server.Close()
			u, err := url.Parse(server.URL)
			require.NoError(t, err)

			lm := manager{
				collector:       &mockCollector{err: tc.collectorError},
				logger:          logger,
				listener:        telemetryapi.NewListener(logger),
				extensionClient: extensionapi.NewClient(logger, u.Host),
			}
			if tc.err != nil {
				err = lm.processEvents(ctx)
				require.Error(t, err)
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, lm.processEvents(ctx))
			}
		})
	}
}

func TestHandleEvent(t *testing.T) {
	type test struct {
		name             string
		err              error
		serverResponses  []string
		collectorError   error
		event            telemetryapi.Event
		collectorStopped bool
		requestID        string
	}
	testCases := []test{
		{
			name: "HandleEvent with event not platform.runtimeDone",
			serverResponses: []string{
				`{"time":"2006-01-02T15:04:05.000Z", "eventType":"INVOKE", "record":{}, "requestId":"1234"}`,
				`{"time":"2006-01-02T15:04:05.000Z", "eventType":"SHUTDOWN", "record":{}, "requestId":"1234"}`,
			},
			event: telemetryapi.Event{
				Type: telemetryapi.PlatformInitRuntimeDone,
			},
			requestID: "1234",
		},
		{
			name: "HandleEvent with event platform.runtimeDone require ID match",
			serverResponses: []string{
				`{"time":"2006-01-02T15:04:05.000Z", "eventType":"INVOKE", "record":{}, "requestId":"1234"}`,
				`{"time":"2006-01-02T15:04:05.000Z", "eventType":"SHUTDOWN", "record":{}, "requestId":"1234"}`,
			},
			event: telemetryapi.Event{
				Type: telemetryapi.PlatformRuntimeDone,
				Record: map[string]any{
					"requestId": "1234",
				},
			},
			collectorStopped: true,
			requestID:        "1234",
		},
		{
			name: "HandleEvent with event platform.runtimeDone require ID does not match",
			serverResponses: []string{
				`{"time":"2006-01-02T15:04:05.000Z", "eventType":"INVOKE", "record":{}, "requestId":"1234"}`,
				`{"time":"2006-01-02T15:04:05.000Z", "eventType":"SHUTDOWN", "record":{}, "requestId":"1234"}`,
			},
			err: errors.New("request ID doesn't match"),
			event: telemetryapi.Event{
				Type: telemetryapi.PlatformRuntimeDone,
			},
			collectorStopped: true,
			requestID:        "1234",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			responseIndex := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				_, err := w.Write([]byte(tc.serverResponses[responseIndex]))
				require.NoError(t, err)
				responseIndex++
				_, err = io.ReadAll(r.Body)
				require.NoError(t, err, "failed to read request body: %v", err)
			}))
			defer server.Close()
			u, err := url.Parse(server.URL)
			require.NoError(t, err)

			collector := &mockCollector{err: tc.collectorError}
			lm := manager{
				collector:       collector,
				logger:          logger,
				listener:        telemetryapi.NewListener(logger),
				extensionClient: extensionapi.NewClient(logger, u.Host),
			}
			go func() {
				require.NoError(t, lm.processEvents(context.Background()))
			}()

			// loop until processEvents is waiting for HandleEvent
			waiting := false
			for !waiting {
				time.Sleep(10 * time.Millisecond)
				lm.reqIDLock.RLock()
				if lm.reqID == tc.requestID {
					waiting = true
				}
				lm.reqIDLock.RUnlock()
			}

			err = lm.HandleEvent(context.Background(), tc.event)
			if tc.err != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tc.collectorStopped {
				lm.wg.Wait()
			}
			require.Equal(t, tc.collectorStopped, collector.stopped)
		})
	}
}
