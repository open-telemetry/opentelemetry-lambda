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

package telemetryapireceiver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestHistogramMetricBuilder_RecordWithAttributes(t *testing.T) {
	startTime := pcommon.NewTimestampFromTime(time.Now().Add(-time.Hour))

	t.Run("single data point with attributes", func(t *testing.T) {
		builder := NewHistogramMetricBuilder(
			"test.histogram",
			"Test histogram with attributes",
			"ms",
			[]float64{1.0, 5.0, 10.0},
			startTime,
			pmetric.AggregationTemporalityCumulative,
		)

		attrs := pcommon.NewMap()
		attrs.PutStr("service.name", "test-service")
		attrs.PutStr("region", "us-east-1")
		attrs.PutInt("instance.id", 42)

		builder.RecordWithAttributes(3.5, attrs)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()
		timestamp := pcommon.NewTimestampFromTime(time.Now())

		builder.AppendDataPoints(scopeMetrics, timestamp)

		require.Equal(t, 1, scopeMetrics.Metrics().Len())
		hist := scopeMetrics.Metrics().At(0).Histogram()
		require.Equal(t, 1, hist.DataPoints().Len())

		dp := hist.DataPoints().At(0)
		assert.Equal(t, 3, dp.Attributes().Len())

		val, ok := dp.Attributes().Get("service.name")
		require.True(t, ok)
		assert.Equal(t, "test-service", val.Str())

		val, ok = dp.Attributes().Get("region")
		require.True(t, ok)
		assert.Equal(t, "us-east-1", val.Str())

		val, ok = dp.Attributes().Get("instance.id")
		require.True(t, ok)
		assert.Equal(t, int64(42), val.Int())
	})

	t.Run("multiple data points with different attributes create separate series", func(t *testing.T) {
		builder := NewHistogramMetricBuilder(
			"test.histogram",
			"Test histogram with different attributes",
			"ms",
			[]float64{1.0, 5.0, 10.0},
			startTime,
			pmetric.AggregationTemporalityCumulative,
		)

		attrs1 := pcommon.NewMap()
		attrs1.PutStr("function", "handler1")

		attrs2 := pcommon.NewMap()
		attrs2.PutStr("function", "handler2")

		builder.RecordWithAttributes(2.0, attrs1)
		builder.RecordWithAttributes(7.0, attrs2)
		builder.RecordWithAttributes(3.0, attrs1) // Same attributes as first, should aggregate

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()
		timestamp := pcommon.NewTimestampFromTime(time.Now())

		builder.AppendDataPoints(scopeMetrics, timestamp)

		require.Equal(t, 1, scopeMetrics.Metrics().Len())
		hist := scopeMetrics.Metrics().At(0).Histogram()
		require.Equal(t, 2, hist.DataPoints().Len())

		var handler1Dp, handler2Dp pmetric.HistogramDataPoint
		for i := 0; i < hist.DataPoints().Len(); i++ {
			dp := hist.DataPoints().At(i)
			val, _ := dp.Attributes().Get("function")
			if val.Str() == "handler1" {
				handler1Dp = dp
			} else if val.Str() == "handler2" {
				handler2Dp = dp
			}
		}

		assert.Equal(t, uint64(2), handler1Dp.Count())
		assert.Equal(t, 5.0, handler1Dp.Sum())

		assert.Equal(t, uint64(1), handler2Dp.Count())
		assert.Equal(t, 7.0, handler2Dp.Sum())
	})

	t.Run("empty attributes creates separate series from attributed data", func(t *testing.T) {
		builder := NewHistogramMetricBuilder(
			"test.histogram",
			"Test with empty and non-empty attributes",
			"ms",
			[]float64{1.0, 5.0, 10.0},
			startTime,
			pmetric.AggregationTemporalityCumulative,
		)

		attrs := pcommon.NewMap()
		attrs.PutStr("key", "value")

		builder.Record(1.0)
		builder.RecordWithAttributes(2.0, attrs)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()
		timestamp := pcommon.NewTimestampFromTime(time.Now())

		builder.AppendDataPoints(scopeMetrics, timestamp)

		require.Equal(t, 1, scopeMetrics.Metrics().Len())
		hist := scopeMetrics.Metrics().At(0).Histogram()
		require.Equal(t, 2, hist.DataPoints().Len())
	})

	t.Run("same attributes with different values aggregate correctly", func(t *testing.T) {
		builder := NewHistogramMetricBuilder(
			"test.histogram",
			"Test aggregation with same attributes",
			"ms",
			[]float64{1.0, 5.0, 10.0},
			startTime,
			pmetric.AggregationTemporalityCumulative,
		)

		for i := 0; i < 5; i++ {
			attrs := pcommon.NewMap()
			attrs.PutStr("operation", "process")
			builder.RecordWithAttributes(float64(i+1), attrs)
		}

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()
		timestamp := pcommon.NewTimestampFromTime(time.Now())

		builder.AppendDataPoints(scopeMetrics, timestamp)

		hist := scopeMetrics.Metrics().At(0).Histogram()
		require.Equal(t, 1, hist.DataPoints().Len())

		dp := hist.DataPoints().At(0)
		assert.Equal(t, uint64(5), dp.Count())
		assert.Equal(t, 15.0, dp.Sum()) // 1+2+3+4+5 = 15
	})
}

func TestCounterMetricBuilder_AddWithAttributes(t *testing.T) {
	startTime := pcommon.NewTimestampFromTime(time.Now().Add(-time.Hour))

	t.Run("single data point with attributes", func(t *testing.T) {
		builder := NewCounterMetricBuilder(
			"test.counter",
			"Test counter with attributes",
			"{count}",
			true,
			startTime,
			pmetric.AggregationTemporalityCumulative,
		)

		attrs := pcommon.NewMap()
		attrs.PutStr("error.type", "timeout")
		attrs.PutStr("service.name", "payment-service")

		builder.AddWithAttributes(5, attrs)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()
		timestamp := pcommon.NewTimestampFromTime(time.Now())

		builder.AppendDataPoints(scopeMetrics, timestamp)

		require.Equal(t, 1, scopeMetrics.Metrics().Len())
		sum := scopeMetrics.Metrics().At(0).Sum()
		require.Equal(t, 1, sum.DataPoints().Len())

		dp := sum.DataPoints().At(0)
		assert.Equal(t, 2, dp.Attributes().Len())
		assert.Equal(t, int64(5), dp.IntValue())

		val, ok := dp.Attributes().Get("error.type")
		require.True(t, ok)
		assert.Equal(t, "timeout", val.Str())
	})

	t.Run("multiple data points with different attributes create separate series", func(t *testing.T) {
		builder := NewCounterMetricBuilder(
			"test.counter",
			"Test counter with different attributes",
			"{count}",
			true,
			startTime,
			pmetric.AggregationTemporalityCumulative,
		)

		attrs1 := pcommon.NewMap()
		attrs1.PutStr("status_code", "200")

		attrs2 := pcommon.NewMap()
		attrs2.PutStr("status_code", "500")

		attrs3 := pcommon.NewMap()
		attrs3.PutStr("status_code", "404")

		builder.AddWithAttributes(100, attrs1)
		builder.AddWithAttributes(5, attrs2)
		builder.AddWithAttributes(10, attrs3)
		builder.AddWithAttributes(50, attrs1)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()
		timestamp := pcommon.NewTimestampFromTime(time.Now())

		builder.AppendDataPoints(scopeMetrics, timestamp)

		sum := scopeMetrics.Metrics().At(0).Sum()
		require.Equal(t, 3, sum.DataPoints().Len())

		valuesByStatus := make(map[string]int64)
		for i := 0; i < sum.DataPoints().Len(); i++ {
			dp := sum.DataPoints().At(i)
			val, _ := dp.Attributes().Get("status_code")
			valuesByStatus[val.Str()] = dp.IntValue()
		}

		assert.Equal(t, int64(150), valuesByStatus["200"])
		assert.Equal(t, int64(5), valuesByStatus["500"])
		assert.Equal(t, int64(10), valuesByStatus["404"])
	})

	t.Run("FaaS errors counter with trigger attribute", func(t *testing.T) {
		builder := NewFaaSErrorsMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)

		httpTrigger := pcommon.NewMap()
		httpTrigger.PutStr("faas.trigger", "http")

		sqsTrigger := pcommon.NewMap()
		sqsTrigger.PutStr("faas.trigger", "pubsub")

		builder.AddWithAttributes(3, httpTrigger)
		builder.AddWithAttributes(1, sqsTrigger)
		builder.AddWithAttributes(2, httpTrigger)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()
		timestamp := pcommon.NewTimestampFromTime(time.Now())

		builder.AppendDataPoints(scopeMetrics, timestamp)

		sum := scopeMetrics.Metrics().At(0).Sum()
		require.Equal(t, 2, sum.DataPoints().Len())

		var httpCount, pubsubCount int64
		for i := 0; i < sum.DataPoints().Len(); i++ {
			dp := sum.DataPoints().At(i)
			val, _ := dp.Attributes().Get("faas.trigger")
			if val.Str() == "http" {
				httpCount = dp.IntValue()
			} else if val.Str() == "pubsub" {
				pubsubCount = dp.IntValue()
			}
		}

		assert.Equal(t, int64(5), httpCount)
		assert.Equal(t, int64(1), pubsubCount)
	})

	t.Run("multiple attributes create unique series", func(t *testing.T) {
		builder := NewCounterMetricBuilder(
			"test.counter",
			"Test with multiple attributes",
			"{count}",
			true,
			startTime,
			pmetric.AggregationTemporalityCumulative,
		)

		attrs1 := pcommon.NewMap()
		attrs1.PutStr("method", "GET")
		attrs1.PutStr("path", "/api/users")

		attrs2 := pcommon.NewMap()
		attrs2.PutStr("method", "GET")
		attrs2.PutStr("path", "/api/orders")

		attrs3 := pcommon.NewMap()
		attrs3.PutStr("method", "POST")
		attrs3.PutStr("path", "/api/users")

		builder.AddWithAttributes(10, attrs1)
		builder.AddWithAttributes(20, attrs2)
		builder.AddWithAttributes(5, attrs3)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()
		timestamp := pcommon.NewTimestampFromTime(time.Now())

		builder.AppendDataPoints(scopeMetrics, timestamp)

		sum := scopeMetrics.Metrics().At(0).Sum()
		require.Equal(t, 3, sum.DataPoints().Len())
	})
}

func TestMetricBuilder_AttributesWithDeltaTemporality(t *testing.T) {
	startTime := pcommon.NewTimestampFromTime(time.Now().Add(-time.Hour))

	t.Run("histogram delta resets attributed data points", func(t *testing.T) {
		builder := NewHistogramMetricBuilder(
			"test.histogram",
			"Test delta histogram with attributes",
			"ms",
			[]float64{1.0, 5.0, 10.0},
			startTime,
			pmetric.AggregationTemporalityDelta,
		)

		attrs := pcommon.NewMap()
		attrs.PutStr("operation", "query")

		builder.RecordWithAttributes(2.0, attrs)
		builder.RecordWithAttributes(3.0, attrs)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()

		ts1 := pcommon.NewTimestampFromTime(time.Now())
		builder.AppendDataPoints(scopeMetrics, ts1)

		hist1 := scopeMetrics.Metrics().At(0).Histogram()
		dp1 := hist1.DataPoints().At(0)
		assert.Equal(t, uint64(2), dp1.Count())
		assert.Equal(t, 5.0, dp1.Sum())

		builder.RecordWithAttributes(7.0, attrs)

		ts2 := pcommon.NewTimestampFromTime(time.Now().Add(time.Second))
		builder.AppendDataPoints(scopeMetrics, ts2)

		hist2 := scopeMetrics.Metrics().At(1).Histogram()
		dp2 := hist2.DataPoints().At(0)
		assert.Equal(t, uint64(1), dp2.Count())
		assert.Equal(t, 7.0, dp2.Sum())
	})

	t.Run("counter delta resets attributed data points", func(t *testing.T) {
		builder := NewCounterMetricBuilder(
			"test.counter",
			"Test delta counter with attributes",
			"{count}",
			true,
			startTime,
			pmetric.AggregationTemporalityDelta,
		)

		attrs := pcommon.NewMap()
		attrs.PutStr("status", "success")

		builder.AddWithAttributes(10, attrs)
		builder.AddWithAttributes(5, attrs)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()

		ts1 := pcommon.NewTimestampFromTime(time.Now())
		builder.AppendDataPoints(scopeMetrics, ts1)

		sum1 := scopeMetrics.Metrics().At(0).Sum()
		dp1 := sum1.DataPoints().At(0)
		assert.Equal(t, int64(15), dp1.IntValue())

		builder.AddWithAttributes(3, attrs)

		ts2 := pcommon.NewTimestampFromTime(time.Now().Add(time.Second))
		builder.AppendDataPoints(scopeMetrics, ts2)

		sum2 := scopeMetrics.Metrics().At(1).Sum()
		dp2 := sum2.DataPoints().At(0)
		assert.Equal(t, int64(3), dp2.IntValue())
	})
}

func TestMetricBuilder_ResetClearsAttributes(t *testing.T) {
	startTime := pcommon.NewTimestampFromTime(time.Now().Add(-time.Hour))

	t.Run("histogram reset clears all attributed data points", func(t *testing.T) {
		builder := NewHistogramMetricBuilder(
			"test.histogram",
			"Test reset clears attributes",
			"ms",
			[]float64{1.0, 5.0, 10.0},
			startTime,
			pmetric.AggregationTemporalityCumulative,
		)

		attrs1 := pcommon.NewMap()
		attrs1.PutStr("region", "us-east")

		attrs2 := pcommon.NewMap()
		attrs2.PutStr("region", "eu-west")

		builder.RecordWithAttributes(1.0, attrs1)
		builder.RecordWithAttributes(2.0, attrs2)

		newStartTime := pcommon.NewTimestampFromTime(time.Now())
		builder.Reset(newStartTime)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()
		timestamp := pcommon.NewTimestampFromTime(time.Now().Add(time.Second))

		builder.AppendDataPoints(scopeMetrics, timestamp)

		assert.Equal(t, 0, scopeMetrics.Metrics().Len())
	})

	t.Run("counter reset clears all attributed data points", func(t *testing.T) {
		builder := NewCounterMetricBuilder(
			"test.counter",
			"Test reset clears attributes",
			"{count}",
			true,
			startTime,
			pmetric.AggregationTemporalityCumulative,
		)

		attrs1 := pcommon.NewMap()
		attrs1.PutStr("tier", "free")

		attrs2 := pcommon.NewMap()
		attrs2.PutStr("tier", "premium")

		builder.AddWithAttributes(10, attrs1)
		builder.AddWithAttributes(100, attrs2)

		newStartTime := pcommon.NewTimestampFromTime(time.Now())
		builder.Reset(newStartTime)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()
		timestamp := pcommon.NewTimestampFromTime(time.Now().Add(time.Second))

		builder.AppendDataPoints(scopeMetrics, timestamp)

		assert.Equal(t, 0, scopeMetrics.Metrics().Len())
	})
}

func TestMetricBuilder_AttributeOrdering(t *testing.T) {
	startTime := pcommon.NewTimestampFromTime(time.Now().Add(-time.Hour))

	t.Run("attribute order does not affect aggregation", func(t *testing.T) {
		builder := NewCounterMetricBuilder(
			"test.counter",
			"Test attribute ordering",
			"{count}",
			true,
			startTime,
			pmetric.AggregationTemporalityCumulative,
		)

		attrs1 := pcommon.NewMap()
		attrs1.PutStr("a", "1")
		attrs1.PutStr("b", "2")
		attrs1.PutStr("c", "3")

		attrs2 := pcommon.NewMap()
		attrs2.PutStr("c", "3")
		attrs2.PutStr("a", "1")
		attrs2.PutStr("b", "2")

		builder.AddWithAttributes(10, attrs1)
		builder.AddWithAttributes(20, attrs2)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()
		timestamp := pcommon.NewTimestampFromTime(time.Now())

		builder.AppendDataPoints(scopeMetrics, timestamp)

		sum := scopeMetrics.Metrics().At(0).Sum()
		require.Equal(t, 1, sum.DataPoints().Len())
		assert.Equal(t, int64(30), sum.DataPoints().At(0).IntValue())
	})
}

func TestMetricBuilder_CumulativeEpochWithAttributes(t *testing.T) {
	startTime := pcommon.NewTimestampFromTime(time.Now().Add(-time.Hour))

	t.Run("cumulative histogram only exports updated attributes in epoch", func(t *testing.T) {
		builder := NewHistogramMetricBuilder(
			"test.histogram",
			"Test epoch tracking",
			"ms",
			[]float64{1.0, 5.0, 10.0},
			startTime,
			pmetric.AggregationTemporalityCumulative,
		)

		attrs1 := pcommon.NewMap()
		attrs1.PutStr("key", "value1")

		attrs2 := pcommon.NewMap()
		attrs2.PutStr("key", "value2")

		builder.RecordWithAttributes(1.0, attrs1)
		builder.RecordWithAttributes(2.0, attrs2)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()
		ts1 := pcommon.NewTimestampFromTime(time.Now())

		builder.AppendDataPoints(scopeMetrics, ts1)
		require.Equal(t, 1, scopeMetrics.Metrics().Len())
		require.Equal(t, 2, scopeMetrics.Metrics().At(0).Histogram().DataPoints().Len())

		builder.RecordWithAttributes(3.0, attrs1)

		ts2 := pcommon.NewTimestampFromTime(time.Now().Add(time.Second))
		builder.AppendDataPoints(scopeMetrics, ts2)

		require.Equal(t, 2, scopeMetrics.Metrics().Len())
		hist2 := scopeMetrics.Metrics().At(1).Histogram()
		require.Equal(t, 1, hist2.DataPoints().Len())

		dp := hist2.DataPoints().At(0)
		val, _ := dp.Attributes().Get("key")
		assert.Equal(t, "value1", val.Str())
		assert.Equal(t, uint64(2), dp.Count())
		assert.Equal(t, 4.0, dp.Sum())
	})

	t.Run("cumulative counter only exports updated attributes in epoch", func(t *testing.T) {
		builder := NewCounterMetricBuilder(
			"test.counter",
			"Test epoch tracking",
			"{count}",
			true,
			startTime,
			pmetric.AggregationTemporalityCumulative,
		)

		attrs1 := pcommon.NewMap()
		attrs1.PutStr("service", "api")

		attrs2 := pcommon.NewMap()
		attrs2.PutStr("service", "worker")

		builder.AddWithAttributes(10, attrs1)
		builder.AddWithAttributes(20, attrs2)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()
		ts1 := pcommon.NewTimestampFromTime(time.Now())

		builder.AppendDataPoints(scopeMetrics, ts1)
		require.Equal(t, 2, scopeMetrics.Metrics().At(0).Sum().DataPoints().Len())

		builder.AddWithAttributes(5, attrs2)

		ts2 := pcommon.NewTimestampFromTime(time.Now().Add(time.Second))
		builder.AppendDataPoints(scopeMetrics, ts2)

		sum2 := scopeMetrics.Metrics().At(1).Sum()
		require.Equal(t, 1, sum2.DataPoints().Len())

		dp := sum2.DataPoints().At(0)
		val, _ := dp.Attributes().Get("service")
		assert.Equal(t, "worker", val.Str())
		assert.Equal(t, int64(25), dp.IntValue())
	})
}
