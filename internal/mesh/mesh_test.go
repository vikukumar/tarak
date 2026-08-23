package mesh

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
)

func TestMesh_CanaryAndCircuitBreaker(t *testing.T) {
	log := zaptest.NewLogger(t)
	engine := NewEngine(log)

	route := &TrafficRoute{
		Name:      "frontend-canary",
		Namespace: "default",
		Host:      "frontend.tarak.mesh",
		Destinations: []Destination{
			{Service: "frontend-v1", Subset: "v1", Weight: 90},
			{Service: "frontend-v2", Subset: "v2", Weight: 10},
		},
	}

	engine.RegisterRoute(route)

	// Test canary selection
	counts := make(map[string]int)
	for i := 0; i < 100; i++ {
		dest := engine.SelectDestination(route)
		counts[dest.Subset]++
	}

	if counts["v1"] == 0 {
		t.Fatalf("expected v1 to receive canary traffic, got: %v", counts)
	}

	// Test Circuit Breaker
	cb := NewCircuitBreaker(2, 50*time.Millisecond)
	if !cb.Allow("db.internal") {
		t.Fatal("expected host to be allowed initially")
	}

	cb.ReportFailure("db.internal")
	cb.ReportFailure("db.internal")

	if cb.Allow("db.internal") {
		t.Fatal("expected breaker to be tripped after 2 failures")
	}

	time.Sleep(60 * time.Millisecond)
	if !cb.Allow("db.internal") {
		t.Fatal("expected breaker to reset after cooldown")
	}

	// Test Rate Limiter
	rl := NewRateLimiter()
	if !rl.Allow("client-1", 2, 1) {
		t.Fatal("expected first request to be allowed")
	}
	if !rl.Allow("client-1", 2, 1) {
		t.Fatal("expected second request to be allowed")
	}
	if rl.Allow("client-1", 2, 1) {
		t.Fatal("expected third request to be rate-limited")
	}

	// Test SPIFFE Generation
	ident, err := GenerateSPIFFEIdentity("default", "web-service", "tarak.mesh")
	if err != nil {
		t.Fatalf("unexpected SPIFFE generation error: %v", err)
	}
	if ident.SPIFFEID != "spiffe://tarak.mesh/ns/default/sa/web-service" {
		t.Fatalf("unexpected SPIFFE ID: %s", ident.SPIFFEID)
	}

	// Test HTTP Handlers
	routeBody, _ := json.Marshal(route)
	req := httptest.NewRequest("POST", "/apis/mesh.tarak.io/v1/routes", bytes.NewReader(routeBody))
	w := httptest.NewRecorder()
	engine.HandleCreateRoute(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", w.Code)
	}
}
