package gcloudtracer

import (
	"log"
	"net"
	"os"
	"testing"
	"time"

	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	basictracer "github.com/opentracing/basictracer-go"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	cloudtracepb "google.golang.org/genproto/googleapis/devtools/cloudtrace/v1"
	"google.golang.org/grpc"
)

var (
	clientOpt option.ClientOption
	mockTrace *mockTraceServer
)

func TestMain(m *testing.M) {
	mockTrace = &mockTraceServer{
		patchTraces: func(ctx context.Context, req *cloudtracepb.PatchTracesRequest) (*google_protobuf.Empty, error) {
			return &google_protobuf.Empty{}, nil
		},
	}
	serv := grpc.NewServer()
	cloudtracepb.RegisterTraceServiceServer(serv, mockTrace)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatal(err)

	}
	go serv.Serve(lis)

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	clientOpt = option.WithGRPCConn(conn)
	os.Exit(m.Run())
}

func TestRecorder(t *testing.T) {
	t.Run("recorder=success", func(t *testing.T) {
		called := false
		recorder, err := NewRecorder(
			context.Background(),
			"test_project",
			WithLogger(testLogger(func(format string, args ...interface{}) {
				called = true
			})),
			WithClientOption(clientOpt),
		)
		assert.NoError(t, err)
		// BufferedByteLimit is set to 1 for test propose, to send the trace immediately
		recorder.bundler.BufferedByteLimit = 1

		recorder.RecordSpan(basictracer.RawSpan{
			Context: basictracer.SpanContext{
				Sampled: true,
				SpanID:  10,
				TraceID: 1,
			},
			Operation:    "test/operation",
			ParentSpanID: 9,
			Start:        time.Unix(1480425868, 0),
			Duration:     5 * time.Second,
			Tags: map[string]interface{}{
				"foo": 10,
				"bar": "foo",
			},
		})
		assert.False(t, called, "logger should not be called")

		assert.Equal(t, &cloudtracepb.PatchTracesRequest{
			ProjectId: "test_project",
			Traces: &cloudtracepb.Traces{
				Traces: []*cloudtracepb.Trace{
					&cloudtracepb.Trace{
						ProjectId: "test_project",
						TraceId:   "00000000000000010000000000000001",
						Spans: []*cloudtracepb.TraceSpan{
							&cloudtracepb.TraceSpan{
								SpanId: 10,
								Kind:   cloudtracepb.TraceSpan_SPAN_KIND_UNSPECIFIED,
								Name:   "test/operation",
								StartTime: &timestamp.Timestamp{
									Seconds: 1480425868,
								},
								EndTime: &timestamp.Timestamp{
									Seconds: 1480425873,
								},
								ParentSpanId: 9,
								Labels: map[string]string{
									"foo": "10",
									"bar": "foo",
								},
							},
						},
					},
				},
			},
		}, mockTrace.patchTracesRequest)
	})
}

func TestRecorderMissingProjectID(t *testing.T) {
	r, err := NewRecorder(context.Background(), "")
	assert.Equal(t, ErrInvalidProjectID, err)
	assert.Nil(t, r)
}

type mockTraceServer struct {
	listTracesRequest  *cloudtracepb.ListTracesRequest
	getTraceRequest    *cloudtracepb.GetTraceRequest
	patchTracesRequest *cloudtracepb.PatchTracesRequest

	listTraces  func(context.Context, *cloudtracepb.ListTracesRequest) (*cloudtracepb.ListTracesResponse, error)
	getTrace    func(context.Context, *cloudtracepb.GetTraceRequest) (*cloudtracepb.Trace, error)
	patchTraces func(context.Context, *cloudtracepb.PatchTracesRequest) (*google_protobuf.Empty, error)
}

func (s *mockTraceServer) ListTraces(ctx context.Context, req *cloudtracepb.ListTracesRequest) (*cloudtracepb.ListTracesResponse, error) {
	s.listTracesRequest = req
	return s.listTraces(ctx, req)
}

func (s *mockTraceServer) GetTrace(ctx context.Context, req *cloudtracepb.GetTraceRequest) (*cloudtracepb.Trace, error) {
	s.getTraceRequest = req
	return s.getTrace(ctx, req)

}

func (s *mockTraceServer) PatchTraces(ctx context.Context, req *cloudtracepb.PatchTracesRequest) (*google_protobuf.Empty, error) {
	s.patchTracesRequest = req
	return s.patchTraces(ctx, req)
}

type testLogger func(format string, args ...interface{})

func (l testLogger) Errorf(format string, args ...interface{}) {
	l(format, args...)
}

func (l testLogger) Infof(format string, args ...interface{}) {
}
