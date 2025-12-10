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
	"sort"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	semconv2 "go.opentelemetry.io/otel/semconv/v1.24.0"
)

const MiB = float64(1 << 20)
const GiB = float64(1 << 30)

var DefaultHistogramBounds = []float64{0.0, 5.0, 10.0, 25.0, 50.0, 75.0, 100.0, 250.0, 500.0, 750.0, 1000.0, 2500.0, 5000.0, 7500.0, 10000.0}
var DurationHistogramBounds = []float64{0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10}
var MemUsageHistogramBounds = []float64{16 * MiB, 32 * MiB, 64 * MiB, 128 * MiB, 256 * MiB, 512 * MiB, 768 * MiB, 1 * GiB, 2 * GiB, 3 * GiB, 4 * GiB, 6 * GiB, 8 * GiB}

type HistogramMetricBuilder struct {
	name        string
	description string
	unit        string
	bounds      []float64
	counts      []uint64
	total       uint64
	sum         float64
	startTime   pcommon.Timestamp
	temporality pmetric.AggregationTemporality
}

func NewHistogramMetricBuilder(name string, description string, unit string, bounds []float64, startTime pcommon.Timestamp, temporality pmetric.AggregationTemporality) *HistogramMetricBuilder {
	b := bounds
	if bounds == nil {
		b = DefaultHistogramBounds
	}

	temp := temporality
	if temporality == pmetric.AggregationTemporalityUnspecified {
		temp = pmetric.AggregationTemporalityCumulative
	}

	counts := make([]uint64, len(b)+1)
	return &HistogramMetricBuilder{
		name:        name,
		description: description,
		unit:        unit,
		bounds:      b,
		counts:      counts,
		startTime:   startTime,
		temporality: temp,
	}
}

func (h *HistogramMetricBuilder) Record(value float64) {
	h.sum += value
	h.total++
	h.counts[sort.SearchFloat64s(h.bounds, value)]++
}

func (h *HistogramMetricBuilder) Reset(timestamp pcommon.Timestamp) {
	h.startTime = timestamp
	h.sum = 0
	h.total = 0

	for i := range h.counts {
		h.counts[i] = 0
	}
}

func (h *HistogramMetricBuilder) AppendDataPoints(scopeMetrics pmetric.ScopeMetrics, timestamp pcommon.Timestamp) {
	metric := scopeMetrics.Metrics().AppendEmpty()
	metric.SetName(h.name)
	metric.SetDescription(h.description)
	metric.SetUnit(h.unit)

	hist := metric.SetEmptyHistogram()
	hist.SetAggregationTemporality(h.temporality)

	dp := hist.DataPoints().AppendEmpty()
	dp.Attributes()
	dp.SetStartTimestamp(h.startTime)
	dp.SetTimestamp(timestamp)
	dp.SetSum(h.sum)
	dp.SetCount(h.total)

	dp.BucketCounts().FromRaw(h.counts)
	dp.ExplicitBounds().FromRaw(h.bounds)

	if h.temporality == pmetric.AggregationTemporalityDelta {
		h.Reset(timestamp)
	}
}

type CounterMetricBuilder struct {
	name        string
	description string
	unit        string
	total       int64
	isMonotonic bool
	temporality pmetric.AggregationTemporality
	startTime   pcommon.Timestamp
}

func NewCounterMetricBuilder(name string, description string, unit string, isMonotonic bool, startTime pcommon.Timestamp, temporality pmetric.AggregationTemporality) *CounterMetricBuilder {
	temp := temporality
	if temporality == pmetric.AggregationTemporalityUnspecified {
		temp = pmetric.AggregationTemporalityCumulative
	}

	return &CounterMetricBuilder{
		name:        name,
		description: description,
		unit:        unit,
		isMonotonic: isMonotonic,
		temporality: temp,
		startTime:   startTime,
	}
}

func (c *CounterMetricBuilder) Add(value int64) {
	c.total += value
}

func (c *CounterMetricBuilder) Reset(timestamp pcommon.Timestamp) {
	c.startTime = timestamp
	c.total = 0
}

func (c *CounterMetricBuilder) AppendDataPoints(scopeMetrics pmetric.ScopeMetrics, timestamp pcommon.Timestamp) {
	metric := scopeMetrics.Metrics().AppendEmpty()
	metric.SetName(c.name)
	metric.SetDescription(c.description)
	metric.SetUnit(c.unit)

	sum := metric.SetEmptySum()
	sum.SetAggregationTemporality(c.temporality)
	sum.SetIsMonotonic(c.isMonotonic)

	dp := sum.DataPoints().AppendEmpty()
	dp.SetStartTimestamp(c.startTime)
	dp.SetTimestamp(timestamp)
	dp.SetIntValue(c.total)

	if c.temporality == pmetric.AggregationTemporalityDelta {
		c.Reset(timestamp)
	}
}

func NewFasSInvokeDurationMetricBuilder(startTime pcommon.Timestamp, temporality pmetric.AggregationTemporality) *HistogramMetricBuilder {
	return NewHistogramMetricBuilder(
		semconv2.FaaSInvokeDurationName,
		semconv2.FaaSInvokeDurationDescription,
		semconv2.FaaSInvokeDurationUnit,
		DurationHistogramBounds,
		startTime,
		temporality,
	)
}

func NewFasSInitDurationMetricBuilder(startTime pcommon.Timestamp, temporality pmetric.AggregationTemporality) *HistogramMetricBuilder {
	return NewHistogramMetricBuilder(
		semconv2.FaaSInitDurationName,
		semconv2.FaaSInitDurationDescription,
		semconv2.FaaSInitDurationUnit,
		DurationHistogramBounds,
		startTime,
		temporality,
	)
}

func NewFaaSMemUsageMetricBuilder(startTime pcommon.Timestamp, temporality pmetric.AggregationTemporality) *HistogramMetricBuilder {
	return NewHistogramMetricBuilder(
		semconv2.FaaSMemUsageName,
		semconv2.FaaSMemUsageDescription,
		semconv2.FaaSMemUsageUnit,
		MemUsageHistogramBounds,
		startTime,
		temporality,
	)
}

func NewFaaSColdstartsMetricBuilder(startTime pcommon.Timestamp, temporality pmetric.AggregationTemporality) *CounterMetricBuilder {
	return NewCounterMetricBuilder(
		semconv2.FaaSColdstartsName,
		semconv2.FaaSColdstartsDescription,
		semconv2.FaaSColdstartsUnit,
		true,
		startTime,
		temporality,
	)
}

func NewFaaSErrorsMetricBuilder(startTime pcommon.Timestamp, temporality pmetric.AggregationTemporality) *CounterMetricBuilder {
	return NewCounterMetricBuilder(
		semconv2.FaaSErrorsName,
		semconv2.FaaSErrorsDescription,
		semconv2.FaaSErrorsUnit,
		true,
		startTime,
		temporality,
	)
}

func NewFaaSInvocationsMetricBuilder(startTime pcommon.Timestamp, temporality pmetric.AggregationTemporality) *CounterMetricBuilder {
	return NewCounterMetricBuilder(
		semconv2.FaaSInvocationsName,
		semconv2.FaaSInvocationsDescription,
		semconv2.FaaSInvocationsUnit,
		true,
		startTime,
		temporality,
	)
}

func NewFaaSTimeoutsMetricBuilder(startTime pcommon.Timestamp, temporality pmetric.AggregationTemporality) *CounterMetricBuilder {
	return NewCounterMetricBuilder(
		semconv2.FaaSTimeoutsName,
		semconv2.FaaSTimeoutsDescription,
		semconv2.FaaSTimeoutsUnit,
		true,
		startTime,
		temporality,
	)
}

type FaaSMetricBuilders struct {
	invokeDurationMetric *HistogramMetricBuilder
	initDurationMetric   *HistogramMetricBuilder
	memUsageMetric       *HistogramMetricBuilder
	coldstartsMetric     *CounterMetricBuilder
	errorsMetric         *CounterMetricBuilder
	invocationsMetric    *CounterMetricBuilder
	timeoutsMetric       *CounterMetricBuilder
}

func NewFaaSMetricBuilders(startTime pcommon.Timestamp, temporality pmetric.AggregationTemporality) *FaaSMetricBuilders {
	return &FaaSMetricBuilders{
		invokeDurationMetric: NewFasSInvokeDurationMetricBuilder(startTime, temporality),
		initDurationMetric:   NewFasSInitDurationMetricBuilder(startTime, temporality),
		memUsageMetric:       NewFaaSMemUsageMetricBuilder(startTime, temporality),
		coldstartsMetric:     NewFaaSColdstartsMetricBuilder(startTime, temporality),
		errorsMetric:         NewFaaSErrorsMetricBuilder(startTime, temporality),
		invocationsMetric:    NewFaaSInvocationsMetricBuilder(startTime, temporality),
		timeoutsMetric:       NewFaaSTimeoutsMetricBuilder(startTime, temporality),
	}
}
