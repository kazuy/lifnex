package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
)

var allowedSigningMethods = []string{jwt.SigningMethodRS256.Alg()}

// Config identifies the authorization server and protected resource expected in access tokens.
type Config struct {
	Issuer   string
	Audience string
	JWKSURL  string
}

// Verifier validates JWT access tokens with keys published by an OAuth authorization server.
type Verifier struct {
	keyfunc keyfunc.Keyfunc
	parser  *jwt.Parser
}

type accessTokenClaims struct {
	jwt.RegisteredClaims
	Scope string `json:"scope,omitempty"`
}

// NewVerifier creates a verifier and starts refreshing keys from the configured JWKS URL.
func NewVerifier(ctx context.Context, config Config) (*Verifier, error) {
	if strings.TrimSpace(config.Issuer) == "" {
		return nil, fmt.Errorf("issuer is required")
	}
	if strings.TrimSpace(config.Audience) == "" {
		return nil, fmt.Errorf("audience is required")
	}
	if strings.TrimSpace(config.JWKSURL) == "" {
		return nil, fmt.Errorf("JWKS URL is required")
	}

	keys, err := keyfunc.NewDefaultCtx(ctx, []string{config.JWKSURL})
	if err != nil {
		return nil, fmt.Errorf("load JWKS: %w", err)
	}

	parser := jwt.NewParser(
		jwt.WithValidMethods(allowedSigningMethods),
		jwt.WithIssuer(config.Issuer),
		jwt.WithAudience(config.Audience),
		jwt.WithExpirationRequired(),
	)

	return &Verifier{keyfunc: keys, parser: parser}, nil
}

// Verify validates a bearer token and returns the claims used by the MCP server.
func (v *Verifier) Verify(ctx context.Context, rawToken string, _ *http.Request) (*mcpauth.TokenInfo, error) {
	claims := &accessTokenClaims{}
	token, err := v.parser.ParseWithClaims(rawToken, claims, v.keyfunc.KeyfuncCtx(ctx))
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("%w: verify JWT", mcpauth.ErrInvalidToken)
	}

	if claims.ExpiresAt == nil {
		return nil, fmt.Errorf("%w: read expiration", mcpauth.ErrInvalidToken)
	}
	if strings.TrimSpace(claims.Subject) == "" {
		return nil, fmt.Errorf("%w: read subject", mcpauth.ErrInvalidToken)
	}

	return &mcpauth.TokenInfo{
		Scopes:     strings.Fields(claims.Scope),
		Expiration: claims.ExpiresAt.Time,
		UserID:     claims.Subject,
	}, nil
}
