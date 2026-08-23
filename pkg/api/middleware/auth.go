// Package middleware provides HTTP middleware for the Tarak API server.
//
// Authentication middleware implements two mechanisms:
//  1. X.509 client certificate — CN=username, O=group
//  2. Bearer token — HMAC-SHA256 signed JWT
//
// On successful authentication, the user identity is stored in the request context
// using a typed key, and downstream handlers retrieve it with UserFromContext.
//
// Anonymous requests are allowed only for /healthz, /readyz, /livez, and /openapi.
// All other requests without credentials are rejected with 401.
//
// Phase 1 does not enforce RBAC (Phase 7). Any authenticated user is permitted.
package middleware

import (
	"context"
	"crypto/x509"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/vikukumar/tarak/pkg/security"
)

// ─── User identity ────────────────────────────────────────────────────────────

// UserInfo represents an authenticated API user.
type UserInfo struct {
	// Username is the authenticated username.
	Username string
	// Groups is the list of groups the user belongs to.
	Groups []string
	// AuthMethod is how the user was authenticated ("cert" or "token").
	AuthMethod string
}

// Anonymous is the identity used for unauthenticated requests on public endpoints.
var Anonymous = &UserInfo{Username: "system:anonymous", Groups: []string{"system:unauthenticated"}}

// contextKey is the unexported type used for context values.
type contextKey int

const userContextKey contextKey = 1

// WithUser returns a new context carrying the given UserInfo.
func WithUser(ctx context.Context, u *UserInfo) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}

// UserFromContext retrieves the UserInfo from a request context.
// Returns Anonymous if no user was set.
func UserFromContext(ctx context.Context) *UserInfo {
	u, ok := ctx.Value(userContextKey).(*UserInfo)
	if !ok || u == nil {
		return Anonymous
	}
	return u
}

// ─── Auth middleware ──────────────────────────────────────────────────────────

// AuthOptions configures the authentication middleware.
type AuthOptions struct {
	// CertPool is the trusted CA pool used to verify client certificates.
	CertPool *x509.CertPool
	// TokenSigner is used to verify bearer tokens.
	TokenSigner *security.TokenSigner
	// StaticTokens is an optional list of pre-issued static tokens (for bootstrap).
	StaticTokens []security.StaticToken
	// Log is the structured logger.
	Log *zap.Logger
	// AllowInsecure allows unauthenticated requests on non-public endpoints.
	// Only set true in development/test environments.
	AllowInsecure bool
}

// isPublicPath determines if an endpoint can be accessed without prior authentication.
func isPublicPath(p string) bool {
	if p == "/healthz" || p == "/readyz" || p == "/livez" || p == "/openapi/v2" || p == "/api" || p == "/apis" || p == "/favicon.ico" {
		return true
	}
	// Embedded UI and assets
	if p == "/dashboard" || strings.HasPrefix(p, "/dashboard/") || strings.HasPrefix(p, "/assets/") {
		return true
	}
	// Auth, Setup & SSO login endpoints
	if strings.HasPrefix(p, "/apis/auth.tarak.io/v1/login") ||
		strings.HasPrefix(p, "/apis/auth.tarak.io/v1/setup") ||
		strings.HasPrefix(p, "/apis/auth.tarak.io/v1/status") ||
		strings.HasPrefix(p, "/apis/auth.tarak.io/v1/providers") {
		return true
	}
	return false
}

// Auth returns an HTTP middleware that authenticates incoming requests.
func Auth(opts AuthOptions) func(http.Handler) http.Handler {
	if opts.Log == nil {
		opts.Log = zap.NewNop()
	}
	staticMap := make(map[string]security.StaticToken, len(opts.StaticTokens))
	for _, st := range opts.StaticTokens {
		staticMap[st.Token] = st
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow public endpoints without auth.
			if isPublicPath(r.URL.Path) {
				r = r.WithContext(WithUser(r.Context(), Anonymous))
				next.ServeHTTP(w, r)
				return
			}

			user, err := authenticate(r, opts.CertPool, opts.TokenSigner, staticMap)
			if err != nil {
				if opts.AllowInsecure {
					// In insecure mode, treat unauthenticated requests as cluster-admin.
					user = &UserInfo{
						Username:   "system:cluster-admin",
						Groups:     []string{"system:masters"},
						AuthMethod: "insecure",
					}
				} else {
					opts.Log.Info("authentication failed",
						zap.String("path", r.URL.Path),
						zap.String("remote", r.RemoteAddr),
						zap.Error(err),
					)
					writeStatus(w, http.StatusUnauthorized, "Unauthorized", "authentication required")
					return
				}
			}

			opts.Log.Debug("authenticated",
				zap.String("user", user.Username),
				zap.String("method", user.AuthMethod),
				zap.String("path", r.URL.Path),
			)

			r = r.WithContext(WithUser(r.Context(), user))
			next.ServeHTTP(w, r)
		})
	}
}

// authenticate attempts to authenticate the request using:
// 1. X.509 client certificate (mTLS)
// 2. Bearer token (Authorization header)
// Returns an error if no valid credential is found.
func authenticate(
	r *http.Request,
	certPool *x509.CertPool,
	signer *security.TokenSigner,
	staticTokens map[string]security.StaticToken,
) (*UserInfo, error) {
	// ── 1. Client certificate ─────────────────────────────────────────────
	if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
		cert := r.TLS.PeerCertificates[0]
		// Verify the cert against the trusted CA pool.
		if certPool != nil {
			opts := x509.VerifyOptions{
				Roots:     certPool,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}
			if _, err := cert.Verify(opts); err == nil {
				return &UserInfo{
					Username:   cert.Subject.CommonName,
					Groups:     cert.Subject.Organization,
					AuthMethod: "cert",
				}, nil
			}
		}
	}

	// ── 2. Bearer token ───────────────────────────────────────────────────
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		tokenVal, err := security.ParseBearerToken(authHeader)
		if err == nil {
			// Check static tokens first (no crypto needed).
			if st, ok := staticTokens[tokenVal]; ok {
				return &UserInfo{
					Username:   st.User,
					Groups:     st.Groups,
					AuthMethod: "static-token",
				}, nil
			}
			// Verify signed token.
			if signer != nil {
				claims, err := signer.Verify(tokenVal)
				if err == nil {
					return &UserInfo{
						Username:   claims.Subject,
						Groups:     claims.Groups,
						AuthMethod: "token",
					}, nil
				}
			}
		}
	}

	return nil, errUnauthenticated
}

var errUnauthenticated = &authError{"no valid credential found"}

type authError struct{ msg string }

func (e *authError) Error() string { return e.msg }

// IsAdmin returns true if the user is in the system:masters group.
func (u *UserInfo) IsAdmin() bool {
	for _, g := range u.Groups {
		if g == "system:masters" {
			return true
		}
	}
	return false
}

// writeStatus writes a Kubernetes-style Status JSON error response.
func writeStatus(w http.ResponseWriter, code int, reason, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	// Minimal Status object — full Status type is in api/meta.
	status := struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
		Status     string `json:"status"`
		Message    string `json:"message"`
		Reason     string `json:"reason"`
		Code       int    `json:"code"`
	}{
		APIVersion: "v1",
		Kind:       "Status",
		Status:     "Failure",
		Message:    message,
		Reason:     reason,
		Code:       code,
	}
	_ = jsonEncode(w, status)
}

// jsonEncode writes v as JSON to w.
func jsonEncode(w http.ResponseWriter, v interface{}) error {
	return encodeJSON(w, v)
}
