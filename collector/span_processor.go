package main

import (
	"context"
	"sync"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// spanProcessor is an sdktrace.spanProcessor implementation that exposes zpages functionality for opentelemetry-go.
//
// It tracks all active spans, and stores samples of spans based on latency for non errored spans,
// and samples for errored spans.
type spanProcessor struct {
	// Cannot keep track of the active Spans per name because the Span interface,
	// allows the name to be changed, and that will leak memory.
	activeSpansStore sync.Map
}

// newSpanProcessor returns a new spanProcessor.
func newSpanProcessor() *spanProcessor {
	return &spanProcessor{}
}

// OnStart adds span as active and reports it with zpages.
func (ssm *spanProcessor) OnStart(_ context.Context, span sdktrace.ReadWriteSpan) {
	sc := span.SpanContext()
	if sc.IsValid() {
		ssm.activeSpansStore.Store(spanKey(sc), span)
	}
}

// OnEnd processes all spans and reports them with zpages.
func (ssm *spanProcessor) OnEnd(span sdktrace.ReadOnlySpan) {
	sc := span.SpanContext()
	if sc.IsValid() {
		ssm.activeSpansStore.Delete(spanKey(sc))
	}
}

// Shutdown does nothing.
func (ssm *spanProcessor) Shutdown(context.Context) error {
	// Do nothing
	return nil
}

// ForceFlush does nothing.
func (ssm *spanProcessor) ForceFlush(context.Context) error {
	// Do nothing
	return nil
}

func (ssm *spanProcessor) activeSpanCount() int {
	i := 0
	ssm.activeSpansStore.Range(func(_, _ any) bool {
		i++
		return true
	})
	return i
}

func spanKey(sc trace.SpanContext) [24]byte {
	var sk [24]byte
	tid := sc.TraceID()
	copy(sk[0:16], tid[:])
	sid := sc.SpanID()
	copy(sk[16:24], sid[:])
	return sk
}
