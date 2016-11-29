package gcloudtracer

import (
	basictracer "github.com/opentracing/basictracer-go"
	opentracing "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
)

// NewTracer creates new basictracer for GCloud StackDriver.
func NewTracer(ctx context.Context, opts ...Option) (opentracing.Tracer, error) {
	recorder, err := NewRecorder(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return basictracer.New(recorder), nil
}
