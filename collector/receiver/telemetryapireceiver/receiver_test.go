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
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/receiver/receivertest"
	semconv "go.opentelemetry.io/collector/semconv/v1.25.0"
)

func TestListenOnAddress(t *testing.T) {
	testCases := []struct {
		desc     string
		testFunc func(*testing.T)
	}{
		{
			desc: "listen on address without AWS_SAM_LOCAL env variable",
			testFunc: func(t *testing.T) {
				addr := listenOnAddress(4325)
				require.EqualValues(t, "sandbox.localdomain:4325", addr)
			},
		},
		{
			desc: "listen on address with AWS_SAM_LOCAL env variable",
			testFunc: func(t *testing.T) {
				t.Setenv("AWS_SAM_LOCAL", "true")
				addr := listenOnAddress(4325)
				require.EqualValues(t, ":4325", addr)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, tc.testFunc)
	}
}

type mockConsumer struct {
	consumed int
}

func (c *mockConsumer) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	c.consumed += td.SpanCount()
	return nil
}

func (c *mockConsumer) ConsumeLogs(ctx context.Context, td plog.Logs) error {
	return nil
}

func (c *mockConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func TestHandler(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc          string
		body          string
		expectedSpans int
	}{
		{
			desc: "empty body",
			body: `{}`,
		},
		{
			desc: "invalid json",
			body: `invalid json`,
		},
		{
			desc: "valid event",
			body: `[{"time":"", "type":"", "record": {}}]`,
		},
		{
			desc: "valid event",
			body: `[{"time":"", "type":"platform.initStart", "record": {}}]`,
		},
		{
			desc: "valid start/end events",
			body: `[
				{"time":"2006-01-02T15:04:04.000Z", "type":"platform.initStart", "record": {}},
				{"time":"2006-01-02T15:04:05.000Z", "type":"platform.initRuntimeDone", "record": {}}
			]`,
			expectedSpans: 1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			consumer := mockConsumer{}
			r, err := newTelemetryAPIReceiver(
				&Config{},
				receivertest.NewNopSettings(Type),
			)
			require.NoError(t, err)
			r.registerTracesConsumer(&consumer)
			req := httptest.NewRequest("POST",
				"http://localhost:53612/someevent", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			r.httpHandler(rec, req)
			require.Equal(t, tc.expectedSpans, consumer.consumed)
		})
	}
}

func TestCreatePlatformInitSpan(t *testing.T) {
	testCases := []struct {
		desc        string
		start       string
		end         string
		expected    int
		expectError bool
	}{
		{
			desc:        "no start/end times",
			expectError: true,
		},
		{
			desc:        "no end time",
			start:       "2006-01-02T15:04:05.000Z",
			expectError: true,
		},
		{
			desc:        "no start times",
			end:         "2006-01-02T15:04:05.000Z",
			expectError: true,
		},
		{
			desc:        "valid times",
			start:       "2006-01-02T15:04:04.000Z",
			end:         "2006-01-02T15:04:05.000Z",
			expected:    1,
			expectError: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			r, err := newTelemetryAPIReceiver(
				&Config{},
				receivertest.NewNopSettings(Type),
			)
			require.NoError(t, err)
			td, err := r.createPlatformInitSpan(tc.start, tc.end)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.Equal(t, tc.expected, td.SpanCount())
			}
		})
	}
}

func TestCreateLogs(t *testing.T) {
	t.Parallel()

	type logInfo struct {
		logType           string
		timestamp         string
		body              string
		severityText      string
		containsRequestId bool
		requestId         string
		severityNumber    plog.SeverityNumber
	}

	testCases := []struct {
		desc         string
		slice        []event
		expectedLogs []logInfo
		expectError  bool
	}{
		{
			desc:         "no slice",
			expectedLogs: []logInfo{},
			expectError:  false,
		},
		{
			desc: "Invalid Timestamp",
			slice: []event{
				{
					Time:   "invalid",
					Type:   "function",
					Record: "[INFO] Hello world, I am an extension!",
				},
			},
			expectError: true,
		},
		{
			desc: "function text",
			slice: []event{
				{
					Time:   "2022-10-12T00:03:50.000Z",
					Type:   "function",
					Record: "[INFO] Hello world, I am an extension!",
				},
			},
			expectedLogs: []logInfo{
				{
					logType:           "function",
					timestamp:         "2022-10-12T00:03:50.000Z",
					body:              "[INFO] Hello world, I am an extension!",
					containsRequestId: false,
					severityNumber:    plog.SeverityNumberUnspecified,
				},
			},
			expectError: false,
		},
		{
			desc: "function text with requestId",
			slice: []event{
				{
					Time: "2022-10-12T00:03:50.000Z",
					Type: "platform.start",
					Record: map[string]any{
						"requestId": "34472c47-5ff0-4df5-a9ad-03776afa5473",
					},
				},
				{
					Time:   "2022-10-12T00:03:50.000Z",
					Type:   "function",
					Record: "[INFO] Hello world, I am an extension!",
				},
				{
					Time:   "2022-10-12T00:03:50.000Z",
					Type:   "platform.runtimeDone",
					Record: map[string]any{},
				},
			},
			expectedLogs: []logInfo{
				{
					logType:           "platform.start",
					timestamp:         "2022-10-12T00:03:50.000Z",
					body:              "",
					containsRequestId: true,
					requestId:         "34472c47-5ff0-4df5-a9ad-03776afa5473",
					severityNumber:    plog.SeverityNumberUnspecified,
				},
				{
					logType:           "function",
					timestamp:         "2022-10-12T00:03:50.000Z",
					body:              "[INFO] Hello world, I am an extension!",
					containsRequestId: true,
					requestId:         "34472c47-5ff0-4df5-a9ad-03776afa5473",
					severityNumber:    plog.SeverityNumberUnspecified,
				},
				{
					logType:           "platform.runtimeDone",
					timestamp:         "2022-10-12T00:03:50.000Z",
					body:              "",
					containsRequestId: true,
					requestId:         "34472c47-5ff0-4df5-a9ad-03776afa5473",
					severityNumber:    plog.SeverityNumberUnspecified,
				},
			},
			expectError: false,
		},
		{
			desc: "function json",
			slice: []event{
				{
					Time: "2022-10-12T00:03:50.000Z",
					Type: "function",
					Record: map[string]any{
						"timestamp": "2022-10-12T00:03:50.000Z",
						"level":     "INFO",
						"requestId": "79b4f56e-95b1-4643-9700-2807f4e68189",
						"message":   "Hello world, I am a function!",
					},
				},
			},
			expectedLogs: []logInfo{
				{
					logType:           "function",
					timestamp:         "2022-10-12T00:03:50.000Z",
					body:              "Hello world, I am a function!",
					containsRequestId: true,
					requestId:         "79b4f56e-95b1-4643-9700-2807f4e68189",
					severityText:      "Info",
					severityNumber:    plog.SeverityNumberInfo,
				},
			},
			expectError: false,
		},
		{
			desc: "extension text",
			slice: []event{
				{
					Time:   "2022-10-12T00:03:50.000Z",
					Type:   "extension",
					Record: "[INFO] Hello world, I am an extension!",
				},
			},
			expectedLogs: []logInfo{
				{
					logType:           "extension",
					timestamp:         "2022-10-12T00:03:50.000Z",
					body:              "[INFO] Hello world, I am an extension!",
					containsRequestId: false,
					severityNumber:    plog.SeverityNumberUnspecified,
				},
			},
			expectError: false,
		},
		{
			desc: "extension json",
			slice: []event{
				{
					Time: "2022-10-12T00:03:50.000Z",
					Type: "extension",
					Record: map[string]any{
						"timestamp": "2022-10-12T00:03:50.000Z",
						"level":     "INFO",
						"requestId": "79b4f56e-95b1-4643-9700-2807f4e68689",
						"message":   "Hello world, I am an extension!",
					},
				},
			},
			expectedLogs: []logInfo{
				{
					logType:           "extension",
					timestamp:         "2022-10-12T00:03:50.000Z",
					body:              "Hello world, I am an extension!",
					containsRequestId: true,
					requestId:         "79b4f56e-95b1-4643-9700-2807f4e68689",
					severityText:      "Info",
					severityNumber:    plog.SeverityNumberInfo,
				},
			},
			expectError: false,
		},
		{
			desc: "extension json anything",
			slice: []event{
				{
					Time: "2022-10-12T00:03:50.000Z",
					Type: "extension",
					Record: map[string]any{
						"timestamp": "2022-10-12T00:03:50.000Z",
						"level":     "anything",
						"requestId": "79b4f56e-95b1-4643-9700-2807f4e68689",
						"message":   "Hello world, I am an extension!",
					},
				},
			},
			expectedLogs: []logInfo{
				{
					logType:           "extension",
					timestamp:         "2022-10-12T00:03:50.000Z",
					body:              "Hello world, I am an extension!",
					containsRequestId: true,
					requestId:         "79b4f56e-95b1-4643-9700-2807f4e68689",
					severityText:      "Unspecified",
					severityNumber:    plog.SeverityNumberUnspecified,
				},
			},
			expectError: false,
		},
		{
			desc: "platform.initStart",
			slice: []event{
				{
					Time:   "2022-10-12T00:03:50.000Z",
					Type:   "platform.initStart",
					Record: map[string]any{},
				},
			},
			expectedLogs: []logInfo{
				{
					logType:           "platform.initStart",
					timestamp:         "2022-10-12T00:03:50.000Z",
					body:              "",
					containsRequestId: false,
					severityNumber:    plog.SeverityNumberUnspecified,
				},
			},
			expectError: false,
		},
		{
			desc: "platform.initRuntimeDone",
			slice: []event{
				{
					Time:   "2022-10-12T00:03:50.000Z",
					Type:   "platform.initRuntimeDone",
					Record: map[string]any{},
				},
			},
			expectedLogs: []logInfo{
				{
					logType:           "platform.initRuntimeDone",
					timestamp:         "2022-10-12T00:03:50.000Z",
					body:              "",
					containsRequestId: false,
					severityNumber:    plog.SeverityNumberUnspecified,
				},
			},
			expectError: false,
		},
		{
			desc: "platform.initReport",
			slice: []event{
				{
					Time:   "2022-10-12T00:03:50.000Z",
					Type:   "platform.initReport",
					Record: map[string]any{},
				},
			},
			expectedLogs: []logInfo{
				{
					logType:           "platform.initReport",
					timestamp:         "2022-10-12T00:03:50.000Z",
					body:              "",
					containsRequestId: false,
					severityNumber:    plog.SeverityNumberUnspecified,
				},
			},
			expectError: false,
		},
		{
			desc: "platform.start",
			slice: []event{
				{
					Time: "2022-10-12T00:03:50.000Z",
					Type: "platform.start",
					Record: map[string]any{
						"requestId": "test-id",
					},
				},
			},
			expectedLogs: []logInfo{
				{
					logType:           "platform.start",
					timestamp:         "2022-10-12T00:03:50.000Z",
					body:              "",
					containsRequestId: true,
					requestId:         "test-id",
					severityNumber:    plog.SeverityNumberUnspecified,
				},
			},
			expectError: false,
		},
		{
			desc: "platform.runtimeDone",
			slice: []event{
				{
					Time:   "2022-10-12T00:03:50.000Z",
					Type:   "platform.runtimeDone",
					Record: map[string]any{},
				},
			},
			expectedLogs: []logInfo{
				{
					logType:           "platform.runtimeDone",
					timestamp:         "2022-10-12T00:03:50.000Z",
					body:              "",
					containsRequestId: false,
					severityNumber:    plog.SeverityNumberUnspecified,
				},
			},
			expectError: false,
		},
		{
			desc: "platform.report",
			slice: []event{
				{
					Time:   "2022-10-12T00:03:50.000Z",
					Type:   "platform.report",
					Record: map[string]any{},
				},
			},
			expectedLogs: []logInfo{
				{
					logType:           "platform.report",
					timestamp:         "2022-10-12T00:03:50.000Z",
					body:              "",
					containsRequestId: false,
					severityNumber:    plog.SeverityNumberUnspecified,
				},
			},
			expectError: false,
		},
		{
			desc: "platform.restoreStart",
			slice: []event{
				{
					Time:   "2022-10-12T00:03:50.000Z",
					Type:   "platform.restoreStart",
					Record: map[string]any{},
				},
			},
			expectedLogs: []logInfo{
				{
					logType:           "platform.restoreStart",
					timestamp:         "2022-10-12T00:03:50.000Z",
					body:              "",
					containsRequestId: false,
					severityNumber:    plog.SeverityNumberUnspecified,
				},
			},
			expectError: false,
		},
		{
			desc: "platform.restoreRuntimeDone",
			slice: []event{
				{
					Time:   "2022-10-12T00:03:50.000Z",
					Type:   "platform.restoreRuntimeDone",
					Record: map[string]any{},
				},
			},
			expectedLogs: []logInfo{
				{
					logType:           "platform.restoreRuntimeDone",
					timestamp:         "2022-10-12T00:03:50.000Z",
					body:              "",
					containsRequestId: false,
					severityNumber:    plog.SeverityNumberUnspecified,
				},
			},
			expectError: false,
		},
		{
			desc: "platform.restoreReport",
			slice: []event{
				{
					Time:   "2022-10-12T00:03:50.000Z",
					Type:   "platform.restoreReport",
					Record: map[string]any{},
				},
			},
			expectedLogs: []logInfo{
				{
					logType:           "platform.restoreReport",
					timestamp:         "2022-10-12T00:03:50.000Z",
					body:              "",
					containsRequestId: false,
					severityNumber:    plog.SeverityNumberUnspecified,
				},
			},
			expectError: false,
		},
		{
			desc: "platform.telemetrySubscription",
			slice: []event{
				{
					Time:   "2022-10-12T00:03:50.000Z",
					Type:   "platform.telemetrySubscription",
					Record: map[string]any{},
				},
			},
			expectedLogs: []logInfo{
				{
					logType:           "platform.telemetrySubscription",
					timestamp:         "2022-10-12T00:03:50.000Z",
					body:              "",
					containsRequestId: false,
					severityNumber:    plog.SeverityNumberUnspecified,
				},
			},
			expectError: false,
		},
		{
			desc: "platform.logsDropped",
			slice: []event{
				{
					Time:   "2022-10-12T00:03:50.000Z",
					Type:   "platform.logsDropped",
					Record: map[string]any{},
				},
			},
			expectedLogs: []logInfo{
				{
					logType:           "platform.logsDropped",
					timestamp:         "2022-10-12T00:03:50.000Z",
					body:              "",
					containsRequestId: false,
					severityNumber:    plog.SeverityNumberUnspecified,
				},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			r, err := newTelemetryAPIReceiver(
				&Config{},
				receivertest.NewNopSettings(Type),
			)
			require.NoError(t, err)
			log, err := r.createLogs(tc.slice)
			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, 1, log.ResourceLogs().Len())
			resourceLog := log.ResourceLogs().At(0)
			require.Equal(t, 1, resourceLog.ScopeLogs().Len())
			scopeLog := resourceLog.ScopeLogs().At(0)
			require.Equal(t, scopeName, scopeLog.Scope().Name())
			require.Equal(t, len(tc.expectedLogs), scopeLog.LogRecords().Len())

			for i, expected := range tc.expectedLogs {
				logRecord := scopeLog.LogRecords().At(i)

				attr, ok := logRecord.Attributes().Get("type")
				require.True(t, ok)
				require.Equal(t, expected.logType, attr.Str())

				expectedTime, err := time.Parse(time.RFC3339, expected.timestamp)
				require.NoError(t, err)
				require.Equal(t, pcommon.NewTimestampFromTime(expectedTime), logRecord.Timestamp())

				requestId, ok := logRecord.Attributes().Get(semconv.AttributeFaaSInvocationID)
				require.Equal(t, expected.containsRequestId, ok)
				if ok {
					require.Equal(t, expected.requestId, requestId.Str())
				}

				require.Equal(t, expected.severityText, logRecord.SeverityText())
				require.Equal(t, expected.severityNumber, logRecord.SeverityNumber())
				require.Equal(t, expected.body, logRecord.Body().Str())
			}
		})
	}
}

func TestCreateLogsWithLogReport(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc               string
		slice              []event
		logReport          bool
		expectedLogRecords int
		expectedType       string
		expectedTimestamp  string
		expectedBody       string
		expectedAttributes map[string]interface{}
		expectError        bool
	}{
		{
			desc: "platform.report with logReport enabled - valid metrics",
			slice: []event{
				{
					Time: "2022-10-12T00:03:50.000Z",
					Type: "platform.report",
					Record: map[string]any{
						"requestId": "test-request-id-123",
						"metrics": map[string]any{
							"durationMs":       123.45,
							"billedDurationMs": float64(124),
							"memorySizeMB":     float64(512),
							"maxMemoryUsedMB":  float64(256),
						},
					},
				},
			},
			logReport:          true,
			expectedLogRecords: 1,
			expectedType:       "platform.report",
			expectedTimestamp:  "2022-10-12T00:03:50.000Z",
			expectedBody:       "REPORT RequestId: test-request-id-123 Duration: 123.45 ms Billed Duration: 124 ms Memory Size: 512 MB Max Memory Used: 256 MB",
			expectError:        false,
		},
		{
			desc: "platform.report with logReport disabled",
			slice: []event{
				{
					Time: "2022-10-12T00:03:50.000Z",
					Type: "platform.report",
					Record: map[string]any{
						"requestId": "test-request-id-123",
						"metrics": map[string]any{
							"durationMs":       123.45,
							"billedDurationMs": 124,
							"memorySizeMB":     512,
							"maxMemoryUsedMB":  256,
						},
					},
				},
			},
			logReport:          false,
			expectedLogRecords: 0,
			expectError:        false,
		},
		{
			desc: "platform.report with logReport enabled - missing requestId",
			slice: []event{
				{
					Time: "2022-10-12T00:03:50.000Z",
					Type: "platform.report",
					Record: map[string]any{
						"metrics": map[string]any{
							"durationMs":       123.45,
							"billedDurationMs": 124,
							"memorySizeMB":     512,
							"maxMemoryUsedMB":  256,
						},
					},
				},
			},
			logReport:          false,
			expectedLogRecords: 0,
			expectError:        false,
		},
		{
			desc: "platform.report with logReport enabled - invalid timestamp",
			slice: []event{
				{
					Time: "invalid-timestamp",
					Type: "platform.report",
					Record: map[string]any{
						"requestId": "test-request-id-123",
						"metrics": map[string]any{
							"durationMs":       123.45,
							"billedDurationMs": 124,
							"memorySizeMB":     512,
							"maxMemoryUsedMB":  256,
						},
					},
				},
			},
			logReport:          false,
			expectedLogRecords: 0,
			expectError:        false,
		},
		{
			desc: "platform.report with logReport enabled - missing metrics",
			slice: []event{
				{
					Time: "2022-10-12T00:03:50.000Z",
					Type: "platform.report",
					Record: map[string]any{
						"requestId": "test-request-id-123",
					},
				},
			},
			logReport:          false,
			expectedLogRecords: 0,
			expectError:        false,
		},
		{
			desc: "platform.report with logReport enabled - invalid metrics format",
			slice: []event{
				{
					Time: "2022-10-12T00:03:50.000Z",
					Type: "platform.report",
					Record: map[string]any{
						"requestId": "test-request-id-123",
						"metrics": map[string]any{
							"durationMs":       "invalid",
							"billedDurationMs": 124,
							"memorySizeMB":     512,
							"maxMemoryUsedMB":  256,
						},
					},
				},
			},
			logReport:          false,
			expectedLogRecords: 0,
			expectError:        false,
		},
		{
			desc: "platform.report with logReport enabled - record not a map",
			slice: []event{
				{
					Time:   "2022-10-12T00:03:50.000Z",
					Type:   "platform.report",
					Record: "invalid record format",
				},
			},
			logReport:          true,
			expectedLogRecords: 1,
			expectError:        false,
			expectedType:       "platform.report",
			expectedTimestamp:  "2022-10-12T00:03:50.000Z",
			expectedBody:       "invalid record format",
		},
		{
			desc: "platform.report with logReport enabled - with initDurationMs",
			slice: []event{
				{
					Time: "2022-10-12T00:03:50.000Z",
					Type: "platform.report",
					Record: map[string]any{
						"requestId": "test-request-id-123",
						"metrics": map[string]any{
							"durationMs":       123.45,
							"billedDurationMs": 124.0,
							"memorySizeMB":     512.0,
							"maxMemoryUsedMB":  256.0,
							"initDurationMs":   50.5,
						},
					},
				},
			},
			logReport:          true,
			expectedLogRecords: 1,
			expectedType:       "platform.report",
			expectedTimestamp:  "2022-10-12T00:03:50.000Z",
			expectedBody:       "REPORT RequestId: test-request-id-123 Duration: 123.45 ms Billed Duration: 124 ms Memory Size: 512 MB Max Memory Used: 256 MB Init Duration: 50.50 ms",
			expectError:        false,
		},
		{
			desc: "platform.report with logReport enabled - with invalid initDurationMs type",
			slice: []event{
				{
					Time: "2022-10-12T00:03:50.000Z",
					Type: "platform.report",
					Record: map[string]any{
						"requestId": "test-request-id-123",
						"metrics": map[string]any{
							"durationMs":       123.45,
							"billedDurationMs": 124.0,
							"memorySizeMB":     512.0,
							"maxMemoryUsedMB":  256.0,
							"initDurationMs":   "invalid-string",
						},
					},
				},
			},
			logReport:          true,
			expectedLogRecords: 1,
			expectedType:       "platform.report",
			expectedTimestamp:  "2022-10-12T00:03:50.000Z",
			expectedBody:       "REPORT RequestId: test-request-id-123 Duration: 123.45 ms Billed Duration: 124 ms Memory Size: 512 MB Max Memory Used: 256 MB",
			expectError:        false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			r, err := newTelemetryAPIReceiver(
				&Config{LogReport: &tc.logReport},
				receivertest.NewNopSettings(Type),
			)
			require.NoError(t, err)
			log, err := r.createLogs(tc.slice)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, 1, log.ResourceLogs().Len())
				resourceLog := log.ResourceLogs().At(0)
				require.Equal(t, 1, resourceLog.ScopeLogs().Len())
				scopeLog := resourceLog.ScopeLogs().At(0)
				require.Equal(t, scopeName, scopeLog.Scope().Name())
				require.Equal(t, tc.expectedLogRecords, scopeLog.LogRecords().Len())
				if scopeLog.LogRecords().Len() > 0 {
					logRecord := scopeLog.LogRecords().At(0)
					attr, ok := logRecord.Attributes().Get("type")
					require.True(t, ok)
					require.Equal(t, tc.expectedType, attr.Str())
					if tc.expectedTimestamp != "" {
						expectedTime, err := time.Parse(time.RFC3339, tc.expectedTimestamp)
						require.NoError(t, err)
						require.Equal(t, pcommon.NewTimestampFromTime(expectedTime), logRecord.Timestamp())
					} else {
						// For invalid timestamps, no timestamp should be set (zero value)
						require.Equal(t, pcommon.Timestamp(0), logRecord.Timestamp())
					}
					require.Equal(t, tc.expectedBody, logRecord.Body().Str())
				}
			}
		})
	}
}

func TestSeverityTextToNumber(t *testing.T) {
	t.Parallel()

	goldenMapping := map[string]plog.SeverityNumber{
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
	for level, number := range goldenMapping {
		require.Equal(t, number, severityTextToNumber(level))
	}

	others := []string{"", "UNKNOWN", "other", "anything"}
	for _, level := range others {
		require.Equal(t, plog.SeverityNumberUnspecified, severityTextToNumber(level))
	}
}
