package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

const (
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
)

// RunHTTP serves MCP messages over Streamable HTTP until the context is canceled or the HTTP server stops.
func RunHTTP(ctx context.Context, logger *slog.Logger, port int, dependencies Dependencies) error {
	httpServer := newHTTPServer(port, dependencies)
	defer func() { _ = httpServer.Close() }()

	listener, err := net.Listen("tcp", httpServer.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", httpServer.Addr, err)
	}
	defer func() { _ = listener.Close() }()

	logger.Info(
		"server listening",
		slog.String("transport", "http"),
		slog.Int("port", port),
	)

	return serveHTTP(ctx, httpServer, listener)
}

func newHTTPServer(port int, dependencies Dependencies) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           newRouter(dependencies),
		ReadHeaderTimeout: readHeaderTimeout,
	}
}

func serveHTTP(ctx context.Context, httpServer *http.Server, listener net.Listener) error {
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- httpServer.Serve(listener)
	}()

	select {
	case err := <-serveErr:
		return normalizeServeError(err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		_ = httpServer.Close()
		<-serveErr

		return fmt.Errorf("failed to shut down HTTP server: %w", err)
	}

	return normalizeServeError(<-serveErr)
}

func normalizeServeError(err error) error {
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return fmt.Errorf("failed to serve HTTP: %w", err)
}
