package telemetryapireceiver

import (
	"errors"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// NOTE: The transformations in this file are *incomplete* for general use and should, thus,
//       not be relied upon in contexts other than this package.

var errUnsupportedInstrumentType = errors.New("instrument type is currently unsupported")

func transformMetric(src metricdata.Metrics, dst pmetric.Metric) error {
	dst.SetName(src.Name)
	dst.SetDescription(src.Description)
	dst.SetUnit(src.Unit)
	switch data := src.Data.(type) {
	// case metricdata.ExponentialHistogram[float64]:
	// 	instrument := dst.SetEmptyExponentialHistogram()
	// 	transformExponentialHistogram(data, instrument)
	// case metricdata.ExponentialHistogram[int64]:
	// 	instrument := dst.SetEmptyExponentialHistogram()
	// 	transformExponentialHistogram(data, instrument)
	case metricdata.Sum[int64]:
		instrument := dst.SetEmptySum()
		transformCounterInt(data, instrument)
	case metricdata.Sum[float64]:
		instrument := dst.SetEmptySum()
		transformCounterFloat(data, instrument)
	default:
		break
		// return errUnsupportedInstrumentType
	}
	return nil
}

/* ---------------------------------------- INSTRUMENTS ---------------------------------------- */

func transformExponentialHistogram[N int64 | float64](
	src metricdata.ExponentialHistogram[N],
	dst pmetric.ExponentialHistogram,
) {
	dst.SetAggregationTemporality(mapTemporality(src.Temporality))
	for _, datapoint := range src.DataPoints {
		dp := dst.DataPoints().AppendEmpty()
		dp.SetCount(datapoint.Count)
		if v, ok := datapoint.Max.Value(); ok {
			dp.SetMax(float64(v))
		}
		if v, ok := datapoint.Max.Value(); ok {
			dp.SetMin(float64(v))
		}
		dp.SetScale(datapoint.Scale)
		dp.SetStartTimestamp(pcommon.NewTimestampFromTime(datapoint.StartTime))
		dp.SetSum(float64(datapoint.Sum))
		dp.SetTimestamp(pcommon.NewTimestampFromTime(datapoint.Time))
		dp.SetZeroCount(datapoint.ZeroCount)
		dp.SetZeroThreshold(datapoint.ZeroThreshold)
		dp.Negative().SetOffset(datapoint.NegativeBucket.Offset)
		dp.Negative().BucketCounts().Append(datapoint.NegativeBucket.Counts...)
		dp.Positive().SetOffset(datapoint.PositiveBucket.Offset)
		dp.Positive().BucketCounts().Append(datapoint.PositiveBucket.Counts...)
	}
}

func transformCounterInt(src metricdata.Sum[int64], dst pmetric.Sum) {
	dst.SetAggregationTemporality(mapTemporality(src.Temporality))
	dst.SetIsMonotonic(src.IsMonotonic)
	for _, datapoint := range src.DataPoints {
		dp := dst.DataPoints().AppendEmpty()
		dp.SetIntValue(datapoint.Value)
		dp.SetStartTimestamp(pcommon.NewTimestampFromTime(datapoint.StartTime))
		dp.SetTimestamp(pcommon.NewTimestampFromTime(datapoint.Time))
	}
}

func transformCounterFloat(src metricdata.Sum[float64], dst pmetric.Sum) {
	dst.SetAggregationTemporality(mapTemporality(src.Temporality))
	dst.SetIsMonotonic(src.IsMonotonic)
	for _, datapoint := range src.DataPoints {
		dp := dst.DataPoints().AppendEmpty()
		dp.SetDoubleValue(datapoint.Value)
		dp.SetStartTimestamp(pcommon.NewTimestampFromTime(datapoint.StartTime))
		dp.SetTimestamp(pcommon.NewTimestampFromTime(datapoint.Time))
	}
}

/* ------------------------------------------- UTILS ------------------------------------------- */

func mapTemporality(t metricdata.Temporality) pmetric.AggregationTemporality {
	switch t {
	case metricdata.CumulativeTemporality:
		return pmetric.AggregationTemporalityCumulative
	case metricdata.DeltaTemporality:
		return pmetric.AggregationTemporalityDelta
	default:
		return pmetric.AggregationTemporalityUnspecified
	}
}
