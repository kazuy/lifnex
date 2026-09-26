package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RunStdio serves MCP messages over standard input and output until the context is canceled or the transport stops.
func RunStdio(ctx context.Context) error {
	return NewMCP().Run(ctx, &mcp.StdioTransport{})
}
