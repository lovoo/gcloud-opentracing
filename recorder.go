package gcloudtracer

import (
	"fmt"
	"log"
	"strconv"

	trace "cloud.google.com/go/trace/apiv1"
	"github.com/golang/protobuf/ptypes/timestamp"
	basictracer "github.com/opentracing/basictracer-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"golang.org/x/net/context"
	pb "google.golang.org/genproto/googleapis/devtools/cloudtrace/v1"
)

var _ basictracer.SpanRecorder = &Recorder{}

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
	if options.log == nil {
		options.log = &defaultLogger{}
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
	traceID := fmt.Sprintf("%016x%016x", sp.Context.TraceID, sp.Context.TraceID)
	nanos := sp.Start.UnixNano()

	req := &pb.PatchTracesRequest{
		ProjectId: r.project,
		Traces: &pb.Traces{
			Traces: []*pb.Trace{
				{
					ProjectId: r.project,
					TraceId:   traceID,
					Spans: []*pb.TraceSpan{
						{
							SpanId: sp.Context.SpanID,
							Kind:   convertSpanKind(sp.Tags),
							Name:   sp.Operation,
							StartTime: &timestamp.Timestamp{
								Seconds: nanos / 1e9,
								Nanos:   int32(nanos % 1e9),
							},
							EndTime: &timestamp.Timestamp{
								Seconds: (nanos + int64(sp.Duration)) / 1e9,
								Nanos:   int32((nanos + int64(sp.Duration)) % 1e9),
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

func convertSpanKind(tags opentracing.Tags) pb.TraceSpan_SpanKind {
	switch tags[string(ext.SpanKind)] {
	case ext.SpanKindRPCServerEnum:
		return pb.TraceSpan_RPC_SERVER
	case ext.SpanKindRPCClientEnum:
		return pb.TraceSpan_RPC_CLIENT
	default:
		return pb.TraceSpan_SPAN_KIND_UNSPECIFIED
	}
}

type defaultLogger struct{}

func (defaultLogger) Errorf(msg string, args ...interface{}) {
	log.Printf(msg, args...)
}
