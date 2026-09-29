package server

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
)

const (
	mcpAccessScope                = "mcp:access"
	protectedResourceMetadataPath = "/.well-known/oauth-protected-resource"
)

// OAuthConfig configures the HTTP transport as an OAuth protected resource.
type OAuthConfig struct {
	AuthorizationServerURL string
	ResourceURL            string
	VerifyToken            auth.TokenVerifier
}

func newRouter(dependencies Dependencies, oauth OAuthConfig) (http.Handler, error) {
	metadataURL, err := resourceMetadataURL(oauth.ResourceURL)
	if err != nil {
		return nil, err
	}
	if oauth.AuthorizationServerURL == "" {
		return nil, fmt.Errorf("OAuth authorization server URL is required")
	}
	if oauth.VerifyToken == nil {
		return nil, fmt.Errorf("OAuth token verifier is required")
	}

	mux := http.NewServeMux()
	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return NewMCP(dependencies) },
		&mcp.StreamableHTTPOptions{Stateless: true},
	)
	requireBearerToken := auth.RequireBearerToken(oauth.VerifyToken, &auth.RequireBearerTokenOptions{
		ResourceMetadataURL: metadataURL,
		Scopes:              []string{mcpAccessScope},
	})
	mux.Handle("/mcp", http.NewCrossOriginProtection().Handler(requireBearerToken(mcpHandler)))
	mux.Handle(protectedResourceMetadataPath, auth.ProtectedResourceMetadataHandler(&oauthex.ProtectedResourceMetadata{
		Resource:               oauth.ResourceURL,
		AuthorizationServers:   []string{oauth.AuthorizationServerURL},
		ScopesSupported:        []string{mcpAccessScope},
		BearerMethodsSupported: []string{"header"},
		ResourceName:           "lifnex",
	}))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})

	return mux, nil
}

func resourceMetadataURL(resourceURL string) (string, error) {
	parsed, err := url.Parse(resourceURL)
	if err != nil {
		return "", fmt.Errorf("parse OAuth resource URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("parse OAuth resource URL %q: absolute URL required", resourceURL)
	}

	return (&url.URL{
		Scheme: parsed.Scheme,
		Host:   parsed.Host,
		Path:   protectedResourceMetadataPath,
	}).String(), nil
}
