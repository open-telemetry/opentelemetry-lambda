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

	telemetryapi "github.com/open-telemetry/opentelemetry-lambda/collector/internal/telemetryapi"
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
			r := newTelemetryAPIReceiver(
				&Config{},
				receivertest.NewNopCreateSettings(),
			)
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
			r := newTelemetryAPIReceiver(
				&Config{},
				receivertest.NewNopCreateSettings(),
			)
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

	testCases := []struct {
		desc                      string
		slice                     []telemetryapi.Event
		expectedLogRecords        int
		expectedType              string
		expectedTimestamp         string
		expectedBody              string
		expectedSeverityText      string
		expectedContainsRequestId bool
		expectedRequestId         string
		expectedSeverityNumber    plog.SeverityNumber
		expectError               bool
	}{
		{
			desc:               "no slice",
			expectedLogRecords: 0,
			expectError:        false,
		},
		{
			desc: "Invalid Timestamp",
			slice: []telemetryapi.Event{
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
			slice: []telemetryapi.Event{
				{
					Time:   "2022-10-12T00:03:50.000Z",
					Type:   "function",
					Record: "[INFO] Hello world, I am an extension!",
				},
			},
			expectedLogRecords:        1,
			expectedType:              "function",
			expectedTimestamp:         "2022-10-12T00:03:50.000Z",
			expectedBody:              "[INFO] Hello world, I am an extension!",
			expectedContainsRequestId: false,
			expectedSeverityText:      "",
			expectedSeverityNumber:    plog.SeverityNumberUnspecified,
			expectError:               false,
		},
		{
			desc: "function json",
			slice: []telemetryapi.Event{
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
			expectedLogRecords:        1,
			expectedType:              "function",
			expectedTimestamp:         "2022-10-12T00:03:50.000Z",
			expectedBody:              "Hello world, I am a function!",
			expectedContainsRequestId: true,
			expectedRequestId:         "79b4f56e-95b1-4643-9700-2807f4e68189",
			expectedSeverityText:      "Info",
			expectedSeverityNumber:    plog.SeverityNumberInfo,
			expectError:               false,
		},
		{
			desc: "extension text",
			slice: []telemetryapi.Event{
				{
					Time:   "2022-10-12T00:03:50.000Z",
					Type:   "extension",
					Record: "[INFO] Hello world, I am an extension!",
				},
			},
			expectedLogRecords:        1,
			expectedType:              "extension",
			expectedTimestamp:         "2022-10-12T00:03:50.000Z",
			expectedBody:              "[INFO] Hello world, I am an extension!",
			expectedContainsRequestId: false,
			expectedSeverityText:      "",
			expectedSeverityNumber:    plog.SeverityNumberUnspecified,
			expectError:               false,
		},
		{
			desc: "extension json",
			slice: []telemetryapi.Event{
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
			expectedLogRecords:        1,
			expectedType:              "extension",
			expectedTimestamp:         "2022-10-12T00:03:50.000Z",
			expectedBody:              "Hello world, I am an extension!",
			expectedContainsRequestId: true,
			expectedRequestId:         "79b4f56e-95b1-4643-9700-2807f4e68689",
			expectedSeverityText:      "Info",
			expectedSeverityNumber:    plog.SeverityNumberInfo,
			expectError:               false,
		},
		{
			desc: "extension json anything",
			slice: []telemetryapi.Event{
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
			expectedLogRecords:        1,
			expectedType:              "extension",
			expectedTimestamp:         "2022-10-12T00:03:50.000Z",
			expectedBody:              "Hello world, I am an extension!",
			expectedContainsRequestId: true,
			expectedRequestId:         "79b4f56e-95b1-4643-9700-2807f4e68689",
			expectedSeverityText:      "Unspecified",
			expectedSeverityNumber:    plog.SeverityNumberUnspecified,
			expectError:               false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			r := newTelemetryAPIReceiver(
				&Config{},
				receivertest.NewNopCreateSettings(),
			)
			log, err := r.createLogs(tc.slice)
			if tc.expectError {
				require.Error(t, err)
			} else {
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
					expectedTime, err := time.Parse(time.RFC3339, tc.expectedTimestamp)
					require.NoError(t, err)
					require.Equal(t, pcommon.NewTimestampFromTime(expectedTime), logRecord.Timestamp())
					requestId, ok := logRecord.Attributes().Get(semconv.AttributeFaaSInvocationID)
					require.Equal(t, tc.expectedContainsRequestId, ok)
					if ok {
						require.Equal(t, tc.expectedRequestId, requestId.Str())
					}
					require.Equal(t, tc.expectedSeverityText, logRecord.SeverityText())
					require.Equal(t, tc.expectedSeverityNumber, logRecord.SeverityNumber())
					require.Equal(t, tc.expectedBody, logRecord.Body().Str())
				}
			}
		})
	}
}

func TestSeverityTextToNumber(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		level  string
		number plog.SeverityNumber
	}{
		{
			level:  "TRACE",
			number: plog.SeverityNumberTrace,
		},
		{
			level:  "TRACE2",
			number: plog.SeverityNumberTrace2,
		},
		{
			level:  "TRACE3",
			number: plog.SeverityNumberTrace3,
		},
		{
			level:  "TRACE4",
			number: plog.SeverityNumberTrace4,
		},
		{
			level:  "DEBUG2",
			number: plog.SeverityNumberDebug2,
		},
		{
			level:  "DEBUG3",
			number: plog.SeverityNumberDebug3,
		},
		{
			level:  "DEBUG4",
			number: plog.SeverityNumberDebug4,
		},
		{
			level:  "INFO",
			number: plog.SeverityNumberInfo,
		},
		{
			level:  "INFO2",
			number: plog.SeverityNumberInfo2,
		},
		{
			level:  "INFO3",
			number: plog.SeverityNumberInfo3,
		},
		{
			level:  "INFO4",
			number: plog.SeverityNumberInfo4,
		},
		{
			level:  "WARN",
			number: plog.SeverityNumberWarn,
		},
		{
			level:  "WARN2",
			number: plog.SeverityNumberWarn2,
		},
		{
			level:  "WARN3",
			number: plog.SeverityNumberWarn3,
		},
		{
			level:  "WARN4",
			number: plog.SeverityNumberWarn4,
		},
		{
			level:  "ERROR",
			number: plog.SeverityNumberError,
		},
		{
			level:  "ERROR2",
			number: plog.SeverityNumberError2,
		},
		{
			level:  "ERROR3",
			number: plog.SeverityNumberError3,
		},
		{
			level:  "ERROR4",
			number: plog.SeverityNumberError4,
		},
		{
			level:  "FATAL",
			number: plog.SeverityNumberFatal,
		},
		{
			level:  "FATAL2",
			number: plog.SeverityNumberFatal2,
		},
		{
			level:  "FATAL3",
			number: plog.SeverityNumberFatal3,
		},
		{
			level:  "FATAL4",
			number: plog.SeverityNumberFatal4,
		},
		{
			level:  "CRITICAL",
			number: plog.SeverityNumberFatal,
		},
		{
			level:  "ALL",
			number: plog.SeverityNumberTrace,
		},
		{
			level:  "WARNING",
			number: plog.SeverityNumberWarn,
		},
		{
			level:  "UNKNOWN",
			number: plog.SeverityNumberUnspecified,
		},
	}
	for _, tc := range testCases {
		require.Equal(t, tc.number, severityTextToNumber(tc.level))

	}
}

func TestParseTimestamp(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		timestamp string
		expected  time.Time
	}{
		{
			timestamp: "2024-07-05T21:12:37Z",
			expected:  time.Date(2024, time.July, 5, 21, 12, 37, 0, time.UTC),
		},
		{
			timestamp: "2024-07-09T10:53:34.689Z",
			expected:  time.Date(2024, time.July, 9, 10, 53, 34, 689*1000*1000, time.UTC),
		},
	}
	for _, tc := range testCases {
		parsed, err := parseTimestamp(tc.timestamp)
		require.NoError(t, err)
		require.Equal(t, tc.expected, parsed)
	}
}
