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

package faasprocessor // import "github.com/open-telemetry/opentelemetry-lambda/collector/processor/faasprocessor"

import (
	"context"

	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
	semconv "go.opentelemetry.io/collector/semconv/v1.22.0"
	"go.uber.org/zap"
)

const (
	telemetryAPIScope = "github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapi"
)

type faasProcessor struct {
	telemetryAPIRuntimeSpans map[string]cachedSpan
	telemetryAPIInitSpans    map[string]cachedSpan
	invocationSpans          map[string]cachedSpan
	logger                   *zap.Logger
	nextConsumer             consumer.Traces
}

func (p *faasProcessor) processTraces(ctx context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	// Remove spans which ought to be matched on the request ID
	td.ResourceSpans().RemoveIf(func(rs ptrace.ResourceSpans) bool {
		resource := rs.Resource()
		rs.ScopeSpans().RemoveIf(func(ss ptrace.ScopeSpans) bool {
			scope := ss.Scope()
			ss.Spans().RemoveIf(func(span ptrace.Span) bool {
				if scope.Name() == telemetryAPIScope {
					if requestID, ok := span.Attributes().Get(semconv.AttributeFaaSInvocationID); ok {
						// This span was issued from the telemetry api, let's cache it for later
						cached := cachedSpan{
							resource: resource,
							scope:    scope,
							span:     span,
						}
						if _, ok := span.Attributes().Get(semconv.AttributeFaaSColdstart); ok {
							// This is the "root" runtime span
							p.telemetryAPIRuntimeSpans[requestID.Str()] = cached
						} else {
							// This is the init span that is a child of the "root" span
							p.telemetryAPIInitSpans[requestID.Str()] = cached
						}
						return true
					}
				} else {
					if requestID, ok := span.Attributes().Get(semconv.AttributeFaaSInvocationID); ok {
						// This span was created by the invoked code, let's cache it and don't
						// process it further here.
						p.invocationSpans[requestID.Str()] = cachedSpan{
							resource: resource,
							scope:    scope,
							span:     span,
						}
						return true
					}
				}
				return false
			})
			return ss.Spans().Len() == 0
		})
		return rs.ScopeSpans().Len() == 0
	})

	// Check for matches on the request ID and add new spans
	for requestID, telemetryApiSpan := range p.telemetryAPIRuntimeSpans {
		if invocationSpan, ok := p.invocationSpans[requestID]; ok {
			// Augment the spans as required
			telemetryApiSpan.span.SetParentSpanID(invocationSpan.span.ParentSpanID())
			telemetryApiSpan.span.SetTraceID(invocationSpan.span.TraceID())
			invocationSpan.span.SetParentSpanID(telemetryApiSpan.span.SpanID())

			// Add spans to the output
			telemetryApiSpan.addToTraces(td)
			invocationSpan.addToTraces(td)

			// Clean up our cache
			delete(p.telemetryAPIRuntimeSpans, requestID)
			delete(p.invocationSpans, requestID)

			// Optionally, there is also an init span. We need to augment its trace ID
			if initSpan, ok := p.telemetryAPIInitSpans[requestID]; ok {
				initSpan.span.SetTraceID(invocationSpan.span.TraceID())
				initSpan.addToTraces(td)
				delete(p.telemetryAPIInitSpans, requestID)
			}
		}
	}

	if td.ResourceSpans().Len() == 0 {
		return td, processorhelper.ErrSkipProcessingData
	}
	return td, nil
}

func newFaasProcessor(
	cfg *Config,
	next consumer.Traces,
	set processor.CreateSettings,
) (*faasProcessor, error) {
	return &faasProcessor{
		telemetryAPIRuntimeSpans: make(map[string]cachedSpan),
		telemetryAPIInitSpans:    make(map[string]cachedSpan),
		invocationSpans:          make(map[string]cachedSpan),
		nextConsumer:             next,
		logger:                   set.Logger,
	}, nil
}

/* ---------------------------------------- CACHED SPAN ---------------------------------------- */

type cachedSpan struct {
	resource pcommon.Resource
	scope    pcommon.InstrumentationScope
	span     ptrace.Span
}

func (s cachedSpan) addToTraces(td ptrace.Traces) {
	resourceSpans := td.ResourceSpans().AppendEmpty()
	s.resource.CopyTo(resourceSpans.Resource())

	scopeSpans := resourceSpans.ScopeSpans().AppendEmpty()
	s.scope.CopyTo(scopeSpans.Scope())

	span := scopeSpans.Spans().AppendEmpty()
	s.span.CopyTo(span)
}
