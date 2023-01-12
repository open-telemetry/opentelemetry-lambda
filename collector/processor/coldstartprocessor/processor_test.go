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
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"go.opentelemetry.io/collector/processor/processortest"
	semconv "go.opentelemetry.io/collector/semconv/v1.5.0"
)

type mockConsumer struct {
	consumed int
}

func (c *mockConsumer) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	c.consumed += td.SpanCount()
	return nil
}

func (c *mockConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func addExecutionSpan(span ptrace.Span) {
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	randSource := rand.New(rand.NewSource(rngSeed))
	tid := pcommon.TraceID{}
	_, _ = randSource.Read(tid[:])
	span.SetTraceID(tid)
	span.Attributes().PutStr(semconv.AttributeFaaSExecution, "af9d5aa4-a685-4c5f-a22b-444f80b3cc28")
}

func TestProcessor(t *testing.T) {
	testCases := []struct {
		desc          string
		input         ptrace.Traces
		expected      ptrace.Traces
		expectedError error
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
			expected: func() ptrace.Traces {
				td := ptrace.NewTraces()
				td.ResourceSpans().AppendEmpty()
				td.ResourceSpans().RemoveIf(func(rs ptrace.ResourceSpans) bool { return true })
				return td
			}(),
			expectedError: processorhelper.ErrSkipProcessingData,
		},
		{
			desc: "execution without coldstart",
			input: func() ptrace.Traces {
				td := ptrace.NewTraces()
				addExecutionSpan(td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty())
				return td
			}(),
			expected: func() ptrace.Traces {
				td := ptrace.NewTraces()
				addExecutionSpan(td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty())
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
				addExecutionSpan(td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty())
				return td
			}(),
			expected: func() ptrace.Traces {
				td := ptrace.NewTraces()
				ss := td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty()
				addExecutionSpan(ss.Spans().AppendEmpty())
				span := ss.Spans().AppendEmpty()
				span.Attributes().PutBool(semconv.AttributeFaaSColdstart, true)
				span.Attributes().PutBool("faas.initialization", true)
				return td
			}(),
		},
		{
			desc: "faas.execution and faas.coldstart with execution is first",
			input: func() ptrace.Traces {
				td := ptrace.NewTraces()
				addExecutionSpan(td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty())
				span := td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
				span.Attributes().PutBool(semconv.AttributeFaaSColdstart, true)
				span.Attributes().PutBool("faas.initialization", true)
				return td
			}(),
			expected: func() ptrace.Traces {
				td := ptrace.NewTraces()
				addExecutionSpan(td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty())
				span := td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
				span.Attributes().PutBool(semconv.AttributeFaaSColdstart, true)
				span.Attributes().PutBool("faas.initialization", true)
				return td
			}(),
		},
	}
	// test case where no spans containing either attributes is there
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := newColdstartProcessor(
				nil,
				nil,
				processortest.NewNopCreateSettings(),
			)
			require.NoError(t, err)
			td, err := c.processTraces(context.Background(), tc.input)
			if tc.expectedError != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.expected.SpanCount(), td.SpanCount())
		})
	}
}
