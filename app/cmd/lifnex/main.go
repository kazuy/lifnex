package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kazuy/lifnex/app/internal/server"
)

func main() {
	logger, err := newLogger(os.Stderr, loadLoggerOptions())
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "configure logger: %v\n", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	options, err := loadTransportOptions()
	if err != nil {
		logger.Error(
			"server configuration failed",
			slog.String("component", "mcp"),
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	if err := run(ctx, logger, options, server.NewDependencies()); err != nil {
		logger.Error(
			"server stopped",
			slog.String("component", "mcp"),
			slog.String("transport", options.transport),
			slog.Any("error", err),
		)
		os.Exit(1)
	}
}
