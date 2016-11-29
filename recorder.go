package gcloudtracer

import (
	"log"
	"strconv"

	trace "cloud.google.com/go/trace/apiv1"
	"github.com/golang/protobuf/ptypes/timestamp"
	basictracer "github.com/opentracing/basictracer-go"
	opentracing "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	cloudtracepb "google.golang.org/genproto/googleapis/devtools/cloudtrace/v1"
)

// Logger defines an interface to log an error.
type Logger interface {
	Errorf(string, ...interface{})
}

// Recorder implements basictracer.SpanRecorder interface
// used to write traces to the GCE StackDriver.
type Recorder struct {
	project     string
	ctx         context.Context
	log         Logger
	traceClient *trace.Client
}

// NewRecorder creates new GCloud StackDriver recorder.
func NewRecorder(ctx context.Context, opts ...Option) (*Recorder, error) {
	var options Options
	for _, o := range opts {
		o(&options)
	}
	if err := options.Valid(); err != nil {
		return nil, err
	}
	c, err := trace.NewClient(ctx, options.external...)
	if err != nil {
		return nil, err
	}
	return &Recorder{
		project:     options.projectID,
		ctx:         ctx,
		traceClient: c,
		log:         options.log,
	}, nil
}

// RecordSpan writes Span to the GCLoud StackDriver.
func (r *Recorder) RecordSpan(sp basictracer.RawSpan) {
	req := &cloudtracepb.PatchTracesRequest{
		ProjectId: r.project,
		Traces: &cloudtracepb.Traces{
			Traces: []*cloudtracepb.Trace{
				&cloudtracepb.Trace{
					ProjectId: r.project,
					TraceId:   strconv.FormatUint(sp.Context.TraceID, 10),
					Spans: []*cloudtracepb.TraceSpan{
						&cloudtracepb.TraceSpan{
							SpanId: sp.Context.SpanID,
							Kind:   cloudtracepb.TraceSpan_SPAN_KIND_UNSPECIFIED,
							Name:   sp.Operation,
							StartTime: &timestamp.Timestamp{
								Seconds: sp.Start.Unix(),
							},
							EndTime: &timestamp.Timestamp{
								Seconds: sp.Start.Add(sp.Duration).Unix(),
							},
							ParentSpanId: sp.ParentSpanID,
							Labels:       convertTags(sp.Tags),
						},
					},
				},
			},
		},
	}

	if err := r.traceClient.PatchTraces(r.ctx, req); err != nil {
		r.log.Errorf("failed to write trace: %v", err)
	}
}

type defaultLogger struct{}

func (defaultLogger) Errorf(msg string, args ...interface{}) {
	log.Printf(msg, args...)
}

func convertTags(tags opentracing.Tags) map[string]string {
	labels := make(map[string]string)
	for k, v := range tags {
		switch v := v.(type) {
		case int:
			labels[k] = strconv.Itoa(v)
		case string:
			labels[k] = v
		}
	}
	return labels
}
