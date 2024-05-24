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
)

func TestListenOnAddress(t *testing.T) {
	testCases := []struct {
		desc     string
		testFunc func(*testing.T)
	}{
		{
			desc: "listen on address without AWS_SAM_LOCAL env variable",
			testFunc: func(t *testing.T) {
				addr := listenOnAddress()
				require.EqualValues(t, "sandbox.localdomain:4325", addr)
			},
		},
		{
			desc: "listen on address with AWS_SAM_LOCAL env variable",
			testFunc: func(t *testing.T) {
				t.Setenv("AWS_SAM_LOCAL", "true")
				addr := listenOnAddress()
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

func (c *mockConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func TestHandler(t *testing.T) {
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
			r.registerTracesConsumer(consumer)
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
	testCases := []struct {
		desc                      string
		slice                     []event
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
			expectedLogRecords:        1,
			expectedType:              "function",
			expectedTimestamp:         "2022-10-12T00:03:50.000Z",
			expectedBody:              "Hello world, I am a function!",
			expectedContainsRequestId: true,
			expectedRequestId:         "79b4f56e-95b1-4643-9700-2807f4e68189",
			expectedSeverityText:      "INFO",
			expectedSeverityNumber:    plog.SeverityNumberInfo,
			expectError:               false,
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
			expectedLogRecords:        1,
			expectedType:              "extension",
			expectedTimestamp:         "2022-10-12T00:03:50.000Z",
			expectedBody:              "Hello world, I am an extension!",
			expectedContainsRequestId: true,
			expectedRequestId:         "79b4f56e-95b1-4643-9700-2807f4e68689",
			expectedSeverityText:      "INFO",
			expectedSeverityNumber:    plog.SeverityNumberInfo,
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
					expectedTime, err := time.Parse(timeFormatLayout, tc.expectedTimestamp)
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
