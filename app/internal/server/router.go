package server

import (
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func newRouter() http.Handler {
	mux := http.NewServeMux()
	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return NewMCP() },
		&mcp.StreamableHTTPOptions{Stateless: true},
	)
	mux.Handle("/mcp", http.NewCrossOriginProtection().Handler(mcpHandler))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})

	return mux
}
