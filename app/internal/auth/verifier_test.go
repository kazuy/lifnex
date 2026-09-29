package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kazuy/lifnex/app/internal/auth"
	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
)

const (
	testIssuer   = "https://authorization-server.example"
	testAudience = "https://resource-server.example/mcp"
	testKeyID    = "test-key"
)

func TestVerifierVerify(t *testing.T) {
	privateKey := newPrivateKey(t)
	verifier := newVerifier(t, privateKey)

	tests := []struct {
		name    string
		claims  jwt.MapClaims
		key     *rsa.PrivateKey
		wantErr bool
	}{
		{
			name: "valid",
			claims: validClaims(jwt.MapClaims{
				"scope": "events:read profile",
			}),
			key: privateKey,
		},
		{name: "wrong issuer", claims: validClaims(jwt.MapClaims{"iss": "https://other.example"}), key: privateKey, wantErr: true},
		{name: "wrong audience", claims: validClaims(jwt.MapClaims{"aud": "https://other.example/mcp"}), key: privateKey, wantErr: true},
		{name: "expired", claims: validClaims(jwt.MapClaims{"exp": time.Now().Add(-time.Minute).Unix()}), key: privateKey, wantErr: true},
		{name: "missing expiration", claims: validClaims(jwt.MapClaims{"exp": nil}), key: privateKey, wantErr: true},
		{name: "missing subject", claims: validClaims(jwt.MapClaims{"sub": ""}), key: privateKey, wantErr: true},
		{name: "wrong signature", claims: validClaims(nil), key: newPrivateKey(t), wantErr: true},
		{name: "invalid scope", claims: validClaims(jwt.MapClaims{"scope": 42}), key: privateKey, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rawToken := signToken(t, test.key, test.claims)
			info, err := verifier.Verify(t.Context(), rawToken, nil)
			if test.wantErr {
				if err == nil {
					t.Fatal("Verify() error = nil, want invalid token error")
				}
				if !errors.Is(err, mcpauth.ErrInvalidToken) {
					t.Fatalf("Verify() error = %v, want invalid token error", err)
				}

				return
			}
			if err != nil {
				t.Fatalf("Verify() error = %v", err)
			}
			if info.UserID != "user-123" {
				t.Errorf("UserID = %q, want %q", info.UserID, "user-123")
			}
			if len(info.Scopes) != 2 || info.Scopes[0] != "events:read" || info.Scopes[1] != "profile" {
				t.Errorf("Scopes = %v, want [events:read profile]", info.Scopes)
			}
			if info.Expiration.IsZero() {
				t.Error("Expiration is zero, want token expiration")
			}
		})
	}
}

func TestNewVerifierRequiresConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		config auth.Config
	}{
		{name: "issuer", config: auth.Config{Audience: testAudience, JWKSURL: "https://example.com/jwks"}},
		{name: "audience", config: auth.Config{Issuer: testIssuer, JWKSURL: "https://example.com/jwks"}},
		{name: "JWKS URL", config: auth.Config{Issuer: testIssuer, Audience: testAudience}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := auth.NewVerifier(t.Context(), test.config); err == nil {
				t.Fatal("NewVerifier() error = nil, want configuration error")
			}
		})
	}
}

func newVerifier(t *testing.T, privateKey *rsa.PrivateKey) *auth.Verifier {
	t.Helper()

	jwks, err := json.Marshal(map[string]any{
		"keys": []map[string]any{{
			"alg": "RS256",
			"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes()),
			"kid": testKeyID,
			"kty": "RSA",
			"n":   base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes()),
			"use": "sig",
		}},
	})
	if err != nil {
		t.Fatalf("marshal JWKS: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jwks)
	}))
	t.Cleanup(server.Close)

	verifier, err := auth.NewVerifier(t.Context(), auth.Config{
		Issuer:   testIssuer,
		Audience: testAudience,
		JWKSURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("NewVerifier() error = %v", err)
	}

	return verifier
}

func validClaims(overrides jwt.MapClaims) jwt.MapClaims {
	claims := jwt.MapClaims{
		"aud": testAudience,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iss": testIssuer,
		"sub": "user-123",
	}
	for key, value := range overrides {
		if value == nil {
			delete(claims, key)

			continue
		}
		claims[key] = value
	}

	return claims
}

func signToken(t *testing.T, key *rsa.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = testKeyID
	rawToken, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	return rawToken
}

func newPrivateKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}

	return privateKey
}
