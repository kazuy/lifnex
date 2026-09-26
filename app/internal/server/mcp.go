package server

import (
	"github.com/kazuy/lifnex/app/internal/handler"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverName    = "lifnex"
	serverVersion = "0.1.0"
)

// NewMCP creates a Lifnex MCP server with all supported tools registered.
func NewMCP() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "hello",
		Description: "Return a hello-world greeting",
	}, handler.Hello)

	return server
}
