package mesh

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestMultiMesh_TenancyAndDNS(t *testing.T) {
	log := zaptest.NewLogger(t)
	mm := NewMultiMeshManager(log)

	// Test DNS resolver
	dns := NewDNSResolver()
	vip, hostnames := dns.RegisterService("finance-mesh", "prod", "payments-svc")
	if !strings.HasPrefix(vip, "240.240.") {
		t.Fatalf("expected 240.240.x.x virtual VIP, got: %s", vip)
	}
	if len(hostnames) < 2 {
		t.Fatalf("expected at least 2 hostnames, got: %v", hostnames)
	}

	resolvedVIP, ok := dns.Resolve("payments-svc.prod.mesh")
	if !ok || resolvedVIP != vip {
		t.Fatalf("expected resolved VIP %s, got %s", vip, resolvedVIP)
	}

	// Test Multi-Mesh Auto-Enrollment
	svc := mm.AutoEnrollWorkload("finance-mesh", "prod", "ledger", 9000, "grpc", []string{"10.244.1.5:9000"})
	if svc.VirtualVIP == "" {
		t.Fatal("expected non-empty virtual VIP")
	}
	if svc.SPIFFEID != "spiffe://finance-mesh.tarak.mesh/ns/prod/sa/ledger" {
		t.Fatalf("unexpected SPIFFE ID: %s", svc.SPIFFEID)
	}

	// Test List Meshes HTTP Handler
	req := httptest.NewRequest("GET", "/apis/mesh.tarak.io/v1/meshes", nil)
	w := httptest.NewRecorder()
	mm.HandleListMeshes(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
}
