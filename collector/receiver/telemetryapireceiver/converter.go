package telemetryapireceiver

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv "go.opentelemetry.io/collector/semconv/v1.25.0"
	"go.uber.org/zap"
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
			logRecord.SetSeverityText(level)
		}
		if msg, ok := record["message"].(string); ok {
			logRecord.Body().SetStr(msg)
		}
		if reqID, ok := record["requestId"].(string); ok {
			logRecord.Attributes().PutStr(semconv.AttributeFaaSInvocationID, reqID)
		}
		if traceID, ok := record["trace_id"].(string); ok {
			if traceBytes, err := hex.DecodeString(traceID); err == nil && len(traceBytes) == 16 {
				var tid pcommon.TraceID
				copy(tid[:], traceBytes)
				logRecord.SetTraceID(tid)
			} else {
				r.logger.Warn("Malformed trace_id found in function log", zap.String("trace_id", traceID))
			}
		}
		if spanID, ok := record["span_id"].(string); ok {
			if spanBytes, err := hex.DecodeString(spanID); err == nil && len(spanBytes) == 8 {
				var sid pcommon.SpanID
				copy(sid[:], spanBytes)
				logRecord.SetSpanID(sid)
			} else {
				r.logger.Warn("Malformed span_id found in function log", zap.String("span_id", spanID))
			}
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
	metricData, ok := record["metrics"].(map[string]interface{})
	if !ok {
		return metrics, fmt.Errorf("metrics field not found in record")
	}
	reqID, _ := record["requestId"].(string)
	ts := pcommon.NewTimestampFromTime(e.getTime())

	for key, value := range metricData {
		if val, ok := value.(float64); ok {
			m := scopeMetrics.Metrics().AppendEmpty()
			m.SetName(fmt.Sprintf("aws.lambda.%s", key))
			unit := "1"
			switch key {
			case "durationMs", "billedDurationMs", "initDurationMs", "restoreDurationMs":
				unit = "ms"
			case "memorySizeMB", "maxMemoryUsedMB":
				unit = "By"
				val = val * 1024 * 1024
			case "producedBytes":
				unit = "By"
			}
			m.SetUnit(unit)
			dp := m.SetEmptyGauge().DataPoints().AppendEmpty()
			dp.SetTimestamp(ts)
			if unit == "By" {
				dp.SetIntValue(int64(val))
			} else {
				dp.SetDoubleValue(val)
			}
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

	if record, ok := e.Record.(map[string]interface{}); ok {
		setSpanStatus(span, record)
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

	if record, ok := e.Record.(map[string]interface{}); ok {
		if reqID, ok := record["requestId"].(string); ok {
			span.Attributes().PutStr(semconv.AttributeFaaSInvocationID, reqID)
		}
		setSpanStatus(span, record)
	}
	return traces, nil
}

// setSpanStatus is a helper to set the status of a span based on the event record.
func setSpanStatus(span ptrace.Span, record map[string]interface{}) {
	if status, ok := record["status"].(string); ok {
		span.Attributes().PutStr("aws.lambda.status", status)
		if status != "success" {
			span.Status().SetCode(ptrace.StatusCodeError)
			if errorType, ok := record["errorType"].(string); ok && errorType != "" {
				span.Status().SetMessage(errorType)
			}
		}
	}
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
