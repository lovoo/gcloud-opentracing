package gcloudtracer

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"time"

	trace "cloud.google.com/go/trace/apiv1"
	"github.com/golang/protobuf/ptypes/timestamp"
	basictracer "github.com/opentracing/basictracer-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"golang.org/x/net/context"
	"google.golang.org/api/support/bundler"
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
	bundler     *bundler.Bundler
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

	rec := &Recorder{
		project:     options.projectID,
		ctx:         ctx,
		traceClient: c,
		log:         options.log,
	}

	bundler := bundler.NewBundler((*pb.Trace)(nil), func(bundle interface{}) {
		traces := bundle.([]*pb.Trace)
		err := rec.upload(traces)
		if err != nil {
			rec.log.Errorf("failed to upload %d traces to the Cloud Trace server. (err = %s)", len(traces), err)
		}
	})
	bundler.DelayThreshold = 2 * time.Second
	bundler.BundleCountThreshold = 100
	// We're not measuring bytes here, we're counting traces and spans as one "byte" each.
	bundler.BundleByteThreshold = 1000
	bundler.BundleByteLimit = 1000
	bundler.BufferedByteLimit = 10000
	rec.bundler = bundler

	return rec, nil
}

// RecordSpan writes Span to the GCLoud StackDriver.
func (r *Recorder) RecordSpan(sp basictracer.RawSpan) {
	traceID := fmt.Sprintf("%016x%016x", sp.Context.TraceID, sp.Context.TraceID)
	nanos := sp.Start.UnixNano()
	labels := convertTags(sp.Tags)
	transposeLabels(labels)
	addLogs(labels, sp.Logs)

	trace := &pb.Trace{
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
				Labels:       labels,
			},
		},
	}
	go func() {
		err := r.bundler.Add(trace, 2) // size = (1 trace + 1 span)
		if err == bundler.ErrOverflow {
			r.log.Errorf("trace upload bundle too full. uploading immediately")
			err = r.upload([]*pb.Trace{trace})
			if err != nil {
				r.log.Errorf("error uploading trace: %s", err)
			}
		}
	}()
}

func (r *Recorder) upload(traces []*pb.Trace) error {
	req := &pb.PatchTracesRequest{
		ProjectId: r.project,
		Traces:    &pb.Traces{Traces: traces},
	}
	return r.traceClient.PatchTraces(r.ctx, req)
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

var labelMap = map[string]string{
	string(ext.PeerHostname):   `trace.cloud.google.com/http/host`,
	string(ext.HTTPMethod):     `trace.cloud.google.com/http/method`,
	string(ext.HTTPStatusCode): `trace.cloud.google.com/http/status_code`,
	string(ext.HTTPUrl):        `trace.cloud.google.com/http/url`,
}

// rewrite well-known opentracing.ext labels into those gcloud-native labels
func transposeLabels(labels map[string]string) {
	for k, t := range labelMap {
		if vv, ok := labels[k]; ok {
			labels[t] = vv
			delete(labels, k)
		}
	}
}

// copy opentracing events into gcloud trace labels
func addLogs(target map[string]string, logs []opentracing.LogRecord) {
	for i, l := range logs {
		buf := bytes.NewBufferString(l.Timestamp.String())
		for j, f := range l.Fields {
			buf.WriteString(f.Key())
			buf.WriteString("=")
			buf.WriteString(fmt.Sprint(f.Value()))
			if j != len(l.Fields)+1 {
				buf.WriteString(" ")
			}
		}
		target[fmt.Sprintf("event_%d", i)] = buf.String()
	}
}

type defaultLogger struct{}

func (defaultLogger) Errorf(msg string, args ...interface{}) {
	log.Printf(msg, args...)
}
