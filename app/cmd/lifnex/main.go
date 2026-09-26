package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/kazuy/lifnex/app/internal/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	logger, err := newLogger(os.Stderr, loadLoggerOptions())
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "configure logger: %v\n", err)
		os.Exit(1)
	}

	mcpServer := server.NewMCP()
	if err := mcpServer.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		logger.Error("MCP server stopped", slog.Any("error", err))
		os.Exit(1)
	}
}
