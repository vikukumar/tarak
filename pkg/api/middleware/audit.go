// Package middleware — Audit logging middleware.
//
// Every API request that reaches a resource handler is recorded in a structured
// audit log entry. The audit log is written to the structured logger and contains:
//
//   - Request timestamp and latency
//   - Authenticated user and groups
//   - HTTP method, path, and response code
//
// Phase 7: configurable audit policy, multiple backends, response body logging.
// Sensitive values (Secret data) are never included in audit log entries.
package middleware

import (
	"crypto/rand"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// AuditEntry is a single audit log record.
type AuditEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	RequestID    string    `json:"requestID"`
	User         string    `json:"user"`
	Groups       []string  `json:"groups"`
	Verb         string    `json:"verb"`
	Resource     string    `json:"resource"`
	Namespace    string    `json:"namespace,omitempty"`
	Name         string    `json:"name,omitempty"`
	ResponseCode int       `json:"responseCode"`
	LatencyMs    float64   `json:"latencyMs"`
}

// verbFromMethod maps HTTP methods to CRUD verbs.
var verbFromMethod = map[string]string{
	http.MethodGet:    "get",
	http.MethodPost:   "create",
	http.MethodPut:    "update",
	http.MethodPatch:  "patch",
	http.MethodDelete: "delete",
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	code int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.code = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) statusCode() int {
	if rw.code == 0 {
		return http.StatusOK
	}
	return rw.code
}

// Audit returns an HTTP middleware that logs every request to the audit log.
func Audit(log *zap.Logger) func(http.Handler) http.Handler {
	if log == nil {
		log = zap.NewNop()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, code: 0}

			next.ServeHTTP(rw, r)

			user := UserFromContext(r.Context())
			verb, ok := verbFromMethod[r.Method]
			if !ok {
				verb = r.Method
			}

			log.Info("audit",
				zap.String("user", user.Username),
				zap.Strings("groups", user.Groups),
				zap.String("verb", verb),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
				zap.Int("code", rw.statusCode()),
				zap.Float64("latencyMs", float64(time.Since(start).Milliseconds())),
				zap.String("remote", r.RemoteAddr),
			)
		})
	}
}

// RequestID returns an HTTP middleware that assigns a unique request ID to each
// request and stores it in the response header X-Request-ID.
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := generateRequestID()
			w.Header().Set("X-Request-ID", id)
			next.ServeHTTP(w, r)
		})
	}
}

// generateRequestID creates a random 8-character request ID.
func generateRequestID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	if _, err := randRead(b); err != nil {
		return "00000000"
	}
	result := make([]byte, 8)
	for i, v := range b {
		result[i] = chars[int(v)%len(chars)]
	}
	return string(result)
}

// randRead fills b with random bytes using crypto/rand.
func randRead(b []byte) (int, error) {
	return rand.Read(b)
}
