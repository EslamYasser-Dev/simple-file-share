package grpcapi

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeRecorder struct {
	method string
	code   string
}

func (f *fakeRecorder) ObserveHTTP(string, string, int, time.Duration, int64) {}
func (f *fakeRecorder) ObserveRPC(method, code string, _ time.Duration) {
	f.method = method
	f.code = code
}
func (f *fakeRecorder) AddActiveStreams(int) {}

func TestUnaryMetricsInterceptorRecords(t *testing.T) {
	rec := &fakeRecorder{}
	itc := unaryMetricsInterceptor(rec)

	resp, err := itc(
		context.Background(),
		"req",
		&grpc.UnaryServerInfo{FullMethod: "/fileshare.v1.AuthService/Login"},
		func(context.Context, any) (any, error) { return "token", nil },
	)
	if err != nil || resp != "token" {
		t.Fatalf("handler passthrough broken: resp=%v err=%v", resp, err)
	}
	if rec.method != "/fileshare.v1.AuthService/Login" || rec.code != "OK" {
		t.Fatalf("recorded %s/%s", rec.method, rec.code)
	}
}

func TestUnaryMetricsInterceptorRecordsErrors(t *testing.T) {
	rec := &fakeRecorder{}
	itc := unaryMetricsInterceptor(rec)

	_, err := itc(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/x/Y"},
		func(context.Context, any) (any, error) {
			return nil, status.Error(codes.PermissionDenied, "no")
		},
	)
	if err == nil {
		t.Fatal("error must propagate")
	}
	if rec.code != "PermissionDenied" {
		t.Fatalf("code = %q, want PermissionDenied", rec.code)
	}
}

func TestStreamMetricsInterceptorRecords(t *testing.T) {
	rec := &fakeRecorder{}
	itc := streamMetricsInterceptor(rec)

	err := itc(
		nil,
		nil,
		&grpc.StreamServerInfo{FullMethod: "/fileshare.v1.EventsService/Subscribe"},
		func(any, grpc.ServerStream) error { return nil },
	)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.method != "/fileshare.v1.EventsService/Subscribe" || rec.code != "OK" {
		t.Fatalf("recorded %s/%s", rec.method, rec.code)
	}
}
