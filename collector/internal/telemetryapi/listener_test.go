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

package telemetryapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func withEnv(t *testing.T, key, value string) {
	t.Helper()
	require.NoError(t, os.Setenv(key, value))
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv(key))
	})
}

func setupListener(t *testing.T) (*Listener, string) {
	t.Helper()
	withEnv(t, "AWS_SAM_LOCAL", "true")
	logger := zaptest.NewLogger(t)
	listener := NewListener(logger)

	address, err := listener.Start()
	require.NoError(t, err)
	return listener, address
}

func submitEvents(t *testing.T, address string, events []Event) {
	t.Helper()
	body, err := json.Marshal(events)
	require.NoError(t, err)

	resp, err := http.Post(address, "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
}

func assertWaitBlocks(t *testing.T, waitDone <-chan error, timeout time.Duration) {
	t.Helper()
	select {
	case err := <-waitDone:
		t.Fatalf("Wait() unexpectedly completed with error: %v", err)
	case <-time.After(timeout):
	}
}

func assertWaitCompletes(t *testing.T, waitDone <-chan error, timeout time.Duration) {
	t.Helper()
	select {
	case err := <-waitDone:
		require.NoError(t, err)
	case <-time.After(timeout):
		t.Fatal("Wait() timed out")
	}
}

type TestEventBuilder struct {
	requestID string
	timestamp time.Time
}

func NewTestEventBuilder(requestID string) *TestEventBuilder {
	return &TestEventBuilder{
		requestID: requestID,
		timestamp: time.Now(),
	}
}

func (b *TestEventBuilder) PlatformStart() Event {
	return Event{
		Type: "platform.start",
		Time: b.timestamp.Format(time.RFC3339),
		Record: map[string]interface{}{
			"requestId": b.requestID,
			"version":   "$LATEST",
		},
	}
}

func (b *TestEventBuilder) PlatformRuntimeDone() Event {
	return Event{
		Type: "platform.runtimeDone",
		Time: b.timestamp.Format(time.RFC3339),
		Record: map[string]interface{}{
			"requestId": b.requestID,
			"status":    "success",
		},
	}
}

func (b *TestEventBuilder) FunctionLog(logLevel, message string) Event {
	return Event{
		Type: "function",
		Time: b.timestamp.Format(time.RFC3339),
		Record: map[string]interface{}{
			"requestId": b.requestID,
			"type":      logLevel,
			"message":   message,
		},
	}
}

func TestNewListener(t *testing.T) {
	logger := zaptest.NewLogger(t)
	listener := NewListener(logger)

	require.NotNil(t, listener, "NewListener() returned nil listener")
	require.Nil(t, listener.httpServer, "httpServer should be initially nil")
	require.NotNil(t, listener.logger, "logger should not be nil")
	require.NotNil(t, listener.queue, "queue should not be nil")
}

func TestListenOnAddress(t *testing.T) {
	testCases := []struct {
		name         string
		envValue     string
		setEnv       bool
		expectedAddr string
	}{
		{
			name:         "AWS_SAM_LOCAL not set",
			setEnv:       false,
			expectedAddr: "sandbox.localdomain",
		},
		{
			name:         "AWS_SAM_LOCAL set to true",
			envValue:     "true",
			setEnv:       true,
			expectedAddr: "",
		},
		{
			name:         "AWS_SAM_LOCAL set to false",
			envValue:     "false",
			setEnv:       true,
			expectedAddr: "sandbox.localdomain",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			require.NoError(t, os.Unsetenv("AWS_SAM_LOCAL"))

			if test.setEnv {
				require.NoError(t, os.Setenv("AWS_SAM_LOCAL", test.envValue))
				defer func() {
					require.NoError(t, os.Unsetenv("AWS_SAM_LOCAL"))
				}()
			}

			addr := listenOnAddress()
			require.Equal(t, test.expectedAddr, addr)
		})
	}
}

func TestListener_StartAndShutdown(t *testing.T) {
	listener, address := setupListener(t)
	require.NotEqual(t, address, "", "Start() should not return an empty address")
	require.True(t, strings.HasPrefix(address, "http://"), "Address should start with http://")
	require.NotNil(t, listener.httpServer, "httpServer should not be nil")

	resp, err := http.Get(address)
	if err != nil {
		t.Errorf("Failed to connect to listener: %v", err)
	} else {
		require.NoError(t, resp.Body.Close())
	}

	listener.Shutdown()
	require.Nil(t, listener.httpServer, "httpServer should be nil after Shutdown()")
}

func TestListener_Shutdown_NotStarted(t *testing.T) {
	logger := zaptest.NewLogger(t)
	listener := NewListener(logger)
	listener.Shutdown()
	require.Nil(t, listener.httpServer, "httpServer should be nil after Shutdown()")
}

func TestListener_httpHandler(t *testing.T) {
	eventBuilder := NewTestEventBuilder("test-request")

	testCases := []struct {
		name          string
		events        []Event
		expectedCount int64
	}{
		{
			name: "single event",
			events: []Event{
				eventBuilder.PlatformStart(),
			},
			expectedCount: 1,
		},
		{
			name: "multiple events",
			events: []Event{
				eventBuilder.PlatformStart(),
				eventBuilder.FunctionLog("INFO", "Received request"),
				eventBuilder.FunctionLog("INFO", "Processing request"),
				eventBuilder.FunctionLog("INFO", "Finished processing request"),
				eventBuilder.PlatformRuntimeDone(),
			},
			expectedCount: 5,
		},
		{
			name:          "empty events array",
			events:        []Event{},
			expectedCount: 0,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			listener, address := setupListener(t)
			defer listener.Shutdown()
			submitEvents(t, address, test.events)
			require.EventuallyWithT(t, func(c *assert.CollectT) {
				require.Equal(c, test.expectedCount, listener.queue.Len())
			}, 1*time.Second, 50*time.Millisecond)
		})
	}
}

func TestListener_httpHandler_InvalidJSON(t *testing.T) {
	withEnv(t, "AWS_SAM_LOCAL", "true")
	logger := zaptest.NewLogger(t)
	listener := NewListener(logger)

	address, err := listener.Start()
	require.NoError(t, err, "Failed to start listener: %v", err)
	defer listener.Shutdown()

	invalidJSON := []byte(`{"invalid": json}`)
	resp, err := http.Post(address, "application/json", bytes.NewReader(invalidJSON))
	require.NoError(t, err, "Failed to post invalid JSON: %v", err)
	require.NoError(t, resp.Body.Close(), "Failed to close response body")

	time.Sleep(50 * time.Millisecond)
	require.Equal(t, listener.queue.Len(), int64(0), "Queue should be empty after invalid JSON")
}

func TestListener_Wait_Success(t *testing.T) {
	eventBuilder := NewTestEventBuilder("target-request")

	testCases := []struct {
		name   string
		events []Event
	}{
		{
			name: "simple request",
			events: []Event{
				eventBuilder.PlatformStart(),
				eventBuilder.FunctionLog("INFO", "Received request"),
				eventBuilder.FunctionLog("INFO", "Processing request"),
				eventBuilder.FunctionLog("INFO", "Finished processing request"),
				eventBuilder.PlatformRuntimeDone(),
			},
		},
		{
			name: "skips wrong request id",
			events: []Event{
				NewTestEventBuilder("other-request-1").PlatformRuntimeDone(),
				eventBuilder.PlatformStart(),
				eventBuilder.FunctionLog("INFO", "Received request"),
				NewTestEventBuilder("other-request-2").PlatformRuntimeDone(),
				eventBuilder.FunctionLog("INFO", "Processing request"),
				eventBuilder.FunctionLog("INFO", "Finished processing request"),
				NewTestEventBuilder("other-request-3").PlatformRuntimeDone(),
				eventBuilder.PlatformRuntimeDone(),
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			listener, address := setupListener(t)
			defer listener.Shutdown()

			waitDone := make(chan error, 1)
			go func() {
				ctx := context.Background()
				waitDone <- listener.Wait(ctx, "target-request")
			}()

			assertWaitBlocks(t, waitDone, 50*time.Millisecond)
			for i, event := range test.events {
				submitEvents(t, address, []Event{event})
				if i < len(test.events)-1 {
					assertWaitBlocks(t, waitDone, 50*time.Millisecond)
				} else {
					assertWaitCompletes(t, waitDone, 1*time.Second)
				}
			}
		})
	}
}

func TestListener_Wait_ContextCanceled(t *testing.T) {
	logger := zaptest.NewLogger(t)
	listener := NewListener(logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := listener.Wait(ctx, "any-req")
	require.Equal(t, context.Canceled, err, "Context should have been canceled")
}
