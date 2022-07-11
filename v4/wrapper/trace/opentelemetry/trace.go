package opentelemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	instrumentationName = "github.com/go-micro/plugins/v4/wrapper/trace/opentelemetry"
)

type traceContextKey struct{}

// StartSpanFromContext returns a new span with the given operation name and options. If a span
// is found in the context, it will be used as the parent of the resulting span.
func StartSpanFromContext(ctx context.Context, tp trace.TracerProvider, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	carrier, ok := ctx.Value(traceContextKey{}).(propagation.MapCarrier)
	if !ok {
		carrier = propagation.MapCarrier{}
	}
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	var tracer trace.Tracer
	if tp != nil {
		tracer = tp.Tracer(instrumentationName)
	} else {
		tracer = otel.Tracer(instrumentationName)
	}
	return tracer.Start(ctx, name, opts...)
}
