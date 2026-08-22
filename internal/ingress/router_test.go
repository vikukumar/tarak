package ingress

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"go.uber.org/zap"
)

func TestIngressRouter_Routing(t *testing.T) {
	// Start a backend test HTTP server
	backendSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "Hello from backend: path=%s host=%s", r.URL.Path, r.Host)
	}))
	defer backendSrv.Close()

	backendURL, err := url.Parse(backendSrv.URL)
	if err != nil {
		t.Fatalf("parse backend url: %v", err)
	}

	router := NewRouter(zap.NewNop())
	router.UpdateRoutes([]Route{
		{
			Host:        "app.vikshro.in",
			Path:        "/api",
			BackendURL:  backendURL,
			ServiceName: "web-app-svc",
			ServicePort: 80,
			Namespace:   "default",
		},
		{
			Host:        "*.vikshro.in",
			Path:        "/",
			BackendURL:  backendURL,
			ServiceName: "wildcard-svc",
			ServicePort: 80,
			Namespace:   "default",
		},
	})

	// 1. Exact host and path match
	req := httptest.NewRequest("GET", "http://app.vikshro.in/api/v1/users", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	// 2. Wildcard host match
	req2 := httptest.NewRequest("GET", "http://blog.vikshro.in/posts", nil)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected status 200 for wildcard, got %d", rec2.Code)
	}

	// 3. Unmatched host should return 404
	req3 := httptest.NewRequest("GET", "http://other-domain.com/", nil)
	rec3 := httptest.NewRecorder()
	router.ServeHTTP(rec3, req3)

	if rec3.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 for unmatched host, got %d", rec3.Code)
	}
}
