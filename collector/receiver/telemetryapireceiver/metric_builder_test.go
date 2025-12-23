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
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

func TestHistogramMetricBuilder_AppendDataPoint(t *testing.T) {
	startTime := pcommon.NewTimestampFromTime(time.Now().Add(-time.Hour))

	tests := []struct {
		name           string
		builder        *HistogramMetricBuilder
		value          float64
		expectedBucket int
	}{
		{
			name:           "FaaS invoke duration - small value",
			builder:        NewFaaSInvokeDurationMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative),
			value:          0.007,
			expectedBucket: 1,
		},
		{
			name:           "FaaS invoke duration - middle value",
			builder:        NewFaaSInvokeDurationMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative),
			value:          0.5,
			expectedBucket: 7,
		},
		{
			name:           "FaaS invoke duration - large value",
			builder:        NewFaaSInvokeDurationMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative),
			value:          15.0,
			expectedBucket: 14,
		},
		{
			name:           "Default bounds - boundary value",
			builder:        NewHistogramMetricBuilder("test.histogram", "Test with default bounds", "By", DefaultHistogramBounds, startTime, pmetric.AggregationTemporalityCumulative),
			value:          100.0,
			expectedBucket: 6,
		},
		{
			name:           "Default bounds - zero value",
			builder:        NewHistogramMetricBuilder("test.histogram", "Test zero value", "By", nil, startTime, pmetric.AggregationTemporalityCumulative),
			value:          0.0,
			expectedBucket: 0,
		},
		{
			name:           "Memory usage histogram",
			builder:        NewFaaSMemUsageMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative),
			value:          256.0 * 1024 * 1024,
			expectedBucket: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := pmetric.NewMetrics()
			rm := metrics.ResourceMetrics().AppendEmpty()
			scopeMetrics := rm.ScopeMetrics().AppendEmpty()

			timestamp := pcommon.NewTimestampFromTime(time.Now())

			tt.builder.Record(tt.value)
			tt.builder.AppendDataPoints(scopeMetrics, timestamp)

			require.Equal(t, 1, scopeMetrics.Metrics().Len())

			metric := scopeMetrics.Metrics().At(0)
			assert.Equal(t, tt.builder.name, metric.Name())
			assert.Equal(t, tt.builder.description, metric.Description())
			assert.Equal(t, tt.builder.unit, metric.Unit())

			assert.Equal(t, pmetric.MetricTypeHistogram, metric.Type())

			hist := metric.Histogram()
			assert.Equal(t, tt.builder.temporality, hist.AggregationTemporality())

			require.Equal(t, 1, hist.DataPoints().Len())
			dp := hist.DataPoints().At(0)
			assert.Equal(t, startTime, dp.StartTimestamp())
			assert.Equal(t, timestamp, dp.Timestamp())
			assert.Equal(t, uint64(1), dp.Count())
			assert.Equal(t, tt.value, dp.Sum())

			assert.Equal(t, len(tt.builder.bounds), dp.ExplicitBounds().Len())
			assert.Equal(t, len(tt.builder.bounds)+1, dp.BucketCounts().Len())

			bucketCounts := dp.BucketCounts().AsRaw()
			for i, count := range bucketCounts {
				if i == tt.expectedBucket {
					assert.Equal(t, uint64(1), count, "expected value in bucket %d", i)
				} else {
					assert.Equal(t, uint64(0), count, "expected no value in bucket %d", i)
				}
			}
		})
	}
}

func TestCounterMetricBuilder_AppendDataPoint(t *testing.T) {
	startTime := pcommon.NewTimestampFromTime(time.Now().Add(-time.Hour))

	tests := []struct {
		name        string
		builder     *CounterMetricBuilder
		value       int64
		isMonotonic bool
	}{
		{
			name:        "FaaS coldstarts counter",
			builder:     NewFaaSColdstartsMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative),
			value:       1,
			isMonotonic: true,
		},
		{
			name:        "FaaS errors counter",
			builder:     NewFaaSErrorsMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative),
			value:       5,
			isMonotonic: true,
		},
		{
			name:        "FaaS invocations counter",
			builder:     NewFaaSInvocationsMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative),
			value:       100,
			isMonotonic: true,
		},
		{
			name:        "FaaS timeouts counter",
			builder:     NewFaaSTimeoutsMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative),
			value:       0,
			isMonotonic: true,
		},
		{
			name:        "Non-monotonic Counter",
			builder:     NewCounterMetricBuilder("test.counter", "Test non-monotonic counter", "{count}", false, startTime, pmetric.AggregationTemporalityCumulative),
			value:       -10,
			isMonotonic: false,
		},
		{
			name:        "Counter with large value",
			builder:     NewCounterMetricBuilder("test.large_counter", "Test large counter value", "{count}", true, startTime, pmetric.AggregationTemporalityCumulative),
			value:       9223372036854775807, // max int64
			isMonotonic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := pmetric.NewMetrics()
			rm := metrics.ResourceMetrics().AppendEmpty()
			scopeMetrics := rm.ScopeMetrics().AppendEmpty()

			timestamp := pcommon.NewTimestampFromTime(time.Now())

			tt.builder.Add(tt.value)
			tt.builder.AppendDataPoints(scopeMetrics, timestamp)

			require.Equal(t, 1, scopeMetrics.Metrics().Len())

			metric := scopeMetrics.Metrics().At(0)
			assert.Equal(t, tt.builder.name, metric.Name())
			assert.Equal(t, tt.builder.description, metric.Description())
			assert.Equal(t, tt.builder.unit, metric.Unit())

			assert.Equal(t, pmetric.MetricTypeSum, metric.Type())

			sum := metric.Sum()
			assert.Equal(t, tt.builder.temporality, sum.AggregationTemporality())
			assert.Equal(t, tt.isMonotonic, sum.IsMonotonic())

			require.Equal(t, 1, sum.DataPoints().Len())
			dp := sum.DataPoints().At(0)
			assert.Equal(t, startTime, dp.StartTimestamp())
			assert.Equal(t, timestamp, dp.Timestamp())
			assert.Equal(t, tt.value, dp.IntValue())
		})
	}
}

func TestFaaSMetricBuilderFactories(t *testing.T) {
	startTime := pcommon.NewTimestampFromTime(time.Now())

	t.Run("NewFasSInvokeDurationMetricBuilder", func(t *testing.T) {
		builder := NewFaaSInvokeDurationMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
		assert.Equal(t, semconv.FaaSInvokeDurationName, builder.name)
		assert.Equal(t, semconv.FaaSInvokeDurationDescription, builder.description)
		assert.Equal(t, semconv.FaaSInvokeDurationUnit, builder.unit)
		assert.Equal(t, DurationHistogramBounds, builder.bounds)
		assert.Equal(t, pmetric.AggregationTemporalityCumulative, builder.temporality)
		assert.Equal(t, startTime, builder.startTime)
	})

	t.Run("NewFasSInitDurationMetricBuilder", func(t *testing.T) {
		builder := NewFaaSInitDurationMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
		assert.Equal(t, semconv.FaaSInitDurationName, builder.name)
		assert.Equal(t, semconv.FaaSInitDurationDescription, builder.description)
		assert.Equal(t, semconv.FaaSInitDurationUnit, builder.unit)
		assert.Equal(t, DurationHistogramBounds, builder.bounds)
		assert.Equal(t, pmetric.AggregationTemporalityCumulative, builder.temporality)
		assert.Equal(t, startTime, builder.startTime)
	})

	t.Run("NewFaaSMemUsageMetricBuilder", func(t *testing.T) {
		builder := NewFaaSMemUsageMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
		assert.Equal(t, semconv.FaaSMemUsageName, builder.name)
		assert.Equal(t, semconv.FaaSMemUsageDescription, builder.description)
		assert.Equal(t, semconv.FaaSMemUsageUnit, builder.unit)
		assert.Equal(t, MemUsageHistogramBounds, builder.bounds)
		assert.Equal(t, pmetric.AggregationTemporalityCumulative, builder.temporality)
		assert.Equal(t, startTime, builder.startTime)
	})

	t.Run("NewFaaSColdstartsMetricBuilder", func(t *testing.T) {
		builder := NewFaaSColdstartsMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
		assert.Equal(t, semconv.FaaSColdstartsName, builder.name)
		assert.Equal(t, semconv.FaaSColdstartsDescription, builder.description)
		assert.Equal(t, semconv.FaaSColdstartsUnit, builder.unit)
		assert.True(t, builder.isMonotonic)
		assert.Equal(t, pmetric.AggregationTemporalityCumulative, builder.temporality)
		assert.Equal(t, startTime, builder.startTime)
	})

	t.Run("NewFaaSErrorsMetricBuilder", func(t *testing.T) {
		builder := NewFaaSErrorsMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
		assert.Equal(t, semconv.FaaSErrorsName, builder.name)
		assert.Equal(t, semconv.FaaSErrorsDescription, builder.description)
		assert.Equal(t, semconv.FaaSErrorsUnit, builder.unit)
		assert.True(t, builder.isMonotonic)
		assert.Equal(t, pmetric.AggregationTemporalityCumulative, builder.temporality)
		assert.Equal(t, startTime, builder.startTime)
	})

	t.Run("NewFaaSInvocationsMetricBuilder", func(t *testing.T) {
		builder := NewFaaSInvocationsMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
		assert.Equal(t, semconv.FaaSInvocationsName, builder.name)
		assert.Equal(t, semconv.FaaSInvocationsDescription, builder.description)
		assert.Equal(t, semconv.FaaSInvocationsUnit, builder.unit)
		assert.True(t, builder.isMonotonic)
		assert.Equal(t, pmetric.AggregationTemporalityCumulative, builder.temporality)
		assert.Equal(t, startTime, builder.startTime)
	})

	t.Run("NewFaaSTimeoutsMetricBuilder", func(t *testing.T) {
		builder := NewFaaSTimeoutsMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
		assert.Equal(t, semconv.FaaSTimeoutsName, builder.name)
		assert.Equal(t, semconv.FaaSTimeoutsDescription, builder.description)
		assert.Equal(t, semconv.FaaSTimeoutsUnit, builder.unit)
		assert.True(t, builder.isMonotonic)
		assert.Equal(t, pmetric.AggregationTemporalityCumulative, builder.temporality)
		assert.Equal(t, startTime, builder.startTime)
	})
}

func TestNewFaaSMetricBuilders(t *testing.T) {
	startTime := pcommon.NewTimestampFromTime(time.Now())
	builders := NewFaaSMetricBuilders(startTime, pmetric.AggregationTemporalityCumulative)

	require.NotNil(t, builders)
	require.NotNil(t, builders.invokeDurationMetric)
	require.NotNil(t, builders.initDurationMetric)
	require.NotNil(t, builders.memUsageMetric)
	require.NotNil(t, builders.coldstartsMetric)
	require.NotNil(t, builders.errorsMetric)
	require.NotNil(t, builders.invocationsMetric)
	require.NotNil(t, builders.timeoutsMetric)

	assert.Equal(t, semconv.FaaSInvokeDurationName, builders.invokeDurationMetric.name)
	assert.Equal(t, semconv.FaaSInitDurationName, builders.initDurationMetric.name)
	assert.Equal(t, semconv.FaaSMemUsageName, builders.memUsageMetric.name)
	assert.Equal(t, semconv.FaaSColdstartsName, builders.coldstartsMetric.name)
	assert.Equal(t, semconv.FaaSErrorsName, builders.errorsMetric.name)
	assert.Equal(t, semconv.FaaSInvocationsName, builders.invocationsMetric.name)
	assert.Equal(t, semconv.FaaSTimeoutsName, builders.timeoutsMetric.name)
}

func TestDefaultHistogramBounds(t *testing.T) {
	expected := []float64{0.0, 5.0, 10.0, 25.0, 50.0, 75.0, 100.0, 250.0, 500.0, 750.0, 1000.0, 2500.0, 5000.0, 7500.0, 10000.0}
	assert.Equal(t, expected, DefaultHistogramBounds)
	assert.Len(t, DefaultHistogramBounds, 15)
}

func TestDurationHistogramBounds(t *testing.T) {
	expected := []float64{0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10}
	assert.Equal(t, expected, DurationHistogramBounds)
	assert.Len(t, DurationHistogramBounds, 14)
}

func TestHistogramBucketPlacement(t *testing.T) {
	startTime := pcommon.NewTimestampFromTime(time.Now().Add(-time.Hour))

	tests := []struct {
		name           string
		bounds         []float64
		value          float64
		expectedBucket int
	}{
		{
			name:           "value less than first bound",
			bounds:         []float64{1.0, 5.0, 10.0},
			value:          0.5,
			expectedBucket: 0,
		},
		{
			name:           "value equals first bound",
			bounds:         []float64{1.0, 5.0, 10.0},
			value:          1.0,
			expectedBucket: 0,
		},
		{
			name:           "value between bounds",
			bounds:         []float64{1.0, 5.0, 10.0},
			value:          3.0,
			expectedBucket: 1,
		},
		{
			name:           "value equals middle bound",
			bounds:         []float64{1.0, 5.0, 10.0},
			value:          5.0,
			expectedBucket: 1,
		},
		{
			name:           "value greater than all bounds",
			bounds:         []float64{1.0, 5.0, 10.0},
			value:          15.0,
			expectedBucket: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewHistogramMetricBuilder(
				"test.histogram",
				"Test bucket placement",
				"1",
				tt.bounds,
				startTime,
				pmetric.AggregationTemporalityCumulative,
			)

			metrics := pmetric.NewMetrics()
			rm := metrics.ResourceMetrics().AppendEmpty()
			scopeMetrics := rm.ScopeMetrics().AppendEmpty()

			timestamp := pcommon.NewTimestampFromTime(time.Now())
			builder.Record(tt.value)
			builder.AppendDataPoints(scopeMetrics, timestamp)

			dp := scopeMetrics.Metrics().At(0).Histogram().DataPoints().At(0)
			bucketCounts := dp.BucketCounts().AsRaw()

			for i, count := range bucketCounts {
				if i == tt.expectedBucket {
					assert.Equal(t, uint64(1), count, "expected value in bucket %d for value %f", i, tt.value)
				} else {
					assert.Equal(t, uint64(0), count, "expected no value in bucket %d for value %f", i, tt.value)
				}
			}
		})
	}
}

func TestHistogramMetricBuilder_CumulativeDataPoints(t *testing.T) {
	startTime := pcommon.NewTimestampFromTime(time.Now().Add(-time.Hour))

	tests := []struct {
		name                string
		builderFn           func() *HistogramMetricBuilder
		values              []float64
		expectedCount       uint64
		expectedSum         float64
		expectedBucketIndex int
		expectedBucketCount uint64
		checkAllBuckets     bool
	}{
		{
			name: "two data points accumulate correctly",
			builderFn: func() *HistogramMetricBuilder {
				return NewFaaSInvokeDurationMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
			},
			values:        []float64{0.1, 0.2},
			expectedCount: 2,
			expectedSum:   0.3,
		},
		{
			name: "multiple data points across different buckets",
			builderFn: func() *HistogramMetricBuilder {
				return NewFaaSInvokeDurationMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
			},
			values:        []float64{0.001, 0.05, 0.5, 1.0, 5.0},
			expectedCount: 5,
			expectedSum:   6.551,
		},
		{
			name: "same bucket receives multiple values",
			builderFn: func() *HistogramMetricBuilder {
				return NewFaaSInvokeDurationMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
			},
			values:              []float64{0.3, 0.35, 0.4, 0.45},
			expectedCount:       4,
			expectedSum:         1.5,
			expectedBucketIndex: 7,
			expectedBucketCount: 4,
			checkAllBuckets:     true,
		},
		{
			name: "zero values accumulate correctly",
			builderFn: func() *HistogramMetricBuilder {
				return NewFaaSInvokeDurationMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
			},
			values:              []float64{0.0, 0.0, 0.0, 0.0, 0.0},
			expectedCount:       5,
			expectedSum:         0.0,
			expectedBucketIndex: 0,
			expectedBucketCount: 5,
			checkAllBuckets:     true,
		},
		{
			name: "large values in overflow bucket",
			builderFn: func() *HistogramMetricBuilder {
				return NewFaaSInvokeDurationMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
			},
			values:              []float64{15.0, 20.0, 100.0},
			expectedCount:       3,
			expectedSum:         135.0,
			expectedBucketIndex: 14,
			expectedBucketCount: 3,
			checkAllBuckets:     true,
		},
		{
			name: "memory usage histogram with realistic values",
			builderFn: func() *HistogramMetricBuilder {
				return NewFaaSMemUsageMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
			},
			values:        []float64{128.0, 256.0, 512.0, 384.0, 192.0},
			expectedCount: 5,
			expectedSum:   1472.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.builderFn()

			metrics := pmetric.NewMetrics()
			rm := metrics.ResourceMetrics().AppendEmpty()
			scopeMetrics := rm.ScopeMetrics().AppendEmpty()

			baseTime := time.Now()
			for i, v := range tt.values {
				ts := pcommon.NewTimestampFromTime(baseTime.Add(time.Duration(i) * time.Second))
				builder.Record(v)
				builder.AppendDataPoints(scopeMetrics, ts)
			}

			require.Equal(t, len(tt.values), scopeMetrics.Metrics().Len())

			lastDp := scopeMetrics.Metrics().At(len(tt.values) - 1).Histogram().DataPoints().At(0)
			assert.Equal(t, tt.expectedCount, lastDp.Count())
			assert.InDelta(t, tt.expectedSum, lastDp.Sum(), 0.0001)

			if tt.checkAllBuckets {
				bucketCounts := lastDp.BucketCounts().AsRaw()
				for i, count := range bucketCounts {
					if i == tt.expectedBucketIndex {
						assert.Equal(t, tt.expectedBucketCount, count, "expected %d values in bucket %d", tt.expectedBucketCount, i)
					} else {
						assert.Equal(t, uint64(0), count, "bucket %d should be empty", i)
					}
				}
			}
		})
	}

	t.Run("start timestamp remains constant across data points", func(t *testing.T) {
		builder := NewFaaSInvokeDurationMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()

		baseTime := time.Now()
		for i := 0; i < 3; i++ {
			ts := pcommon.NewTimestampFromTime(baseTime.Add(time.Duration(i) * time.Second))
			builder.Record(float64(i) * 0.1)
			builder.AppendDataPoints(scopeMetrics, ts)
		}

		for i := 0; i < 3; i++ {
			dp := scopeMetrics.Metrics().At(i).Histogram().DataPoints().At(0)
			assert.Equal(t, startTime, dp.StartTimestamp())
		}
	})
}

func TestCounterMetricBuilder_CumulativeDataPoints(t *testing.T) {
	startTime := pcommon.NewTimestampFromTime(time.Now().Add(-time.Hour))

	tests := []struct {
		name          string
		builderFn     func() *CounterMetricBuilder
		values        []int64
		expectedTotal int64
	}{
		{
			name: "two data points accumulate correctly",
			builderFn: func() *CounterMetricBuilder {
				return NewFaaSInvocationsMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
			},
			values:        []int64{5, 3},
			expectedTotal: 8,
		},
		{
			name: "multiple increments accumulate correctly",
			builderFn: func() *CounterMetricBuilder {
				return NewFaaSInvocationsMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
			},
			values:        []int64{1, 2, 3, 4, 5},
			expectedTotal: 15,
		},
		{
			name: "zero increments do not change total",
			builderFn: func() *CounterMetricBuilder {
				return NewFaaSInvocationsMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
			},
			values:        []int64{10, 0, 0, 5},
			expectedTotal: 15,
		},
		{
			name: "coldstarts counter increments by one",
			builderFn: func() *CounterMetricBuilder {
				return NewFaaSColdstartsMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
			},
			values:        []int64{1, 1, 1},
			expectedTotal: 3,
		},
		{
			name: "errors counter accumulates",
			builderFn: func() *CounterMetricBuilder {
				return NewFaaSErrorsMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
			},
			values:        []int64{2, 0, 1, 5, 0, 3},
			expectedTotal: 11,
		},
		{
			name: "timeouts counter accumulates",
			builderFn: func() *CounterMetricBuilder {
				return NewFaaSTimeoutsMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
			},
			values:        []int64{1, 1},
			expectedTotal: 2,
		},
		{
			name: "large values accumulate without overflow",
			builderFn: func() *CounterMetricBuilder {
				return NewFaaSInvocationsMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
			},
			values:        []int64{1000000000, 1000000000, 1000000000},
			expectedTotal: 3000000000,
		},
		{
			name: "non-monotonic counter allows negative deltas",
			builderFn: func() *CounterMetricBuilder {
				return NewCounterMetricBuilder("test.gauge", "Test gauge", "{count}", false, startTime, pmetric.AggregationTemporalityCumulative)
			},
			values:        []int64{10, -3, 5, -7},
			expectedTotal: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.builderFn()

			metrics := pmetric.NewMetrics()
			rm := metrics.ResourceMetrics().AppendEmpty()
			scopeMetrics := rm.ScopeMetrics().AppendEmpty()

			baseTime := time.Now()
			for i, v := range tt.values {
				ts := pcommon.NewTimestampFromTime(baseTime.Add(time.Duration(i) * time.Second))
				builder.Add(v)
				builder.AppendDataPoints(scopeMetrics, ts)
			}

			require.Equal(t, len(tt.values), scopeMetrics.Metrics().Len())

			lastDp := scopeMetrics.Metrics().At(len(tt.values) - 1).Sum().DataPoints().At(0)
			assert.Equal(t, tt.expectedTotal, lastDp.IntValue())
		})
	}

	t.Run("start timestamp remains constant across data points", func(t *testing.T) {
		builder := NewFaaSInvocationsMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()

		baseTime := time.Now()
		for i := 0; i < 3; i++ {
			ts := pcommon.NewTimestampFromTime(baseTime.Add(time.Duration(i) * time.Second))
			builder.Add(int64(i + 1))
			builder.AppendDataPoints(scopeMetrics, ts)
		}

		for i := 0; i < 3; i++ {
			dp := scopeMetrics.Metrics().At(i).Sum().DataPoints().At(0)
			assert.Equal(t, startTime, dp.StartTimestamp())
		}
	})

	t.Run("monotonic property is set correctly", func(t *testing.T) {
		monotonicBuilder := NewFaaSInvocationsMetricBuilder(startTime, pmetric.AggregationTemporalityCumulative)
		nonMonotonicBuilder := NewCounterMetricBuilder("test.gauge", "Test gauge", "{count}", false, startTime, pmetric.AggregationTemporalityCumulative)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()

		ts := pcommon.NewTimestampFromTime(time.Now())
		monotonicBuilder.Add(1)
		nonMonotonicBuilder.Add(1)
		monotonicBuilder.AppendDataPoints(scopeMetrics, ts)
		nonMonotonicBuilder.AppendDataPoints(scopeMetrics, ts)

		assert.True(t, scopeMetrics.Metrics().At(0).Sum().IsMonotonic())
		assert.False(t, scopeMetrics.Metrics().At(1).Sum().IsMonotonic())
	})
}

func TestHistogramMetricBuilder_AggregationTemporality(t *testing.T) {
	startTime := pcommon.NewTimestampFromTime(time.Now().Add(-time.Hour))

	t.Run("unspecified temporality defaults to cumulative", func(t *testing.T) {
		builder := NewHistogramMetricBuilder(
			"test.histogram",
			"Test histogram",
			"ms",
			nil,
			startTime,
			pmetric.AggregationTemporalityUnspecified,
		)

		assert.Equal(t, pmetric.AggregationTemporalityCumulative, builder.temporality)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()

		timestamp := pcommon.NewTimestampFromTime(time.Now())
		builder.Record(1.0)
		builder.AppendDataPoints(scopeMetrics, timestamp)

		hist := scopeMetrics.Metrics().At(0).Histogram()
		assert.Equal(t, pmetric.AggregationTemporalityCumulative, hist.AggregationTemporality())
	})

	t.Run("cumulative temporality accumulates values", func(t *testing.T) {
		builder := NewHistogramMetricBuilder(
			"test.histogram",
			"Test histogram",
			"ms",
			[]float64{1.0, 5.0, 10.0},
			startTime,
			pmetric.AggregationTemporalityCumulative,
		)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()

		baseTime := time.Now()
		values := []float64{2.0, 3.0, 7.0}

		for i, v := range values {
			ts := pcommon.NewTimestampFromTime(baseTime.Add(time.Duration(i) * time.Second))
			builder.Record(v)
			builder.AppendDataPoints(scopeMetrics, ts)
		}

		require.Equal(t, 3, scopeMetrics.Metrics().Len())

		dp1 := scopeMetrics.Metrics().At(0).Histogram().DataPoints().At(0)
		assert.Equal(t, uint64(1), dp1.Count())
		assert.Equal(t, 2.0, dp1.Sum())
		assert.Equal(t, startTime, dp1.StartTimestamp())

		dp2 := scopeMetrics.Metrics().At(1).Histogram().DataPoints().At(0)
		assert.Equal(t, uint64(2), dp2.Count())
		assert.Equal(t, 5.0, dp2.Sum())
		assert.Equal(t, startTime, dp2.StartTimestamp())

		dp3 := scopeMetrics.Metrics().At(2).Histogram().DataPoints().At(0)
		assert.Equal(t, uint64(3), dp3.Count())
		assert.Equal(t, 12.0, dp3.Sum())
		assert.Equal(t, startTime, dp3.StartTimestamp())
	})

	t.Run("delta temporality resets after each append", func(t *testing.T) {
		builder := NewHistogramMetricBuilder(
			"test.histogram",
			"Test histogram",
			"ms",
			[]float64{1.0, 5.0, 10.0},
			startTime,
			pmetric.AggregationTemporalityDelta,
		)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()

		baseTime := time.Now()
		ts1 := pcommon.NewTimestampFromTime(baseTime)
		ts2 := pcommon.NewTimestampFromTime(baseTime.Add(time.Second))
		ts3 := pcommon.NewTimestampFromTime(baseTime.Add(2 * time.Second))

		builder.Record(2.0)
		builder.Record(3.0)
		builder.AppendDataPoints(scopeMetrics, ts1)

		builder.Record(7.0)
		builder.AppendDataPoints(scopeMetrics, ts2)

		builder.Record(1.5)
		builder.Record(8.0)
		builder.AppendDataPoints(scopeMetrics, ts3)

		require.Equal(t, 3, scopeMetrics.Metrics().Len())

		dp1 := scopeMetrics.Metrics().At(0).Histogram().DataPoints().At(0)
		assert.Equal(t, uint64(2), dp1.Count())
		assert.Equal(t, 5.0, dp1.Sum())
		assert.Equal(t, startTime, dp1.StartTimestamp())

		dp2 := scopeMetrics.Metrics().At(1).Histogram().DataPoints().At(0)
		assert.Equal(t, uint64(1), dp2.Count())
		assert.Equal(t, 7.0, dp2.Sum())
		assert.Equal(t, ts1, dp2.StartTimestamp())

		dp3 := scopeMetrics.Metrics().At(2).Histogram().DataPoints().At(0)
		assert.Equal(t, uint64(2), dp3.Count())
		assert.Equal(t, 9.5, dp3.Sum())
		assert.Equal(t, ts2, dp3.StartTimestamp())
	})

	t.Run("delta temporality resets bucket counts", func(t *testing.T) {
		builder := NewHistogramMetricBuilder(
			"test.histogram",
			"Test histogram",
			"ms",
			[]float64{5.0, 10.0},
			startTime,
			pmetric.AggregationTemporalityDelta,
		)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := rm.ScopeMetrics().AppendEmpty()

		baseTime := time.Now()
		ts1 := pcommon.NewTimestampFromTime(baseTime)
		ts2 := pcommon.NewTimestampFromTime(baseTime.Add(time.Second))

		builder.Record(1.0)
		builder.Record(2.0)
		builder.Record(3.0)
		builder.AppendDataPoints(scopeMetrics, ts1)

		builder.Record(15.0)
		builder.Record(20.0)
		builder.AppendDataPoints(scopeMetrics, ts2)

		dp1 := scopeMetrics.Metrics().At(0).Histogram().DataPoints().At(0)
		buckets1 := dp1.BucketCounts().AsRaw()
		assert.Equal(t, uint64(3), buckets1[0])
		assert.Equal(t, uint64(0), buckets1[1])
		assert.Equal(t, uint64(0), buckets1[2])

		dp2 := scopeMetrics.Metrics().At(1).Histogram().DataPoints().At(0)
		buckets2 := dp2.BucketCounts().AsRaw()
		assert.Equal(t, uint64(0), buckets2[0])
		assert.Equal(t, uint64(0), buckets2[1])
		assert.Equal(t, uint64(2), buckets2[2])
	})
}
