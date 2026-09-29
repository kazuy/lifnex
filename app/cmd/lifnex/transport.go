package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	oauth "github.com/kazuy/lifnex/app/internal/auth"
	"github.com/kazuy/lifnex/app/internal/server"
)

const (
	stdioTransport = "stdio"
	httpTransport  = "http"
	defaultPort    = 8080
)

type transportOptions struct {
	transport string
	port      int
	oauth     oauth.Config
}

func loadTransportOptions() (transportOptions, error) {
	transport := strings.ToLower(strings.TrimSpace(os.Getenv("TRANSPORT")))
	if transport == "" {
		transport = stdioTransport
	}
	if transport != stdioTransport && transport != httpTransport {
		return transportOptions{}, fmt.Errorf("failed to parse TRANSPORT %q: unsupported value", transport)
	}

	port := defaultPort
	if value := strings.TrimSpace(os.Getenv("PORT")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return transportOptions{}, fmt.Errorf("failed to parse PORT %q: %w", value, err)
		}
		if parsed < 1 || parsed > 65535 {
			return transportOptions{}, fmt.Errorf("failed to validate PORT %q: must be between 1 and 65535", value)
		}
		port = parsed
	}

	options := transportOptions{transport: transport, port: port}
	if transport == httpTransport {
		options.oauth = oauth.Config{
			Issuer:   strings.TrimSpace(os.Getenv("OAUTH_ISSUER_URL")),
			Audience: strings.TrimSpace(os.Getenv("OAUTH_RESOURCE_URL")),
			JWKSURL:  strings.TrimSpace(os.Getenv("OAUTH_JWKS_URL")),
		}
		if options.oauth.Issuer == "" {
			return transportOptions{}, fmt.Errorf("failed to validate OAUTH_ISSUER_URL: required for HTTP transport")
		}
		if options.oauth.Audience == "" {
			return transportOptions{}, fmt.Errorf("failed to validate OAUTH_RESOURCE_URL: required for HTTP transport")
		}
		if options.oauth.JWKSURL == "" {
			return transportOptions{}, fmt.Errorf("failed to validate OAUTH_JWKS_URL: required for HTTP transport")
		}
	}

	return options, nil
}

func run(ctx context.Context, logger *slog.Logger, options transportOptions, dependencies server.Dependencies) error {
	switch options.transport {
	case stdioTransport:
		return server.RunStdio(ctx, dependencies)
	case httpTransport:
		verifier, err := oauth.NewVerifier(ctx, options.oauth)
		if err != nil {
			return fmt.Errorf("failed to configure OAuth: %w", err)
		}

		return server.RunHTTP(ctx, logger, options.port, dependencies, server.OAuthConfig{
			AuthorizationServerURL: options.oauth.Issuer,
			ResourceURL:            options.oauth.Audience,
			VerifyToken:            verifier.Verify,
		})
	default:
		return fmt.Errorf("failed to run transport %q: unsupported value", options.transport)
	}
}
