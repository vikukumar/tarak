// Package security — bearer token generation and validation.
//
// Tarak uses HMAC-SHA256 signed tokens for bootstrap admin access and service
// account tokens in Phase 1.  This is a simple, auditable implementation using
// only stdlib crypto — no external JWT library required.
//
// Token format:
//
//	base64url(header) . base64url(payload) . base64url(HMAC-SHA256(header.payload, secret))
//
// Header:  {"alg":"HS256","typ":"JWT"}
// Payload: {"sub":"<username>","groups":["<group>",...],"iss":"tarak","iat":<unix>,"exp":<unix>}
//
// Tokens are NOT encrypted; they are signed.  Sensitive values are never stored
// in the token body.
package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ─── TokenClaims ──────────────────────────────────────────────────────────────

// TokenClaims contains the authenticated identity extracted from a valid token.
type TokenClaims struct {
	// Subject is the username.
	Subject string
	// Groups is the list of groups the user belongs to.
	Groups []string
	// IssuedAt is when the token was issued.
	IssuedAt time.Time
	// ExpiresAt is when the token expires.
	ExpiresAt time.Time
}

// Expired returns true if the token has expired.
func (c *TokenClaims) Expired() bool {
	return time.Now().After(c.ExpiresAt)
}

// ─── TokenSigner ──────────────────────────────────────────────────────────────

// TokenSigner signs and verifies bearer tokens.
type TokenSigner struct {
	secret []byte
}

// NewTokenSigner creates a TokenSigner with the given HMAC secret.
// The secret should be at least 32 bytes of random data.
func NewTokenSigner(secret []byte) *TokenSigner {
	return &TokenSigner{secret: secret}
}

// GenerateSecret creates a new random 64-byte signing secret.
func GenerateSecret() ([]byte, error) {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("generate token secret: %w", err)
	}
	return b, nil
}

// StaticToken is a pre-issued token with a fixed value, used for bootstrapping.
// These tokens bypass the HMAC verification path and are validated by exact string match.
// They should be replaced with properly signed tokens after initial cluster setup.
type StaticToken struct {
	// Token is the raw token value (the bearer token string).
	Token string
	// User is the username this token represents.
	User string
	// Groups is the list of groups this token grants.
	Groups []string
}

// jwtHeader is the fixed base64url-encoded JWT header.
// {"alg":"HS256","typ":"JWT"}
var jwtHeader = base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

type jwtPayload struct {
	Sub    string   `json:"sub"`
	Groups []string `json:"groups,omitempty"`
	Iss    string   `json:"iss"`
	Iat    int64    `json:"iat"`
	Exp    int64    `json:"exp"`
}

// Issue creates a new signed token for the given user and groups.
func (s *TokenSigner) Issue(subject string, groups []string, ttl time.Duration) (string, error) {
	now := time.Now()
	payload := jwtPayload{
		Sub:    subject,
		Groups: groups,
		Iss:    "tarak",
		Iat:    now.Unix(),
		Exp:    now.Add(ttl).Unix(),
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal token payload: %w", err)
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	sig := s.sign(jwtHeader + "." + payloadB64)
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)

	return jwtHeader + "." + payloadB64 + "." + sigB64, nil
}

// Verify validates a token string and returns its claims.
// Returns an error if the token is invalid, expired, or tampered with.
func (s *TokenSigner) Verify(token string) (*TokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Verify signature.
	signingInput := parts[0] + "." + parts[1]
	expectedSig := s.sign(signingInput)
	providedSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("decode token signature: %w", err)
	}
	if !hmac.Equal(expectedSig, providedSig) {
		return nil, fmt.Errorf("invalid token signature")
	}

	// Decode and parse the payload.
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode token payload: %w", err)
	}
	var payload jwtPayload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return nil, fmt.Errorf("parse token payload: %w", err)
	}

	// Verify issuer.
	if payload.Iss != "tarak" {
		return nil, fmt.Errorf("invalid token issuer: %q", payload.Iss)
	}

	claims := &TokenClaims{
		Subject:   payload.Sub,
		Groups:    payload.Groups,
		IssuedAt:  time.Unix(payload.Iat, 0),
		ExpiresAt: time.Unix(payload.Exp, 0),
	}

	if claims.Expired() {
		return nil, fmt.Errorf("token has expired at %s", claims.ExpiresAt)
	}

	return claims, nil
}

// sign computes HMAC-SHA256 over the given input using the signer's secret.
func (s *TokenSigner) sign(input string) []byte {
	h := hmac.New(sha256.New, s.secret)
	h.Write([]byte(input))
	return h.Sum(nil)
}

// ParseBearerToken extracts the token value from an Authorization header value.
// The header value must be of the form "Bearer <token>".
func ParseBearerToken(authHeader string) (string, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", fmt.Errorf("authorization header is not a Bearer token")
	}
	token := strings.TrimSpace(authHeader[len(prefix):])
	if token == "" {
		return "", fmt.Errorf("empty Bearer token")
	}
	return token, nil
}
