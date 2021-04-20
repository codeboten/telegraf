package opentelemetry

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	metricsService "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestClientWithRecoverableError(t *testing.T) {
	listener := newLocalListener()
	grpcServer := grpc.NewServer()
	mockMetricsServer := metricServiceServer{
		status: nil,
	}
	metricsService.RegisterMetricsServiceServer(grpcServer, &mockMetricsServer)
	go func() {
		_ = grpcServer.Serve(listener)
	}()
	defer grpcServer.Stop()

	u, err := url.Parse("https://" + listener.Addr().String())
	require.NoError(t, err)

	client := client{
		logger:  testutil.Logger{},
		url:     u,
		timeout: time.Second,
	}
	_, err = client.connect(context.Background())
	require.True(t, isRecoverable(err), "expected recoverableError in error %v", err)
}

func TestClientWithUnrecoverableError(t *testing.T) {
	thing := newLocalListener()
	grpcServer := grpc.NewServer()
	mockMetricsServer := metricServiceServer{
		status: status.New(codes.InvalidArgument, "the request was missing some important arguments, change the arguments before retrying the request"),
	}
	metricsService.RegisterMetricsServiceServer(grpcServer, &mockMetricsServer)
	go func() {
		_ = grpcServer.Serve(thing)
	}()
	defer grpcServer.Stop()

	u, err := url.Parse("http://" + thing.Addr().String())
	require.NoError(t, err)

	client := client{
		logger:  testutil.Logger{},
		url:     u,
		timeout: time.Second,
	}

	err = client.ping(context.Background())
	require.False(t, isRecoverable(err), "expected unrecoverableError in error %v", err)

	_, err = client.connect(context.Background())
	require.False(t, isRecoverable(err), "expected unrecoverableError in error %v", err)
}

func TestEmptyRequest(t *testing.T) {
	serverURL, err := url.Parse("http://localhost:12345")
	require.NoError(t, err)

	c := client{
		logger:  testutil.Logger{},
		url:     serverURL,
		timeout: time.Second,
	}

	err = c.store(&metricsService.ExportMetricsServiceRequest{})
	require.NoError(t, err)
}
