package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	eventmodel "github.com/kazuy/lifnex/app/internal/model/event"
	eventusecase "github.com/kazuy/lifnex/app/internal/usecase/event"
	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
)

const (
	testAuthorizationServerURL = "https://authorization-server.example"
	testResourceURL            = "https://resource-server.example/mcp"
)

func TestHTTPHealthEndpoint(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()
	newTestRouter(t).ServeHTTP(response, request)

	result := response.Result()
	defer result.Body.Close()
	if result.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", result.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if got := string(body); got != "ok\n" {
		t.Errorf("body = %q, want %q", got, "ok\n")
	}
}

func TestHTTPProtectedResourceMetadata(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodGet, protectedResourceMetadataPath, nil)
	response := httptest.NewRecorder()
	newTestRouter(t).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	var metadata oauthex.ProtectedResourceMetadata
	if err := json.NewDecoder(response.Body).Decode(&metadata); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}
	if metadata.Resource != testResourceURL {
		t.Errorf("resource = %q, want %q", metadata.Resource, testResourceURL)
	}
	if len(metadata.AuthorizationServers) != 1 || metadata.AuthorizationServers[0] != testAuthorizationServerURL {
		t.Errorf("authorization servers = %v, want [%s]", metadata.AuthorizationServers, testAuthorizationServerURL)
	}
	if len(metadata.BearerMethodsSupported) != 1 || metadata.BearerMethodsSupported[0] != "header" {
		t.Errorf("bearer methods = %v, want [header]", metadata.BearerMethodsSupported)
	}
}

func TestHTTPMCPRequiresBearerToken(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	response := httptest.NewRecorder()
	newTestRouter(t).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	wantChallenge := `Bearer resource_metadata="https://resource-server.example/.well-known/oauth-protected-resource"`
	if got := response.Header().Get("WWW-Authenticate"); got != wantChallenge {
		t.Errorf("WWW-Authenticate = %q, want %q", got, wantChallenge)
	}
}

func TestHTTPCrossOriginRequestIsRejected(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	request.Header.Set("Origin", "https://attacker.example")
	response := httptest.NewRecorder()
	newTestRouter(t).ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
}

func TestHTTPHelloTool(t *testing.T) {
	t.Parallel()

	httpServer := httptest.NewServer(newTestRouter(t))
	t.Cleanup(httpServer.Close)

	client := mcp.NewClient(&mcp.Implementation{Name: "lifnex-http-test", Version: "0.1.0"}, nil)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{
		Endpoint: httpServer.URL + "/mcp",
		HTTPClient: &http.Client{Transport: roundTripperFunc(func(request *http.Request) (*http.Response, error) {
			request.Header.Set("Authorization", "Bearer valid-token")

			return http.DefaultTransport.RoundTrip(request)
		})},
	}, nil)
	if err != nil {
		t.Fatalf("connect client: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "hello"})
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
		t.Errorf("text = %q, want %q", text.Text, "Hello, world!")
	}
}

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()

	router, err := newRouter(testDependencies(), testOAuthConfig())
	if err != nil {
		t.Fatalf("newRouter() error = %v", err)
	}

	return router
}

func testOAuthConfig() OAuthConfig {
	return OAuthConfig{
		AuthorizationServerURL: testAuthorizationServerURL,
		ResourceURL:            testResourceURL,
		VerifyToken: func(_ context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
			if token != "valid-token" {
				return nil, auth.ErrInvalidToken
			}

			return &auth.TokenInfo{
				Expiration: time.Now().Add(time.Hour),
				UserID:     "user-123",
			}, nil
		},
	}
}

func testDependencies() Dependencies {
	return Dependencies{
		EventSearch: eventusecase.NewSearch(testEventProvider{}),
	}
}

type testEventProvider struct{}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (function roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func (testEventProvider) Search(context.Context, eventmodel.SearchCondition) ([]eventmodel.Event, int, error) {
	return nil, 0, nil
}
