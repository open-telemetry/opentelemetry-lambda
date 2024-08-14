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

	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
	semconv "go.opentelemetry.io/collector/semconv/v1.5.0"
	"go.uber.org/zap"
)

type faasExecution struct {
	span     ptrace.Span
	scope    pcommon.InstrumentationScope
	resource pcommon.Resource
}

type coldstartProcessor struct {
	coldstartSpan *ptrace.Span
	faasExecution *faasExecution
	logger        *zap.Logger
	nextConsumer  consumer.Traces
	reported      bool // whether the cold start has already been reported
}

func (p *coldstartProcessor) processTraces(ctx context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	if p.reported {
		return td, nil
	}
	td.ResourceSpans().RemoveIf(func(rs ptrace.ResourceSpans) bool {
		resource := rs.Resource()
		rs.ScopeSpans().RemoveIf(func(ss ptrace.ScopeSpans) bool {
			scope := ss.Scope()
			ss.Spans().RemoveIf(func(span ptrace.Span) bool {
				if p.reported {
					return false
				}
				if attr, ok := span.Attributes().Get(semconv.AttributeFaaSColdstart); ok && attr.Bool() {
					if p.faasExecution == nil {
						sp := ptrace.NewSpan()
						p.coldstartSpan = &sp
						span.CopyTo(*p.coldstartSpan)
						return true
					} else {
						p.faasExecution.scope.CopyTo(scope)
						p.faasExecution.resource.CopyTo(resource)
						span.SetParentSpanID(p.faasExecution.span.ParentSpanID())
						span.SetTraceID(p.faasExecution.span.TraceID())
						p.reported = true
						return false
					}
				}
				if _, ok := span.Attributes().Get(semconv.AttributeFaaSExecution); ok {
					if p.coldstartSpan == nil {
						p.faasExecution = &faasExecution{
							span:     ptrace.NewSpan(),
							scope:    pcommon.NewInstrumentationScope(),
							resource: pcommon.NewResource(),
						}

						scope.CopyTo(p.faasExecution.scope)
						resource.CopyTo(p.faasExecution.resource)
						span.CopyTo(p.faasExecution.span)
					} else {
						s := ss.Spans().AppendEmpty()
						p.coldstartSpan.CopyTo(s)
						s.SetParentSpanID(span.ParentSpanID())
						s.SetTraceID(span.TraceID())
						p.reported = true
						p.coldstartSpan = nil
					}
				}
				return false
			})
			return ss.Spans().Len() == 0
		})
		return rs.ScopeSpans().Len() == 0
	})

	if td.ResourceSpans().Len() == 0 {
		return td, processorhelper.ErrSkipProcessingData
	}
	return td, nil
}

func newColdstartProcessor(
	cfg *Config,
	next consumer.Traces,
	set processor.Settings,
) (*coldstartProcessor, error) {
	return &coldstartProcessor{
		nextConsumer: next,
		logger:       set.Logger,
	}, nil
}
