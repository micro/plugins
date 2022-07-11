package opentelemetry

import (
	"context"
	"fmt"

	"go-micro.dev/v4/client"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/server"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// NewClientWrapper returns a client.Wrapper
// that adds monitoring to outgoing requests.
func NewClientWrapper(tracerProvider ...trace.TracerProvider) client.Wrapper {
	return func(c client.Client) client.Client {
		w := &clientWrapper{Client: c}
		if len(tracerProvider) > 0 {
			w.tp = tracerProvider[0]
		}
		return w
	}
}

// NewCallWrapper accepts an opentracing Tracer and returns a Call Wrapper
func NewCallWrapper(tracerProvider ...trace.TracerProvider) client.CallWrapper {
	return func(cf client.CallFunc) client.CallFunc {
		return func(ctx context.Context, node *registry.Node, req client.Request, rsp interface{}, opts client.CallOptions) error {
			var tp trace.TracerProvider
			if len(tracerProvider) > 0 && tracerProvider[0] != nil {
				tp = tracerProvider[0]
			}
			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
			ctx, span := StartSpanFromContext(ctx, tp, name)
			defer span.End()
			if err := cf(ctx, node, req, rsp, opts); err != nil {
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
				return err
			}
			return nil
		}
	}
}

// NewHandlerWrapper accepts an opentracing Tracer and returns a Handler Wrapper
func NewHandlerWrapper(tracerProvider ...trace.TracerProvider) server.HandlerWrapper {
	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			var tp trace.TracerProvider
			if len(tracerProvider) > 0 && tracerProvider[0] != nil {
				tp = tracerProvider[0]
			}
			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
			ctx, span := StartSpanFromContext(ctx, tp, name)
			defer span.End()
			if err := h(ctx, req, rsp); err != nil {
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
				return err
			}
			return nil
		}
	}
}

// NewSubscriberWrapper accepts an opentracing Tracer and returns a Subscriber Wrapper
func NewSubscriberWrapper(tracerProvider ...trace.TracerProvider) server.SubscriberWrapper {
	return func(next server.SubscriberFunc) server.SubscriberFunc {
		return func(ctx context.Context, msg server.Message) error {
			var tp trace.TracerProvider
			if len(tracerProvider) > 0 && tracerProvider[0] != nil {
				tp = tracerProvider[0]
			}
			name := "Sub from " + msg.Topic()
			ctx, span := StartSpanFromContext(ctx, tp, name)
			defer span.End()
			if err := next(ctx, msg); err != nil {
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
				return err
			}
			return nil
		}
	}
}
