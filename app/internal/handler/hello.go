package handler

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Hello returns a static greeting.
func Hello(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Hello, world!"},
		},
	}, nil, nil
}
