package server

import (
	"bytes"
	"context"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"testing"
)

func TestRunHTTPReturnsBindErrorWithoutListeningLog(t *testing.T) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	port := listener.Addr().(*net.TCPAddr).Port
	var output bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&output, nil))

	err = RunHTTP(t.Context(), logger, port, testDependencies())
	if err == nil {
		t.Fatal("RunHTTP() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "failed to listen on :"+strconv.Itoa(port)) {
		t.Errorf("RunHTTP() error = %q, want listen failure", err)
	}
	if strings.Contains(output.String(), "server listening") {
		t.Errorf("log output contains listening event: %q", output.String())
	}
}

func TestRunHTTPStopsWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	if err := RunHTTP(ctx, logger, 0, testDependencies()); err != nil {
		t.Fatalf("RunHTTP() error = %v, want nil", err)
	}
}
