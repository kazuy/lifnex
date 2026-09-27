package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/kazuy/lifnex/app/internal/server"
)

const (
	stdioTransport = "stdio"
	httpTransport  = "http"
	defaultPort    = 8080
)

type transportOptions struct {
	transport string
	port      int
}

func loadTransportOptions() (transportOptions, error) {
	transport := strings.ToLower(strings.TrimSpace(os.Getenv("TRANSPORT")))
	if transport == "" {
		transport = stdioTransport
	}
	if transport != stdioTransport && transport != httpTransport {
		return transportOptions{}, fmt.Errorf("failed to parse TRANSPORT %q: unsupported value", transport)
	}

	port := defaultPort
	if value := strings.TrimSpace(os.Getenv("PORT")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return transportOptions{}, fmt.Errorf("failed to parse PORT %q: %w", value, err)
		}
		if parsed < 1 || parsed > 65535 {
			return transportOptions{}, fmt.Errorf("failed to validate PORT %q: must be between 1 and 65535", value)
		}
		port = parsed
	}

	return transportOptions{transport: transport, port: port}, nil
}

func run(ctx context.Context, logger *slog.Logger, options transportOptions) error {
	switch options.transport {
	case stdioTransport:
		return server.RunStdio(ctx)
	case httpTransport:
		return server.RunHTTP(ctx, logger, options.port)
	default:
		return fmt.Errorf("failed to run transport %q: unsupported value", options.transport)
	}
}
