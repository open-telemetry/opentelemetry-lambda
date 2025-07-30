package telemetryapireceiver

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv "go.opentelemetry.io/collector/semconv/v1.25.0"
	"go.uber.org/zap"
)

// TestCreateLogs tests the basic functionality of log creation
func TestCreateLogs(t *testing.T) {
	r := &telemetryAPIReceiver{
		resource: pcommon.NewResource(),
		logger:   zap.NewNop(),
	}
	sampleEvent := event{
		Time: "2022-10-12T00:03:50.000Z",
		Type: "function",
		Record: map[string]interface{}{
			"timestamp": "2022-10-12T00:03:50.000Z",
			"level":     "INFO",
			"requestId": "test-req-id-123",
			"message":   "Hello world!",
		},
	}

	logs, err := r.createLogs(sampleEvent)
	require.NoError(t, err)
	require.Equal(t, 1, logs.LogRecordCount())

	logRecord := logs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
	require.Equal(t, "Hello world!", logRecord.Body().Str())
	require.Equal(t, plog.SeverityNumberInfo, logRecord.SeverityNumber())
	val, _ := logRecord.Attributes().Get(semconv.AttributeFaaSInvocationID)
	require.Equal(t, "test-req-id-123", val.Str())
}

func TestCreateLogs_WithTraceAndSpanID(t *testing.T) {
	r := &telemetryAPIReceiver{
		resource: pcommon.NewResource(),
		logger:   zap.NewNop(),
	}

	// Test with valid trace and span IDs (32 hex chars for trace, 16 hex chars for span)
	sampleEvent := event{
		Time: "2022-10-12T00:03:50.000Z",
		Type: "function",
		Record: map[string]interface{}{
			"timestamp": "2022-10-12T00:03:50.000Z",
			"level":     "INFO",
			"requestId": "test-req-id-123",
			"message":   "Hello world with trace context!",
			"trace_id":  "80e1afed08e019fc1110464cfa66635c", // 32 hex chars
			"span_id":   "7a085853722dc6d2",                 // 16 hex chars
		},
	}

	logs, err := r.createLogs(sampleEvent)
	require.NoError(t, err)
	require.Equal(t, 1, logs.LogRecordCount())

	logRecord := logs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
	require.Equal(t, "Hello world with trace context!", logRecord.Body().Str())

	require.NotEqual(t, pcommon.NewTraceIDEmpty(), logRecord.TraceID())
	require.NotEqual(t, pcommon.NewSpanIDEmpty(), logRecord.SpanID())

	val, _ := logRecord.Attributes().Get(semconv.AttributeFaaSInvocationID)
	require.Equal(t, "test-req-id-123", val.Str())
}

// TestCreateLogs_EdgeCases tests various edge cases and error scenarios
func TestCreateLogs_EdgeCases(t *testing.T) {
	r := &telemetryAPIReceiver{
		resource: pcommon.NewResource(),
		logger:   zap.NewNop(),
	}

	tests := []struct {
		name     string
		event    event
		wantBody string
	}{
		{
			name: "plain string record",
			event: event{
				Time:   "2022-10-12T00:03:50.000Z",
				Type:   "function",
				Record: "plain string log message",
			},
			wantBody: "plain string log message",
		},
		{
			name: "empty message",
			event: event{
				Time: "2022-10-12T00:03:50.000Z",
				Type: "function",
				Record: map[string]interface{}{
					"message": "",
				},
			},
			wantBody: "",
		},
		{
			name: "missing message field",
			event: event{
				Time: "2022-10-12T00:03:50.000Z",
				Type: "function",
				Record: map[string]interface{}{
					"level": "INFO",
				},
			},
			wantBody: "",
		},
		{
			name: "invalid trace_id - wrong length",
			event: event{
				Time: "2022-10-12T00:03:50.000Z",
				Type: "function",
				Record: map[string]interface{}{
					"message":  "test message",
					"trace_id": "80e1afed", // too short
				},
			},
			wantBody: "test message",
		},
		{
			name: "invalid span_id - wrong format",
			event: event{
				Time: "2022-10-12T00:03:50.000Z",
				Type: "function",
				Record: map[string]interface{}{
					"message": "test message",
					"span_id": "invalid-span-id",
				},
			},
			wantBody: "test message",
		},
		{
			name: "invalid timestamp format",
			event: event{
				Time: "2022-10-12T00:03:50.000Z",
				Type: "function",
				Record: map[string]interface{}{
					"timestamp": "invalid-timestamp",
					"message":   "test message",
				},
			},
			wantBody: "test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs, err := r.createLogs(tt.event)
			require.NoError(t, err)
			require.Equal(t, 1, logs.LogRecordCount())

			logRecord := logs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
			assert.Equal(t, tt.wantBody, logRecord.Body().Str())
		})
	}
}

// TestCreateLogs_AllSeverityLevels tests all possible severity level mappings
func TestCreateLogs_AllSeverityLevels(t *testing.T) {
	r := &telemetryAPIReceiver{
		resource: pcommon.NewResource(),
		logger:   zap.NewNop(),
	}

	tests := []struct {
		level    string
		expected plog.SeverityNumber
	}{
		{"TRACE", plog.SeverityNumberTrace},
		{"DEBUG", plog.SeverityNumberDebug},
		{"INFO", plog.SeverityNumberInfo},
		{"WARN", plog.SeverityNumberWarn},
		{"WARNING", plog.SeverityNumberWarn},
		{"ERROR", plog.SeverityNumberError},
		{"FATAL", plog.SeverityNumberFatal},
		{"CRITICAL", plog.SeverityNumberFatal},
		{"unknown", plog.SeverityNumberUnspecified},
		{"", plog.SeverityNumberUnspecified},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			event := event{
				Time: "2022-10-12T00:03:50.000Z",
				Type: "function",
				Record: map[string]interface{}{
					"level":   tt.level,
					"message": "test message",
				},
			}

			logs, err := r.createLogs(event)
			require.NoError(t, err)

			logRecord := logs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
			assert.Equal(t, tt.expected, logRecord.SeverityNumber())
		})
	}
}

// TestCreateMetrics tests basic metrics creation
func TestCreateMetrics(t *testing.T) {
	r := &telemetryAPIReceiver{
		resource: pcommon.NewResource(),
		logger:   zap.NewNop(),
	}
	sampleRecord := map[string]interface{}{
		"requestId": "test-req-id-456",
		"metrics": map[string]interface{}{
			"durationMs": 150.5,
		},
	}
	sampleEvent := event{
		Time:   time.Now().Format(time.RFC3339),
		Type:   "platform.report",
		Record: sampleRecord,
	}

	metrics, err := r.createMetrics(sampleEvent)
	require.NoError(t, err)
	require.Equal(t, 1, metrics.MetricCount())

	metric := metrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
	dp := metric.Gauge().DataPoints().At(0)
	assert.Equal(t, 150.5, dp.DoubleValue())
	val, _ := dp.Attributes().Get(semconv.AttributeFaaSInvocationID)
	assert.Equal(t, "test-req-id-456", val.Str())
}

// TestCreateMetrics_AllMetricTypes tests all supported metric types and units
func TestCreateMetrics_AllMetricTypes(t *testing.T) {
	r := &telemetryAPIReceiver{
		resource: pcommon.NewResource(),
		logger:   zap.NewNop(),
	}

	tests := []struct {
		name         string
		metricKey    string
		value        float64
		expectedUnit string
		expectInt    bool
		expectedVal  interface{}
	}{
		{"duration", "durationMs", 150.5, "ms", false, 150.5},
		{"billed duration", "billedDurationMs", 200.0, "ms", false, 200.0},
		{"init duration", "initDurationMs", 1000.0, "ms", false, 1000.0},
		{"restore duration", "restoreDurationMs", 500.0, "ms", false, 500.0},
		{"memory size", "memorySizeMB", 128.0, "By", true, int64(128 * 1024 * 1024)},
		{"max memory used", "maxMemoryUsedMB", 64.0, "By", true, int64(64 * 1024 * 1024)},
		{"produced bytes", "producedBytes", 1024.0, "By", true, int64(1024)},
		{"custom metric", "customMetric", 42.0, "1", false, 42.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sampleRecord := map[string]interface{}{
				"requestId": "test-req-id",
				"metrics": map[string]interface{}{
					tt.metricKey: tt.value,
				},
			}
			event := event{
				Time:   time.Now().Format(time.RFC3339),
				Type:   "platform.report",
				Record: sampleRecord,
			}

			metrics, err := r.createMetrics(event)
			require.NoError(t, err)
			require.Equal(t, 1, metrics.MetricCount())

			metric := metrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
			assert.Equal(t, fmt.Sprintf("aws.lambda.%s", tt.metricKey), metric.Name())
			assert.Equal(t, tt.expectedUnit, metric.Unit())

			dp := metric.Gauge().DataPoints().At(0)
			if tt.expectInt {
				assert.Equal(t, tt.expectedVal, dp.IntValue())
			} else {
				assert.Equal(t, tt.expectedVal, dp.DoubleValue())
			}
		})
	}
}

// TestCreateMetrics_ErrorCases tests error scenarios for metrics creation
func TestCreateMetrics_ErrorCases(t *testing.T) {
	r := &telemetryAPIReceiver{
		resource: pcommon.NewResource(),
		logger:   zap.NewNop(),
	}

	tests := []struct {
		name        string
		record      interface{}
		expectError bool
	}{
		{
			name:        "non-map record",
			record:      "not a map",
			expectError: true,
		},
		{
			name: "missing metrics field",
			record: map[string]interface{}{
				"requestId": "test-req-id",
			},
			expectError: true,
		},
		{
			name: "metrics field not a map",
			record: map[string]interface{}{
				"requestId": "test-req-id",
				"metrics":   "not a map",
			},
			expectError: true,
		},
		{
			name: "empty metrics",
			record: map[string]interface{}{
				"requestId": "test-req-id",
				"metrics":   map[string]interface{}{},
			},
			expectError: false, // Should succeed but create no metrics
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := event{
				Time:   time.Now().Format(time.RFC3339),
				Type:   "platform.report",
				Record: tt.record,
			}

			metrics, err := r.createMetrics(event)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.name == "empty metrics" {
					assert.Equal(t, 0, metrics.MetricCount())
				}
			}
		})
	}
}

// TestCreateInvokeSpan tests invoke span creation
func TestCreateInvokeSpan(t *testing.T) {
	r := &telemetryAPIReceiver{
		resource: pcommon.NewResource(),
		logger:   zap.NewNop(),
	}
	startTime := time.Now()
	state := invocationState{start: startTime}
	sampleRecord := map[string]interface{}{
		"requestId": "test-req-id-789",
		"status":    "error",
		"errorType": "Runtime.ExitError",
	}
	endEvent := event{
		Time:   startTime.Add(100 * time.Millisecond).Format(time.RFC3339),
		Type:   "platform.runtimeDone",
		Record: sampleRecord,
	}

	traces, err := r.createInvokeSpan(endEvent, state)
	require.NoError(t, err)
	require.Equal(t, 1, traces.SpanCount())

	span := traces.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0)
	assert.Equal(t, "platform.invoke", span.Name())
	assert.Equal(t, pcommon.NewTimestampFromTime(startTime), span.StartTimestamp())
	assert.Equal(t, ptrace.StatusCodeError, span.Status().Code())
	assert.Equal(t, "Runtime.ExitError", span.Status().Message())
	val, _ := span.Attributes().Get(semconv.AttributeFaaSInvocationID)
	assert.Equal(t, "test-req-id-789", val.Str())
}

// TestCreateInitSpan tests init span creation
func TestCreateInitSpan(t *testing.T) {
	r := &telemetryAPIReceiver{
		resource:      pcommon.NewResource(),
		logger:        zap.NewNop(),
		initStartTime: time.Now(),
	}

	tests := []struct {
		name           string
		record         interface{}
		expectedStatus ptrace.StatusCode
	}{
		{
			name: "successful init",
			record: map[string]interface{}{
				"status": "success",
			},
			expectedStatus: ptrace.StatusCodeUnset,
		},
		{
			name: "failed init",
			record: map[string]interface{}{
				"status":    "error",
				"errorType": "Runtime.InitError",
			},
			expectedStatus: ptrace.StatusCodeError,
		},
		{
			name:           "non-map record",
			record:         "not a map",
			expectedStatus: ptrace.StatusCodeUnset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := event{
				Time:   r.initStartTime.Add(2 * time.Second).Format(time.RFC3339),
				Type:   "platform.initRuntimeDone",
				Record: tt.record,
			}

			traces, err := r.createInitSpan(event)
			require.NoError(t, err)
			require.Equal(t, 1, traces.SpanCount())

			span := traces.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0)
			assert.Equal(t, "platform.init", span.Name())
			assert.Equal(t, ptrace.SpanKindInternal, span.Kind())
			assert.Equal(t, tt.expectedStatus, span.Status().Code())

			// Check coldstart attribute
			val, exists := span.Attributes().Get(semconv.AttributeFaaSColdstart)
			assert.True(t, exists)
			assert.True(t, val.Bool())
		})
	}
}

// TestSeverityTextToNumber tests the severity mapping function
func TestSeverityTextToNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected plog.SeverityNumber
	}{
		{"TRACE", plog.SeverityNumberTrace},
		{"trace", plog.SeverityNumberTrace}, // case insensitive
		{"DEBUG", plog.SeverityNumberDebug},
		{"INFO", plog.SeverityNumberInfo},
		{"WARN", plog.SeverityNumberWarn},
		{"WARNING", plog.SeverityNumberWarn},
		{"ERROR", plog.SeverityNumberError},
		{"FATAL", plog.SeverityNumberFatal},
		{"CRITICAL", plog.SeverityNumberFatal},
		{"unknown", plog.SeverityNumberUnspecified},
		{"", plog.SeverityNumberUnspecified},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := severityTextToNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSpanStatusMapping tests the setSpanStatus helper function
func TestSpanStatusMapping(t *testing.T) {
	traces := ptrace.NewTraces()
	span := traces.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()

	tests := []struct {
		name           string
		record         map[string]interface{}
		expectedStatus ptrace.StatusCode
		expectedMsg    string
	}{
		{
			name: "success status",
			record: map[string]interface{}{
				"status": "success",
			},
			expectedStatus: ptrace.StatusCodeUnset,
			expectedMsg:    "",
		},
		{
			name: "error status with error type",
			record: map[string]interface{}{
				"status":    "error",
				"errorType": "Runtime.ExitError",
			},
			expectedStatus: ptrace.StatusCodeError,
			expectedMsg:    "Runtime.ExitError",
		},
		{
			name: "error status without error type",
			record: map[string]interface{}{
				"status": "failure",
			},
			expectedStatus: ptrace.StatusCodeError,
			expectedMsg:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset span for each test
			span = traces.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0)
			span.Status().SetCode(ptrace.StatusCodeUnset)
			span.Status().SetMessage("")

			setSpanStatus(span, tt.record)

			assert.Equal(t, tt.expectedStatus, span.Status().Code())
			assert.Equal(t, tt.expectedMsg, span.Status().Message())
		})
	}
}
