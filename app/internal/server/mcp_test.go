package server_test

import (
	"testing"

	"github.com/kazuy/lifnex/app/internal/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPServerListsTools(t *testing.T) {
	t.Parallel()

	expectedTools := []struct {
		name        string
		description string
	}{
		{
			name:        "hello",
			description: "Return a hello-world greeting",
		},
	}

	clientSession := connectClient(t)
	tools, err := clientSession.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}

	if len(tools.Tools) != len(expectedTools) {
		t.Fatalf("tool count = %d, want %d", len(tools.Tools), len(expectedTools))
	}

	toolsByName := make(map[string]*mcp.Tool, len(tools.Tools))
	for _, tool := range tools.Tools {
		toolsByName[tool.Name] = tool
	}

	for _, expected := range expectedTools {
		tool, ok := toolsByName[expected.name]
		if !ok {
			t.Errorf("tool %q is not registered", expected.name)
			continue
		}

		if tool.Description != expected.description {
			t.Errorf("tool %q description = %q, want %q", expected.name, tool.Description, expected.description)
		}
	}
}

func TestHelloTool(t *testing.T) {
	t.Parallel()

	clientSession := connectClient(t)
	result, err := clientSession.CallTool(t.Context(), &mcp.CallToolParams{Name: "hello"})
	if err != nil {
		t.Fatalf("call hello: %v", err)
	}

	if result.IsError {
		t.Fatal("hello returned a tool error")
	}
	if len(result.Content) != 1 {
		t.Fatalf("content length = %d, want 1", len(result.Content))
	}

	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("content type = %T, want *mcp.TextContent", result.Content[0])
	}

	if text.Text != "Hello, world!" {
		t.Fatalf("text = %q, want %q", text.Text, "Hello, world!")
	}
}

func connectClient(t *testing.T) *mcp.ClientSession {
	t.Helper()

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.NewMCP().Connect(t.Context(), serverTransport, nil)
	if err != nil {
		t.Fatalf("connect server: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "lifnex-test",
		Version: "0.1.0",
	}, nil)
	clientSession, err := client.Connect(t.Context(), clientTransport, nil)
	if err != nil {
		t.Fatalf("connect client: %v", err)
	}
	t.Cleanup(func() { _ = clientSession.Close() })

	return clientSession
}
