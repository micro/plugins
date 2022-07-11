package opentelemetry

import (
	"context"
	"fmt"

	"go-micro.dev/v4/client"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type clientWrapper struct {
	client.Client

	tp trace.TracerProvider
}

func (w *clientWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
	ctx, span := StartSpanFromContext(ctx, w.tp, name)
	defer span.End()
	if err := w.Client.Call(ctx, req, rsp, opts...); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}
	return nil
}

func (w *clientWrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
	ctx, span := StartSpanFromContext(ctx, w.tp, name)
	defer span.End()
	stream, err := w.Client.Stream(ctx, req, opts...)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
	}
	return stream, err
}

func (w *clientWrapper) Publish(ctx context.Context, p client.Message, opts ...client.PublishOption) error {
	name := fmt.Sprintf("Pub to %s", p.Topic())
	ctx, span := StartSpanFromContext(ctx, w.tp, name)
	defer span.End()
	if err := w.Client.Publish(ctx, p, opts...); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}
	return nil
}
