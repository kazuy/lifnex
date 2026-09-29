package server_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	eventmodel "github.com/kazuy/lifnex/app/internal/model/event"
	"github.com/kazuy/lifnex/app/internal/server"
	eventusecase "github.com/kazuy/lifnex/app/internal/usecase/event"
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
		{
			name:        "search_events",
			description: "Search for Kawasaki City events by date range, title keyword, and location",
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

func TestSearchEventsTool(t *testing.T) {
	t.Parallel()

	provider := &eventProviderStub{
		events:     []eventmodel.Event{{Title: "かわさきジャズ"}},
		totalCount: 1,
	}
	clientSession := connectClientWithProvider(t, provider)

	result, err := clientSession.CallTool(t.Context(), &mcp.CallToolParams{
		Name: "search_events",
		Arguments: map[string]any{
			"from":      "2026-10-01",
			"to":        "2026-10-31",
			"keyword":   "音楽",
			"locations": []string{"幸区"},
		},
	})
	if err != nil {
		t.Fatalf("call search_events: %v", err)
	}
	if result.IsError {
		t.Fatal("search_events returned a tool error")
	}

	wantCondition := eventmodel.SearchCondition{
		From:      time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC),
		To:        time.Date(2026, time.October, 31, 0, 0, 0, 0, time.UTC),
		Keyword:   "音楽",
		Locations: []string{"幸区"},
	}
	if !reflect.DeepEqual(provider.condition, wantCondition) {
		t.Errorf("condition = %#v, want %#v", provider.condition, wantCondition)
	}
}

func TestSearchEventsToolReturnsToolError(t *testing.T) {
	t.Parallel()

	clientSession := connectClientWithProvider(t, &eventProviderStub{err: errors.New("events unavailable")})

	result, err := clientSession.CallTool(t.Context(), &mcp.CallToolParams{
		Name: "search_events",
		Arguments: map[string]any{
			"from": "2026-10-01",
			"to":   "2026-10-31",
		},
	})
	if err != nil {
		t.Fatalf("call search_events: %v", err)
	}
	if !result.IsError {
		t.Fatal("search_events returned success, want tool error")
	}
}

func connectClient(t *testing.T) *mcp.ClientSession {
	t.Helper()

	return connectClientWithProvider(t, &eventProviderStub{})
}

func connectClientWithProvider(t *testing.T, provider *eventProviderStub) *mcp.ClientSession {
	t.Helper()

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	dependencies := server.Dependencies{
		EventSearch: eventusecase.NewSearch(provider),
	}
	serverSession, err := server.NewMCP(dependencies).Connect(t.Context(), serverTransport, nil)
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

type eventProviderStub struct {
	events     []eventmodel.Event
	totalCount int
	condition  eventmodel.SearchCondition
	err        error
}

func (p *eventProviderStub) Search(_ context.Context, condition eventmodel.SearchCondition) ([]eventmodel.Event, int, error) {
	p.condition = condition

	return p.events, p.totalCount, p.err
}
