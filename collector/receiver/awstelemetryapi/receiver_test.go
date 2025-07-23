package awstelemetryapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"
)

func TestHttpHandler_Metrics(t *testing.T) {
	// Setup: Create a mock consumer (a "sink") to receive the metrics
	sink := new(consumertest.MetricsSink)

	// Use the corrected function name: NewNopSettings
	r, err := newTelemetryAPIReceiver(&Config{}, receivertest.NewNopSettings(component.MustNewType(typeStr)))
	require.NoError(t, err)
	r.registerMetricsConsumer(sink)

	// Create a sample HTTP request with a platform.report event
	sampleRecord := map[string]interface{}{
		"requestId": "test-req-id",
		"metrics":   map[string]interface{}{"durationMs": 100.0},
	}
	payload, _ := json.Marshal([]event{
		{Time: time.Now().Format(time.RFC3339), Type: "platform.report", Record: sampleRecord},
	})

	req, _ := http.NewRequest("POST", "/", bytes.NewReader(payload))
	rr := httptest.NewRecorder()

	// Execute the httpHandler
	r.httpHandler(rr, req)

	// Assert the results
	require.Equal(t, http.StatusOK, rr.Code)

	// Use the correct method to check the count: len(sink.AllMetrics())
	require.Len(t, sink.AllMetrics(), 1, "sink should have received one metric payload")
	allMetrics := sink.AllMetrics()[0]
	numDataPoints := allMetrics.MetricCount()
	require.Equal(t, 1, numDataPoints, "payload should contain one metric")
}

func TestHttpHandler_Logs(t *testing.T) {
	// Setup: Create a sink for logs
	sink := new(consumertest.LogsSink)

	// Use the corrected function name: NewNopSettings
	r, err := newTelemetryAPIReceiver(&Config{}, receivertest.NewNopSettings(component.MustNewType(typeStr)))
	require.NoError(t, err)
	r.registerLogsConsumer(sink)

	// Create a sample HTTP request with a function event
	payload, _ := json.Marshal([]event{
		{Time: time.Now().Format(time.RFC3339), Type: "function", Record: "hello world"},
	})

	req, _ := http.NewRequest("POST", "/", bytes.NewReader(payload))
	rr := httptest.NewRecorder()

	// Execute
	r.httpHandler(rr, req)

	// Assert
	require.Equal(t, http.StatusOK, rr.Code)
	// Use the correct method to check the count
	require.Len(t, sink.AllLogs(), 1, "sink should have received one log payload")
	require.Equal(t, 1, sink.AllLogs()[0].LogRecordCount(), "payload should contain one log record")
}
