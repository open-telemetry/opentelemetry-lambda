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

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatautil"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

const MiB = float64(1 << 20)
const GiB = float64(1 << 30)

var DefaultHistogramBounds = []float64{0.0, 5.0, 10.0, 25.0, 50.0, 75.0, 100.0, 250.0, 500.0, 750.0, 1000.0, 2500.0, 5000.0, 7500.0, 10000.0}
var DurationHistogramBounds = []float64{0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10}
var MemUsageHistogramBounds = []float64{16 * MiB, 32 * MiB, 64 * MiB, 128 * MiB, 256 * MiB, 512 * MiB, 768 * MiB, 1 * GiB, 2 * GiB, 3 * GiB, 4 * GiB, 6 * GiB, 8 * GiB}

type histogramDataPoint struct {
	attributes  pcommon.Map
	counts      []uint64
	total       uint64
	sum         float64
	startTime   pcommon.Timestamp
	lastUpdated uint64 // epoch when this data point was last updated
}

func newHistogramDataPoint(attrs pcommon.Map, numBuckets int, startTime pcommon.Timestamp, epoch uint64) *histogramDataPoint {
	return &histogramDataPoint{
		attributes:  attrs,
		counts:      make([]uint64, numBuckets),
		startTime:   startTime,
		lastUpdated: epoch,
	}
}

type HistogramMetricBuilder struct {
	name        string
	description string
	unit        string
	bounds      []float64
	dataPoints  map[[16]byte]*histogramDataPoint
	startTime   pcommon.Timestamp
	temporality pmetric.AggregationTemporality
	epoch       uint64 // current epoch counter
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

	return &HistogramMetricBuilder{
		name:        name,
		description: description,
		unit:        unit,
		bounds:      b,
		dataPoints:  make(map[[16]byte]*histogramDataPoint),
		startTime:   startTime,
		temporality: temp,
		epoch:       0,
	}
}

func (h *HistogramMetricBuilder) Record(value float64) {
	h.RecordWithAttributes(value, pcommon.NewMap())
}

func (h *HistogramMetricBuilder) RecordWithAttributes(value float64, attrs pcommon.Map) {
	key := pdatautil.MapHash(attrs)
	dp, exists := h.dataPoints[key]
	if !exists {
		dp = newHistogramDataPoint(attrs, len(h.bounds)+1, h.startTime, h.epoch)
		h.dataPoints[key] = dp
	}

	dp.sum += value
	dp.total++
	dp.counts[sort.SearchFloat64s(h.bounds, value)]++
	dp.lastUpdated = h.epoch
}

func (h *HistogramMetricBuilder) RecordWithMap(value float64, attrs map[string]any) error {
	m := pcommon.NewMap()
	err := m.FromRaw(attrs)
	if err != nil {
		return err
	}
	h.RecordWithAttributes(value, m)
	return nil
}

func (h *HistogramMetricBuilder) Reset(timestamp pcommon.Timestamp) {
	h.startTime = timestamp
	clear(h.dataPoints)
	h.epoch = 0
}

func (h *HistogramMetricBuilder) AppendDataPoints(scopeMetrics pmetric.ScopeMetrics, timestamp pcommon.Timestamp) {
	export := h.temporality == pmetric.AggregationTemporalityDelta
	if !export {
		for _, hdp := range h.dataPoints {
			if hdp.lastUpdated == h.epoch {
				export = true
				break
			}
		}
	}

	if !export {
		return
	}

	metric := scopeMetrics.Metrics().AppendEmpty()
	metric.SetName(h.name)
	metric.SetDescription(h.description)
	metric.SetUnit(h.unit)

	hist := metric.SetEmptyHistogram()
	hist.SetAggregationTemporality(h.temporality)

	for _, hdp := range h.dataPoints {
		// For cumulative: only export if updated in current epoch
		if h.temporality == pmetric.AggregationTemporalityCumulative && hdp.lastUpdated != h.epoch {
			continue
		}

		dp := hist.DataPoints().AppendEmpty()
		hdp.attributes.CopyTo(dp.Attributes())
		dp.SetStartTimestamp(hdp.startTime)
		dp.SetTimestamp(timestamp)
		dp.SetSum(hdp.sum)
		dp.SetCount(hdp.total)
		dp.BucketCounts().FromRaw(hdp.counts)
		dp.ExplicitBounds().FromRaw(h.bounds)
	}

	if h.temporality == pmetric.AggregationTemporalityDelta {
		h.Reset(timestamp)
	} else {
		// For cumulative, increment epoch for next collection cycle
		h.epoch++
	}
}

type counterDataPoint struct {
	attributes  pcommon.Map
	total       int64
	startTime   pcommon.Timestamp
	lastUpdated uint64 // epoch when this data point was last updated
}

func newCounterDataPoint(attrs pcommon.Map, startTime pcommon.Timestamp, epoch uint64) *counterDataPoint {
	return &counterDataPoint{
		attributes:  attrs,
		startTime:   startTime,
		lastUpdated: epoch,
	}
}

type CounterMetricBuilder struct {
	name        string
	description string
	unit        string
	dataPoints  map[[16]byte]*counterDataPoint
	isMonotonic bool
	temporality pmetric.AggregationTemporality
	startTime   pcommon.Timestamp
	epoch       uint64 // current epoch counter
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
		dataPoints:  make(map[[16]byte]*counterDataPoint),
		isMonotonic: isMonotonic,
		temporality: temp,
		startTime:   startTime,
		epoch:       0,
	}
}

func (c *CounterMetricBuilder) Add(value int64) {
	c.AddWithAttributes(value, pcommon.NewMap())
}

func (c *CounterMetricBuilder) AddWithAttributes(value int64, attrs pcommon.Map) {
	key := pdatautil.MapHash(attrs)
	dp, exists := c.dataPoints[key]
	if !exists {
		dp = newCounterDataPoint(attrs, c.startTime, c.epoch)
		c.dataPoints[key] = dp
	}
	dp.total += value
	dp.lastUpdated = c.epoch
}

func (c *CounterMetricBuilder) AddWithMap(value int64, attrs map[string]any) error {
	m := pcommon.NewMap()
	err := m.FromRaw(attrs)
	if err != nil {
		return err
	}
	c.AddWithAttributes(value, m)
	return nil
}

func (c *CounterMetricBuilder) Reset(timestamp pcommon.Timestamp) {
	c.startTime = timestamp
	clear(c.dataPoints)
	c.epoch = 0
}

func (c *CounterMetricBuilder) AppendDataPoints(scopeMetrics pmetric.ScopeMetrics, timestamp pcommon.Timestamp) {
	export := c.temporality == pmetric.AggregationTemporalityDelta
	if !export {
		for _, cdp := range c.dataPoints {
			if cdp.lastUpdated == c.epoch {
				export = true
				break
			}
		}
	}

	if !export {
		return
	}

	metric := scopeMetrics.Metrics().AppendEmpty()
	metric.SetName(c.name)
	metric.SetDescription(c.description)
	metric.SetUnit(c.unit)

	sum := metric.SetEmptySum()
	sum.SetAggregationTemporality(c.temporality)
	sum.SetIsMonotonic(c.isMonotonic)

	for _, cdp := range c.dataPoints {
		// For cumulative: only export if updated in current epoch
		if c.temporality == pmetric.AggregationTemporalityCumulative && cdp.lastUpdated != c.epoch {
			continue
		}

		dp := sum.DataPoints().AppendEmpty()
		cdp.attributes.CopyTo(dp.Attributes())
		dp.SetStartTimestamp(cdp.startTime)
		dp.SetTimestamp(timestamp)
		dp.SetIntValue(cdp.total)
	}

	if c.temporality == pmetric.AggregationTemporalityDelta {
		c.Reset(timestamp)
	} else {
		// For cumulative, increment epoch for next collection cycle
		c.epoch++
	}
}

func NewFaaSInvokeDurationMetricBuilder(startTime pcommon.Timestamp, temporality pmetric.AggregationTemporality) *HistogramMetricBuilder {
	return NewHistogramMetricBuilder(
		semconv.FaaSInvokeDurationName,
		semconv.FaaSInvokeDurationDescription,
		semconv.FaaSInvokeDurationUnit,
		DurationHistogramBounds,
		startTime,
		temporality,
	)
}

func NewFaaSInitDurationMetricBuilder(startTime pcommon.Timestamp, temporality pmetric.AggregationTemporality) *HistogramMetricBuilder {
	return NewHistogramMetricBuilder(
		semconv.FaaSInitDurationName,
		semconv.FaaSInitDurationDescription,
		semconv.FaaSInitDurationUnit,
		DurationHistogramBounds,
		startTime,
		temporality,
	)
}

func NewFaaSMemUsageMetricBuilder(startTime pcommon.Timestamp, temporality pmetric.AggregationTemporality) *HistogramMetricBuilder {
	return NewHistogramMetricBuilder(
		semconv.FaaSMemUsageName,
		semconv.FaaSMemUsageDescription,
		semconv.FaaSMemUsageUnit,
		MemUsageHistogramBounds,
		startTime,
		temporality,
	)
}

func NewFaaSColdstartsMetricBuilder(startTime pcommon.Timestamp, temporality pmetric.AggregationTemporality) *CounterMetricBuilder {
	return NewCounterMetricBuilder(
		semconv.FaaSColdstartsName,
		semconv.FaaSColdstartsDescription,
		semconv.FaaSColdstartsUnit,
		true,
		startTime,
		temporality,
	)
}

func NewFaaSErrorsMetricBuilder(startTime pcommon.Timestamp, temporality pmetric.AggregationTemporality) *CounterMetricBuilder {
	return NewCounterMetricBuilder(
		semconv.FaaSErrorsName,
		semconv.FaaSErrorsDescription,
		semconv.FaaSErrorsUnit,
		true,
		startTime,
		temporality,
	)
}

func NewFaaSInvocationsMetricBuilder(startTime pcommon.Timestamp, temporality pmetric.AggregationTemporality) *CounterMetricBuilder {
	return NewCounterMetricBuilder(
		semconv.FaaSInvocationsName,
		semconv.FaaSInvocationsDescription,
		semconv.FaaSInvocationsUnit,
		true,
		startTime,
		temporality,
	)
}

func NewFaaSTimeoutsMetricBuilder(startTime pcommon.Timestamp, temporality pmetric.AggregationTemporality) *CounterMetricBuilder {
	return NewCounterMetricBuilder(
		semconv.FaaSTimeoutsName,
		semconv.FaaSTimeoutsDescription,
		semconv.FaaSTimeoutsUnit,
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
		invokeDurationMetric: NewFaaSInvokeDurationMetricBuilder(startTime, temporality),
		initDurationMetric:   NewFaaSInitDurationMetricBuilder(startTime, temporality),
		memUsageMetric:       NewFaaSMemUsageMetricBuilder(startTime, temporality),
		coldstartsMetric:     NewFaaSColdstartsMetricBuilder(startTime, temporality),
		errorsMetric:         NewFaaSErrorsMetricBuilder(startTime, temporality),
		invocationsMetric:    NewFaaSInvocationsMetricBuilder(startTime, temporality),
		timeoutsMetric:       NewFaaSTimeoutsMetricBuilder(startTime, temporality),
	}
}
