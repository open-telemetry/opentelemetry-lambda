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

package coldstartprocessor // import "github.com/open-telemetry/opentelemetry-lambda/collector/processor/coldstartprocessor"

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"hash"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"github.com/cespare/xxhash"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"go.opentelemetry.io/collector/processor/processortest"
	semconv "go.opentelemetry.io/collector/semconv/v1.5.0"
	"go.uber.org/multierr"
)

func TestProcessor(t *testing.T) {
	executionTraceID := getTraceID()
	testCases := []struct {
		desc          string
		input         ptrace.Traces
		expected      ptrace.Traces
		expectedError error
		reported      bool
	}{
		{
			desc:          "no traces",
			input:         ptrace.NewTraces(),
			expected:      ptrace.NewTraces(),
			expectedError: processorhelper.ErrSkipProcessingData,
		},
		{
			desc: "coldstart without execution",
			input: func() ptrace.Traces {
				td := ptrace.NewTraces()
				span := td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
				span.Attributes().PutBool(semconv.AttributeFaaSColdstart, true)
				return td
			}(),
			expected:      ptrace.NewTraces(),
			expectedError: processorhelper.ErrSkipProcessingData,
		},
		{
			desc: "execution without coldstart",
			input: func() ptrace.Traces {
				td := ptrace.NewTraces()
				addExecutionSpan(td, executionTraceID)
				return td
			}(),
			expected: func() ptrace.Traces {
				td := ptrace.NewTraces()
				addExecutionSpan(td, executionTraceID)
				return td
			}(),
		},
		{
			desc: "faas.execution and faas.coldstart with coldstart is first",
			input: func() ptrace.Traces {
				td := ptrace.NewTraces()
				span := td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
				span.Attributes().PutBool(semconv.AttributeFaaSColdstart, true)
				span.Attributes().PutBool("faas.initialization", true)
				addExecutionSpan(td, executionTraceID)
				return td
			}(),
			expected: func() ptrace.Traces {
				td := ptrace.NewTraces()
				rs := td.ResourceSpans().AppendEmpty()
				ss := rs.ScopeSpans().AppendEmpty()

				rs.Resource().Attributes().PutStr("resource-attr", "faas-execution")
				ss.Scope().SetName("app/execution")
				executionSpan(ss.Spans().AppendEmpty(), executionTraceID)
				initializationSpan(ss.Spans().AppendEmpty(), executionTraceID)
				return td
			}(),
			reported: true,
		},
		{
			desc: "faas.execution and faas.coldstart with execution is first",
			input: func() ptrace.Traces {
				td := ptrace.NewTraces()
				addExecutionSpan(td, executionTraceID)
				span := td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
				span.Attributes().PutBool(semconv.AttributeFaaSColdstart, true)
				span.Attributes().PutBool("faas.initialization", true)
				return td
			}(),
			expected: func() ptrace.Traces {
				td := ptrace.NewTraces()
				rs := td.ResourceSpans().AppendEmpty()
				ss := rs.ScopeSpans().AppendEmpty()
				rs.Resource().Attributes().PutStr("resource-attr", "faas-execution")
				ss.Scope().SetName("app/execution")
				executionSpan(ss.Spans().AppendEmpty(), executionTraceID)
				rs = td.ResourceSpans().AppendEmpty()
				ss = rs.ScopeSpans().AppendEmpty()
				rs.Resource().Attributes().PutStr("resource-attr", "faas-execution")
				ss.Scope().SetName("app/execution")
				initializationSpan(ss.Spans().AppendEmpty(), executionTraceID)
				return td
			}(),
			reported: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := newColdstartProcessor(
				nil,
				nil,
				processortest.NewNopSettings(Type),
			)
			require.NoError(t, err)
			td, err := c.processTraces(context.Background(), tc.input)
			require.Equal(t, tc.expectedError, err)

			require.Equal(t, tc.expected.SpanCount(), td.SpanCount())
			require.Equal(t, tc.reported, c.reported)
			require.NoError(t, compareTraces(tc.expected, td))
		})
	}
}

func TestMultipleProcessTraces(t *testing.T) {
	c, err := newColdstartProcessor(
		nil,
		nil,
		processortest.NewNopSettings(Type),
	)
	require.NoError(t, err)
	expected := ptrace.NewTraces()
	input := ptrace.NewTraces()
	addExecutionSpan(input, getTraceID())
	input.CopyTo(expected)
	output, err := c.processTraces(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 1, output.SpanCount())
	require.NoError(t, compareTraces(expected, output))
	require.False(t, c.reported)

	input = ptrace.NewTraces()
	expected = ptrace.NewTraces()
	span := input.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	span.Attributes().PutBool(semconv.AttributeFaaSColdstart, true)
	span.Attributes().PutBool("faas.initialization", true)
	input.CopyTo(expected)
	output, err = c.processTraces(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, expected.SpanCount(), output.SpanCount())
	require.Error(t, compareTraces(expected, output))
	attr, ok := output.ResourceSpans().At(0).Resource().Attributes().Get("resource-attr")
	require.True(t, ok)
	require.Equal(t, "faas-execution", attr.AsString())
	require.Equal(t, "app/execution", output.ResourceSpans().At(0).ScopeSpans().At(0).Scope().Name())
	require.True(t, c.reported)

	c, err = newColdstartProcessor(
		nil,
		nil,
		processortest.NewNopSettings(Type),
	)
	require.NoError(t, err)
	input = ptrace.NewTraces()
	expected = ptrace.NewTraces()
	span = input.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	span.Attributes().PutBool(semconv.AttributeFaaSColdstart, true)
	span.Attributes().PutBool("faas.initialization", true)
	input.CopyTo(expected)
	output, err = c.processTraces(context.Background(), input)
	require.Error(t, err)
	require.Equal(t, 0, output.SpanCount())
	require.False(t, c.reported)

	expected = ptrace.NewTraces()
	input = ptrace.NewTraces()
	addExecutionSpan(input, getTraceID())
	input.CopyTo(expected)
	output, err = c.processTraces(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 2, output.SpanCount())
	require.Error(t, compareTraces(expected, output))
	attr, ok = output.ResourceSpans().At(0).Resource().Attributes().Get("resource-attr")
	require.True(t, ok)
	require.Equal(t, "faas-execution", attr.AsString())
	require.Equal(t, "app/execution", output.ResourceSpans().At(0).ScopeSpans().At(0).Scope().Name())
	require.True(t, c.reported)
}

func getTraceID() pcommon.TraceID {
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	randSource := rand.New(rand.NewSource(rngSeed))
	tid := pcommon.TraceID{}
	_, _ = randSource.Read(tid[:])
	return tid
}

func addExecutionSpan(td ptrace.Traces, id pcommon.TraceID) {
	rs := td.ResourceSpans().AppendEmpty()
	rs.Resource().Attributes().PutStr("resource-attr", "faas-execution")
	ss := rs.ScopeSpans().AppendEmpty()
	ss.Scope().SetName("app/execution")
	span := ss.Spans().AppendEmpty()
	span.SetTraceID(id)
	span.Attributes().PutStr(semconv.AttributeFaaSExecution, "af9d5aa4-a685-4c5f-a22b-444f80b3cc28")
}

func executionSpan(span ptrace.Span, id pcommon.TraceID) {
	span.SetTraceID(id)
	span.Attributes().PutStr(semconv.AttributeFaaSExecution, "af9d5aa4-a685-4c5f-a22b-444f80b3cc28")
}

func initializationSpan(span ptrace.Span, id pcommon.TraceID) {
	span.SetTraceID(id)
	span.Attributes().PutBool(semconv.AttributeFaaSColdstart, true)
	span.Attributes().PutBool("faas.initialization", true)
}

var (
	extraByte       = []byte{'\xf3'}
	keyPrefix       = []byte{'\xf4'}
	valEmpty        = []byte{'\xf5'}
	valBytesPrefix  = []byte{'\xf6'}
	valStrPrefix    = []byte{'\xf7'}
	valBoolTrue     = []byte{'\xf8'}
	valBoolFalse    = []byte{'\xf9'}
	valIntPrefix    = []byte{'\xfa'}
	valDoublePrefix = []byte{'\xfb'}
	valMapPrefix    = []byte{'\xfc'}
	valMapSuffix    = []byte{'\xfd'}
	valSlicePrefix  = []byte{'\xfe'}
	valSliceSuffix  = []byte{'\xff'}
)

// mapHash return a hash for the provided map.
// Maps with the same underlying key/value pairs in different order produce the same deterministic hash value.
func mapHash(m pcommon.Map) [16]byte {
	h := xxhash.New()
	writeMapHash(h, m)
	return hashSum128(h)
}

func writeMapHash(h hash.Hash, m pcommon.Map) {
	keys := make([]string, 0, m.Len())
	m.Range(func(k string, v pcommon.Value) bool {
		keys = append(keys, k)
		return true
	})
	sort.Strings(keys)
	for _, k := range keys {
		v, _ := m.Get(k)
		h.Write(keyPrefix)
		h.Write([]byte(k))
		writeValueHash(h, v)
	}
}

func writeSliceHash(h hash.Hash, sl pcommon.Slice) {
	for i := 0; i < sl.Len(); i++ {
		writeValueHash(h, sl.At(i))
	}
}

func writeValueHash(h hash.Hash, v pcommon.Value) {
	switch v.Type() {
	case pcommon.ValueTypeStr:
		h.Write(valStrPrefix)
		h.Write([]byte(v.Str()))
	case pcommon.ValueTypeBool:
		if v.Bool() {
			h.Write(valBoolTrue)
		} else {
			h.Write(valBoolFalse)
		}
	case pcommon.ValueTypeInt:
		h.Write(valIntPrefix)
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(v.Int()))
		h.Write(b)
	case pcommon.ValueTypeDouble:
		h.Write(valDoublePrefix)
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, math.Float64bits(v.Double()))
		h.Write(b)
	case pcommon.ValueTypeMap:
		h.Write(valMapPrefix)
		writeMapHash(h, v.Map())
		h.Write(valMapSuffix)
	case pcommon.ValueTypeSlice:
		h.Write(valSlicePrefix)
		writeSliceHash(h, v.Slice())
		h.Write(valSliceSuffix)
	case pcommon.ValueTypeBytes:
		h.Write(valBytesPrefix)
		h.Write(v.Bytes().AsRaw())
	case pcommon.ValueTypeEmpty:
		h.Write(valEmpty)
	}
}

// hashSum128 returns a [16]byte hash sum.
func hashSum128(h hash.Hash) [16]byte {
	b := make([]byte, 0, 16)
	b = h.Sum(b)

	// Append an extra byte to generate another part of the hash sum
	_, _ = h.Write(extraByte)
	b = h.Sum(b)

	res := [16]byte{}
	copy(res[:], b)
	return res
}

func sortResourceSpans(a, b ptrace.ResourceSpans) bool {
	if a.SchemaUrl() < b.SchemaUrl() {
		return true
	}
	aAttrs := mapHash(a.Resource().Attributes())
	bAttrs := mapHash(b.Resource().Attributes())
	return bytes.Compare(aAttrs[:], bAttrs[:]) < 0
}

func sortSpansInstrumentationLibrary(a, b ptrace.ScopeSpans) bool {
	if a.SchemaUrl() < b.SchemaUrl() {
		return true
	}
	if a.Scope().Name() < b.Scope().Name() {
		return true
	}
	if a.Scope().Version() < b.Scope().Version() {
		return true
	}
	return false
}

func sortSpanSlice(a, b ptrace.Span) bool {
	aAttrs := mapHash(a.Attributes())
	bAttrs := mapHash(b.Attributes())
	return bytes.Compare(aAttrs[:], bAttrs[:]) < 0
}

// CompareTraces compares each part of two given Traces and returns
// an error if they don't match. The error describes what didn't match.
func compareTraces(expected, actual ptrace.Traces) error {
	exp, act := ptrace.NewTraces(), ptrace.NewTraces()
	expected.CopyTo(exp)
	actual.CopyTo(act)

	expectedSpans, actualSpans := exp.ResourceSpans(), act.ResourceSpans()
	if expectedSpans.Len() != actualSpans.Len() {
		return fmt.Errorf("amount of ResourceSpans between Traces are not equal expected: %d, actual: %d",
			expectedSpans.Len(),
			actualSpans.Len())
	}

	// sort ResourceSpans
	expectedSpans.Sort(sortResourceSpans)
	actualSpans.Sort(sortResourceSpans)

	numResources := expectedSpans.Len()

	// Keep track of matching resources so that each can only be matched once
	matchingResources := make(map[ptrace.ResourceSpans]ptrace.ResourceSpans, numResources)

	var errs error
	for e := 0; e < numResources; e++ {
		er := expectedSpans.At(e)
		var foundMatch bool
		for a := 0; a < numResources; a++ {
			ar := actualSpans.At(a)
			if _, ok := matchingResources[ar]; ok {
				continue
			}
			if reflect.DeepEqual(er.Resource().Attributes().AsRaw(), ar.Resource().Attributes().AsRaw()) {
				foundMatch = true
				matchingResources[ar] = er
				break
			}
		}
		if !foundMatch {
			errs = multierr.Append(errs, fmt.Errorf("missing expected resource with attributes: %v", er.Resource().Attributes().AsRaw()))
		}
	}

	for i := 0; i < numResources; i++ {
		if _, ok := matchingResources[actualSpans.At(i)]; !ok {
			errs = multierr.Append(errs, fmt.Errorf("extra resource with attributes: %v", actualSpans.At(i).Resource().Attributes().AsRaw()))
		}
	}

	if errs != nil {
		return errs
	}

	for ar, er := range matchingResources {
		if err := CompareResourceSpans(er, ar); err != nil {
			return err
		}
	}

	return nil
}

// CompareResourceSpans compares each part of two given ResourceSpans and returns
// an error if they don't match. The error describes what didn't match.
func CompareResourceSpans(expected, actual ptrace.ResourceSpans) error {
	eilms := expected.ScopeSpans()
	ailms := actual.ScopeSpans()

	if eilms.Len() != ailms.Len() {
		return fmt.Errorf("number of instrumentation libraries does not match expected: %d, actual: %d", eilms.Len(),
			ailms.Len())
	}

	// sort InstrumentationLibrary
	eilms.Sort(sortSpansInstrumentationLibrary)
	ailms.Sort(sortSpansInstrumentationLibrary)

	for i := 0; i < eilms.Len(); i++ {
		eilm, ailm := eilms.At(i), ailms.At(i)
		eil, ail := eilm.Scope(), ailm.Scope()

		if eil.Name() != ail.Name() {
			return fmt.Errorf("instrumentation library Name does not match expected: %s, actual: %s", eil.Name(), ail.Name())
		}
		if eil.Version() != ail.Version() {
			return fmt.Errorf("instrumentation library Version does not match expected: %s, actual: %s", eil.Version(), ail.Version())
		}
		if err := CompareSpanSlices(eilm.Spans(), ailm.Spans()); err != nil {
			return err
		}
	}
	return nil
}

// CompareSpanSlices compares each part of two given SpanSlices and returns
// an error if they don't match. The error describes what didn't match.
func CompareSpanSlices(expected, actual ptrace.SpanSlice) error {
	if expected.Len() != actual.Len() {
		return fmt.Errorf("number of spans does not match expected: %d, actual: %d", expected.Len(), actual.Len())
	}

	expected.Sort(sortSpanSlice)
	actual.Sort(sortSpanSlice)

	numSpans := expected.Len()

	// Keep track of matching spans so that each span can only be matched once
	matchingSpans := make(map[ptrace.Span]ptrace.Span, numSpans)

	var errs error
	for e := 0; e < numSpans; e++ {
		elr := expected.At(e)
		var foundMatch bool
		for a := 0; a < numSpans; a++ {
			alr := actual.At(a)
			if _, ok := matchingSpans[alr]; ok {
				continue
			}
			if reflect.DeepEqual(elr.Attributes().AsRaw(), alr.Attributes().AsRaw()) {
				foundMatch = true
				matchingSpans[alr] = elr
				break
			}
		}
		if !foundMatch {
			errs = multierr.Append(errs, fmt.Errorf("span missing expected resource with attributes: %v", elr.Attributes().AsRaw()))
		}
	}

	for i := 0; i < numSpans; i++ {
		if _, ok := matchingSpans[actual.At(i)]; !ok {
			errs = multierr.Append(errs, fmt.Errorf("span has extra record with attributes: %v", actual.At(i).Attributes().AsRaw()))
		}
	}

	if errs != nil {
		return errs
	}

	for alr, elr := range matchingSpans {
		if err := CompareSpans(alr, elr); err != nil {
			return multierr.Combine(fmt.Errorf("span with attributes: %v, does not match expected %v", alr.Attributes().AsRaw(), elr.Attributes().AsRaw()), err)
		}
	}
	return nil
}

// CompareSpans compares each part of two given Span and returns
// an error if they don't match. The error describes what didn't match.
func CompareSpans(expected, actual ptrace.Span) error {
	if expected.TraceID() != actual.TraceID() {
		return fmt.Errorf("span TraceID doesn't match expected: %d, actual: %d",
			expected.TraceID(),
			actual.TraceID())
	}

	if expected.SpanID() != actual.SpanID() {
		return fmt.Errorf("span SpanID doesn't match expected: %d, actual: %d",
			expected.SpanID(),
			actual.SpanID())
	}

	if expected.TraceState().AsRaw() != actual.TraceState().AsRaw() {
		return fmt.Errorf("span TraceState doesn't match expected: %s, actual: %s",
			expected.TraceState().AsRaw(),
			actual.TraceState().AsRaw())
	}

	if expected.ParentSpanID() != actual.ParentSpanID() {
		return fmt.Errorf("span ParentSpanID doesn't match expected: %d, actual: %d",
			expected.ParentSpanID(),
			actual.ParentSpanID())
	}

	if expected.Name() != actual.Name() {
		return fmt.Errorf("span Name doesn't match expected: %s, actual: %s",
			expected.Name(),
			actual.Name())
	}

	if expected.Kind() != actual.Kind() {
		return fmt.Errorf("span Kind doesn't match expected: %d, actual: %d",
			expected.Kind(),
			actual.Kind())
	}

	if expected.StartTimestamp() != actual.StartTimestamp() {
		return fmt.Errorf("span StartTimestamp doesn't match expected: %d, actual: %d",
			expected.StartTimestamp(),
			actual.StartTimestamp())
	}

	if expected.EndTimestamp() != actual.EndTimestamp() {
		return fmt.Errorf("span EndTimestamp doesn't match expected: %d, actual: %d",
			expected.EndTimestamp(),
			actual.EndTimestamp())
	}

	if !reflect.DeepEqual(expected.Attributes().AsRaw(), actual.Attributes().AsRaw()) {
		return fmt.Errorf("span Attributes doesn't match expected: %s, actual: %s",
			expected.Attributes().AsRaw(),
			actual.Attributes().AsRaw())
	}

	if expected.DroppedAttributesCount() != actual.DroppedAttributesCount() {
		return fmt.Errorf("span DroppedAttributesCount doesn't match expected: %d, actual: %d",
			expected.DroppedAttributesCount(),
			actual.DroppedAttributesCount())
	}

	if !reflect.DeepEqual(expected.Events(), actual.Events()) {
		return fmt.Errorf("span Events doesn't match expected: %v, actual: %v",
			expected.Events(),
			actual.Events())
	}

	if expected.DroppedEventsCount() != actual.DroppedEventsCount() {
		return fmt.Errorf("span DroppedEventsCount doesn't match expected: %d, actual: %d",
			expected.DroppedEventsCount(),
			actual.DroppedEventsCount())
	}

	if !reflect.DeepEqual(expected.Links(), actual.Links()) {
		return fmt.Errorf("span Links doesn't match expected: %v, actual: %v",
			expected.Links(),
			actual.Links())
	}

	if expected.DroppedLinksCount() != actual.DroppedLinksCount() {
		return fmt.Errorf("span DroppedLinksCount doesn't match expected: %d, actual: %d",
			expected.DroppedLinksCount(),
			actual.DroppedLinksCount())
	}

	if !reflect.DeepEqual(expected.Status(), actual.Status()) {
		return fmt.Errorf("span Status doesn't match expected: %v, actual: %v",
			expected.Status(),
			actual.Status())
	}

	return nil
}
