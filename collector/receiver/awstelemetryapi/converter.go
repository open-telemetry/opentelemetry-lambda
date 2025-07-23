package awstelemetryapi

import (
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv "go.opentelemetry.io/collector/semconv/v1.25.0"
)

// createLogs converts a "function" or "extension" event into plog.Logs.
func (r *telemetryAPIReceiver) createLogs(e event) (plog.Logs, error) {
	logs := plog.NewLogs()
	resourceLog := logs.ResourceLogs().AppendEmpty()
	r.resource.CopyTo(resourceLog.Resource())
	scopeLog := resourceLog.ScopeLogs().AppendEmpty()
	scopeLog.Scope().SetName(scopeName)

	logRecord := scopeLog.LogRecords().AppendEmpty()
	logRecord.Attributes().PutStr("lambda.event.type", e.Type)
	logRecord.SetTimestamp(pcommon.NewTimestampFromTime(e.getTime()))
	logRecord.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))

	// This logic correctly handles both JSON-structured and plain-text function logs.
	if record, ok := e.Record.(map[string]interface{}); ok {
		if timestamp, ok := record["timestamp"].(string); ok {
			if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
				logRecord.SetTimestamp(pcommon.NewTimestampFromTime(t))
			}
		}
		if level, ok := record["level"].(string); ok {
			logRecord.SetSeverityNumber(severityTextToNumber(level))
			logRecord.SetSeverityText(logRecord.SeverityNumber().String())
		}
		if reqID, ok := record["requestId"].(string); ok {
			logRecord.Attributes().PutStr(semconv.AttributeFaaSInvocationID, reqID)
		}
		if msg, ok := record["message"].(string); ok {
			logRecord.Body().SetStr(msg)
		}
	} else if line, ok := e.Record.(string); ok {
		logRecord.Body().SetStr(line)
	}

	return logs, nil
}

// createMetrics converts a "platform.report" event into pmetric.Metrics.
func (r *telemetryAPIReceiver) createMetrics(e event) (pmetric.Metrics, error) {
	metrics := pmetric.NewMetrics()
	resourceMetrics := metrics.ResourceMetrics().AppendEmpty()
	r.resource.CopyTo(resourceMetrics.Resource())
	scopeMetrics := resourceMetrics.ScopeMetrics().AppendEmpty()
	scopeMetrics.Scope().SetName(scopeName)

	record, ok := e.Record.(map[string]interface{})
	if !ok {
		return metrics, fmt.Errorf("metric event record is not a map")
	}

	// The platform.report event contains a 'metrics' object with key-value pairs.
	metricData, ok := record["metrics"].(map[string]interface{})
	if !ok {
		return metrics, fmt.Errorf("metrics field not found in record")
	}

	// It also contains the 'requestId' for the invocation.
	reqID, _ := record["requestId"].(string)

	ts := pcommon.NewTimestampFromTime(e.getTime())

	for key, value := range metricData {
		if val, ok := value.(float64); ok {
			// Create a new metric for each key-value pair.
			m := scopeMetrics.Metrics().AppendEmpty()
			m.SetName(fmt.Sprintf("aws.lambda.%s", key))
			m.SetUnit("1")

			dp := m.SetEmptyGauge().DataPoints().AppendEmpty()
			dp.SetTimestamp(ts)
			dp.SetDoubleValue(val)
			if reqID != "" {
				dp.Attributes().PutStr(semconv.AttributeFaaSInvocationID, reqID)
			}
		}
	}
	return metrics, nil
}

// createInitSpan creates a trace span for the Lambda init phase.
func (r *telemetryAPIReceiver) createInitSpan(e event) (ptrace.Traces, error) {
	traces := ptrace.NewTraces()
	rs := traces.ResourceSpans().AppendEmpty()
	r.resource.CopyTo(rs.Resource())
	span := rs.ScopeSpans().AppendEmpty().Spans().AppendEmpty()

	span.SetName("platform.init")
	span.SetKind(ptrace.SpanKindInternal)
	span.Attributes().PutBool(semconv.AttributeFaaSColdstart, true)
	span.SetStartTimestamp(pcommon.NewTimestampFromTime(r.initStartTime))
	span.SetEndTimestamp(pcommon.NewTimestampFromTime(e.getTime()))

	// Add attributes from the event record for more context.
	if record, ok := e.Record.(map[string]interface{}); ok {
		if status, ok := record["status"].(string); ok {
			span.Attributes().PutStr("aws.lambda.init.status", status)
			if status != "success" {
				span.Status().SetCode(ptrace.StatusCodeError)
				if errorType, ok := record["errorType"].(string); ok {
					span.Status().SetMessage(errorType)
				}
			}
		}
	}
	return traces, nil
}

// createInvokeSpan creates a trace span for the Lambda invoke phase.
func (r *telemetryAPIReceiver) createInvokeSpan(e event, state invocationState) (ptrace.Traces, error) {
	traces := ptrace.NewTraces()
	rs := traces.ResourceSpans().AppendEmpty()
	r.resource.CopyTo(rs.Resource())
	span := rs.ScopeSpans().AppendEmpty().Spans().AppendEmpty()

	span.SetName("platform.invoke")
	span.SetKind(ptrace.SpanKindServer)
	span.SetStartTimestamp(pcommon.NewTimestampFromTime(state.start))
	span.SetEndTimestamp(pcommon.NewTimestampFromTime(e.getTime()))

	// Add attributes from the event record for more context.
	if record, ok := e.Record.(map[string]interface{}); ok {
		if reqID, ok := record["requestId"].(string); ok {
			span.Attributes().PutStr(semconv.AttributeFaaSInvocationID, reqID)
		}
		if status, ok := record["status"].(string); ok {
			span.Attributes().PutStr("aws.lambda.invoke.status", status)
			if status != "success" {
				span.Status().SetCode(ptrace.StatusCodeError)
				if errorType, ok := record["errorType"].(string); ok {
					span.Status().SetMessage(errorType)
				}
			}
		}
	}
	return traces, nil
}

// severityTextToNumber is a helper function preserved from your original code.
func severityTextToNumber(severityText string) plog.SeverityNumber {
	mapping := map[string]plog.SeverityNumber{
		"TRACE":    plog.SeverityNumberTrace,
		"DEBUG":    plog.SeverityNumberDebug,
		"INFO":     plog.SeverityNumberInfo,
		"WARN":     plog.SeverityNumberWarn,
		"WARNING":  plog.SeverityNumberWarn,
		"ERROR":    plog.SeverityNumberError,
		"FATAL":    plog.SeverityNumberFatal,
		"CRITICAL": plog.SeverityNumberFatal,
	}
	if s, ok := mapping[strings.ToUpper(severityText)]; ok {
		return s
	}
	return plog.SeverityNumberUnspecified
}
