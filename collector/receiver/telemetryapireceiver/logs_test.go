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
	"go.opentelemetry.io/collector/receiver/receivertest"
	semconv "go.opentelemetry.io/collector/semconv/v1.25.0"
)

func TestListenOnLogsAddress(t *testing.T) {
	testCases := []struct {
		desc     string
		testFunc func(*testing.T)
	}{
		{
			desc: "listen on address without AWS_SAM_LOCAL env variable",
			testFunc: func(t *testing.T) {
				addr := listenOnLogsAddress()
				require.EqualValues(t, "sandbox.localdomain:4327", addr)
			},
		},
		{
			desc: "listen on address with AWS_SAM_LOCAL env variable",
			testFunc: func(t *testing.T) {
				t.Setenv("AWS_SAM_LOCAL", "true")
				addr := listenOnLogsAddress()
				require.EqualValues(t, ":4327", addr)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, tc.testFunc)
	}
}

type mockLogsConsumer struct {
	consumed int
}

func (c *mockLogsConsumer) ConsumeLogs(ctx context.Context, td plog.Logs) error {
	c.consumed += td.LogRecordCount()
	return nil
}

func (c *mockLogsConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func TestLogsHandler(t *testing.T) {
	testCases := []struct {
		desc         string
		body         string
		expectedLogs int
	}{
		{
			desc:         "empty body",
			body:         `{}`,
			expectedLogs: 0,
		},
		{
			desc:         "invalid json",
			body:         `invalid json`,
			expectedLogs: 0,
		},
		{
			desc:         "valid event but no time",
			body:         `[{"time":"", "type":"", "record": {}}]`,
			expectedLogs: 0,
		},
		{
			desc: "platform.initStart",
			body: `[
				{
					"time": "2024-05-15T18:10:29.635Z",
					"type": "platform.initStart",
					"record": {
						"functionName": "opentelemetry-lambda-nodejs-experimental-arm64",
						"functionVersion": "$LATEST",
						"initializationType": "on-demand",
						"phase": "init",
						"runtimeVersion": "nodejs:20.v22",
						"runtimeVersionArn": "arn:aws:lambda:us-east-1::runtime:da57c20c4b965d5b75540f6865a35fc8030358e33ec44ecfed33e90901a27a72"
					}
				}
			]`,
			expectedLogs: 1,
		},
		{
			desc: "platform.telemetrySubscription",
			body: `[
				{
					"time": "2024-05-15T18:10:30.010Z",
					"type": "platform.telemetrySubscription",
					"record": {
						"name": "collector",
						"state": "Subscribed",
						"types": [
							"platform"
						]
					}
				}
			]`,
			expectedLogs: 1,
		},
		{
			desc: "platform.telemetrySubscription",
			body: `[
				{
					"time": "2024-05-15T18:10:30.511Z",
					"type": "platform.telemetrySubscription",
					"record": {
						"name": "collector",
						"state": "Subscribed",
						"types": [
							"platform",
							"function"
						]
					}
				}
			]`,
			expectedLogs: 1,
		},
		{
			desc: "platform.initRuntimeDone",
			body: `[
				{
					"time": "2024-05-15T23:58:26.857Z",
					"type": "platform.initRuntimeDone",
					"record": {
						"initializationType": "on-demand",
						"phase": "init",
						"status": "success"
					}
				}
			]`,
			expectedLogs: 1,
		},
		{
			desc: "platform.extension",
			body: `[
				{
					"time": "2024-05-15T23:58:26.857Z",
					"type": "platform.extension",
					"record": {
						"events": [
							"INVOKE",
							"SHUTDOWN"
						],
						"name": "collector",
						"state": "Ready"
					}
				}
			]`,
			expectedLogs: 1,
		},
		{
			desc: "platform.initReport",
			body: `[
				{
					"time": "2024-05-15T23:58:26.858Z",
					"type": "platform.initReport",
					"record": {
						"initializationType": "on-demand",
						"metrics": {
							"durationMs": 1819.081
						},
						"phase": "init",
						"status": "success"
					}
				}
			]`,
			expectedLogs: 1,
		},
		{
			desc: "platform.runtimeDone",
			body: `[
				{
					"time": "2024-05-15T23:58:35.063Z",
					"type": "platform.runtimeDone",
					"record": {
						"metrics": {
							"durationMs": 8202.659,
							"producedBytes": 50
						},
						"requestId": "882e9658-570e-4b2f-aaa8-5dfb88f7eccb",
						"spans": [
							{
								"durationMs": 8200.68,
								"name": "responseLatency",
								"start": "2024-05-15T23:58:26.860Z"
							},
							{
								"durationMs": 0.226,
								"name": "responseDuration",
								"start": "2024-05-15T23:58:35.061Z"
							},
							{
								"durationMs": 1.502,
								"name": "runtimeOverhead",
								"start": "2024-05-15T23:58:35.061Z"
							}
						],
						"status": "success"
					}
				}
			]`,
			expectedLogs: 1,
		},
		{
			desc: "platform.report",
			body: `[
				{
					"time": "2024-05-15T23:58:39.317Z",
					"type": "platform.report",
					"record": {
						"metrics": {
							"billedDurationMs": 12456,
							"durationMs": 12455.155,
							"initDurationMs": 1819.881,
							"maxMemoryUsedMB": 128,
							"memorySizeMB": 128
						},
						"requestId": "882e9658-570e-4b2f-aaa8-5dfb88f7eccb",
						"status": "success"
					}
				}
			]`,
			expectedLogs: 1,
		},
		{
			desc: "platform.start",
			body: `[
				{
					"time": "2024-05-15T23:58:41.153Z",
					"type": "platform.start",
					"record": {
						"requestId": "15b0ebbb-5cf8-49e2-8cbe-1d58a18330d2",
						"version": "$LATEST"
					}
				}
			]`,
			expectedLogs: 1,
		},
		{
			desc: "platform.restoreStart",
			body: `[
				{
					"time": "2022-10-12T00:00:15.064Z",
					"type": "platform.restoreStart",
					"record": {
						"runtimeVersion": "nodejs-14.v3",
						"runtimeVersionArn": "arn",
						"functionName": "myFunction",
						"functionVersion": "$LATEST",
						"instanceId": "82561ce0-53dd-47d1-90e0-c8f5e063e62e",
						"instanceMaxMemory": 256
					}
				}
			]`,
			expectedLogs: 1,
		},
		{
			desc: "platform.restoreRuntimeDone",
			body: `[
				{
					"time": "2022-10-12T00:00:15.064Z",
					"type": "platform.restoreRuntimeDone",
					"record": {
						"status": "success",
						"spans": [
							{
								"name": "someTimeSpan",
								"start": "2022-08-02T12:01:23:521Z",
								"durationMs": 80.0
							}
						]
					}
				}
			]`,
			expectedLogs: 1,
		},
		{
			desc: "platform.restoreReport",
			body: `[
				{
					"time": "2022-10-12T00:00:15.064Z",
					"type": "platform.restoreReport",
					"record": {
						"status": "success",
						"metrics": {
							"durationMs": 15.19
						},
						"spans": [
							{
								"name": "someTimeSpan",
								"start": "2022-08-02T12:01:23:521Z",
								"durationMs": 30.0
							}
						]
					}
				}
			]`,
			expectedLogs: 1,
		},
		{
			desc: "platform.logsDropped",
			body: `[
				{
					"time": "2022-10-12T00:02:35.000Z",
					"type": "platform.logsDropped",
					"record": {
						"droppedBytes": 12345,
						"droppedRecords": 123,
						"reason": "Some logs were dropped because the downstream consumer is slower than the logs production rate"
					}
				}
			]`,
			expectedLogs: 1,
		},
		{
			desc: "function",
			body: `[
				{
					"time": "2024-05-15T23:59:20.159Z",
					"type": "function",
					"record": "2024-05-15T23:59:20.159Z\t8c181f94-d34c-4c65-abed-6977e17dd06b\tWARN\twarn from console\n"
				},
				{
					"time": "2022-10-12T00:03:50.000Z",
					"type": "function",
					"record": {
						"timestamp": "2022-10-12T00:03:50.000Z",
						"level": "INFO",
						"requestId": "79b4f56e-95b1-4643-9700-2807f4e68189",
						"message": "Hello world, I am a function!"
					}
				}
			]`,
			expectedLogs: 2,
		},
		{
			desc: "extension",
			body: `[
				{
					"time": "2022-10-12T00:03:50.000Z",
					"type": "extension",
					"record": "[INFO] Hello world, I am an extension!"
				},
				{
					"time": "2022-10-12T00:03:50.000Z",
					"type": "extension",
					"record": {
					   "timestamp": "2022-10-12T00:03:50.000Z",
					   "level": "INFO",
					   "requestId": "79b4f56e-95b1-4643-9700-2807f4e68189",
					   "message": "Hello world, I am an extension!"
					}
				}
			]`,
			expectedLogs: 2,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			consumer := mockLogsConsumer{}
			r, err := newTelemetryAPILogsReceiver(
				&Config{},
				&consumer,
				receivertest.NewNopCreateSettings(),
			)
			require.NoError(t, err)
			req := httptest.NewRequest("POST",
				"http://localhost:53612/someevent", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			r.httpHandler(rec, req)
			require.Equal(t, tc.expectedLogs, consumer.consumed)
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
		{
			desc: "platform.initReport",
			slice: []event{
				{
					Time: "2024-05-15T23:58:26.858Z",
					Type: "platform.initReport",
					Record: map[string]any{
						"initializationType": "on-demand",
						"metrics": map[string]any{
							"durationMs": 1819.081,
						},
						"phase":  "init",
						"status": "success",
					},
				},
			},
			expectedLogRecords:        1,
			expectedType:              "platform.initReport",
			expectedTimestamp:         "2024-05-15T23:58:26.858Z",
			expectedBody:              "{\"initializationType\":\"on-demand\",\"metrics\":{\"durationMs\":1819.081},\"phase\":\"init\",\"status\":\"success\"}",
			expectedContainsRequestId: false,
			expectedSeverityText:      "INFO",
			expectedSeverityNumber:    plog.SeverityNumberInfo,
			expectError:               false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			r, err := newTelemetryAPILogsReceiver(
				&Config{},
				nil,
				receivertest.NewNopCreateSettings(),
			)
			require.NoError(t, err)
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
