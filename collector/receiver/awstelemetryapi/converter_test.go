package awstelemetryapi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv "go.opentelemetry.io/collector/semconv/v1.25.0"
)

// This test is adapted from your original receiver_test.go to test the converter function directly.
func TestCreateLogs(t *testing.T) {
	r := &telemetryAPIReceiver{resource: pcommon.NewResource()}
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

func TestCreateMetrics(t *testing.T) {
	r := &telemetryAPIReceiver{resource: pcommon.NewResource()}
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

func TestCreateInvokeSpan(t *testing.T) {
	r := &telemetryAPIReceiver{resource: pcommon.NewResource()}
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
