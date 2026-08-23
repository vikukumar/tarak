package loadbalancer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestLoadBalancer_DetectorAndPool(t *testing.T) {
	log := zaptest.NewLogger(t)
	detector := NewIPDetector(log)

	lanIP := detector.DetectLocalLANIP()
	if lanIP == "" {
		t.Fatal("expected non-empty local LAN IP")
	}

	pool := NewIPPool("203.0.113.50", lanIP, log)

	// Test public VIP allocation
	vip1, err := pool.Allocate("default/frontend-svc", true)
	if err != nil {
		t.Fatalf("unexpected error allocating VIP: %v", err)
	}
	if vip1 != "203.0.113.50" {
		t.Fatalf("expected public VIP 203.0.113.50, got %s", vip1)
	}

	// Test pool VIP allocation
	vip2, err := pool.Allocate("default/backend-svc", false)
	if err != nil {
		t.Fatalf("unexpected error allocating VIP: %v", err)
	}
	if vip2 == "" {
		t.Fatal("expected non-empty pool VIP")
	}

	// Test controller status
	ctrl := NewController(log)
	ctrl.Start(context.Background())

	req := httptest.NewRequest("GET", "/apis/networking.tarak.io/v1/loadbalancer/status", nil)
	w := httptest.NewRecorder()
	ctrl.HandleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Healthy") {
		t.Fatalf("expected body to contain Healthy, got: %s", w.Body.String())
	}
}
