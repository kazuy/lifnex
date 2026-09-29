package main

import (
	"testing"

	oauth "github.com/kazuy/lifnex/app/internal/auth"
)

func TestLoadTransportOptions(t *testing.T) {
	tests := []struct {
		name      string
		transport string
		port      string
		issuer    string
		resource  string
		jwks      string
		want      transportOptions
		wantErr   string
	}{
		{name: "defaults", want: transportOptions{transport: stdioTransport, port: defaultPort}},
		{
			name:      "http",
			transport: "HTTP",
			port:      "9090",
			issuer:    "https://authorization-server.example",
			resource:  "https://resource-server.example/mcp",
			jwks:      "https://authorization-server.example/jwks",
			want: transportOptions{
				transport: httpTransport,
				port:      9090,
				oauth: oauth.Config{
					Issuer:   "https://authorization-server.example",
					Audience: "https://resource-server.example/mcp",
					JWKSURL:  "https://authorization-server.example/jwks",
				},
			},
		},
		{name: "invalid transport", transport: "sse", wantErr: `failed to parse TRANSPORT "sse": unsupported value`},
		{name: "invalid port text", transport: "http", port: "abc", wantErr: `failed to parse PORT "abc": strconv.Atoi: parsing "abc": invalid syntax`},
		{name: "invalid port range", transport: "http", port: "65536", wantErr: `failed to validate PORT "65536": must be between 1 and 65535`},
		{name: "missing issuer", transport: "http", resource: "https://resource-server.example/mcp", jwks: "https://authorization-server.example/jwks", wantErr: "failed to validate OAUTH_ISSUER_URL: required for HTTP transport"},
		{name: "missing resource", transport: "http", issuer: "https://authorization-server.example", jwks: "https://authorization-server.example/jwks", wantErr: "failed to validate OAUTH_RESOURCE_URL: required for HTTP transport"},
		{name: "missing JWKS URL", transport: "http", issuer: "https://authorization-server.example", resource: "https://resource-server.example/mcp", wantErr: "failed to validate OAUTH_JWKS_URL: required for HTTP transport"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("TRANSPORT", test.transport)
			t.Setenv("PORT", test.port)
			t.Setenv("OAUTH_ISSUER_URL", test.issuer)
			t.Setenv("OAUTH_RESOURCE_URL", test.resource)
			t.Setenv("OAUTH_JWKS_URL", test.jwks)

			got, err := loadTransportOptions()
			if test.wantErr != "" {
				if err == nil {
					t.Fatalf("loadTransportOptions() error = nil, want %q", test.wantErr)
				}
				if err.Error() != test.wantErr {
					t.Errorf("loadTransportOptions() error = %q, want %q", err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("loadTransportOptions() error = %v, want nil", err)
			}
			if got != test.want {
				t.Errorf("loadTransportOptions() = %+v, want %+v", got, test.want)
			}
		})
	}
}
