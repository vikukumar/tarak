package zerotrust

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestZeroTrust_Evaluation(t *testing.T) {
	log := zaptest.NewLogger(t)
	mgr := NewManager(log)

	// Test default-deny on unauthorized service
	allowed, _ := mgr.Evaluate("spiffe://tarak.mesh/ns/default/sa/guest", "api-service", "GET", "/api/v1/users")
	if allowed {
		t.Fatal("expected unauthorized caller to be denied")
	}

	// Test authorized caller
	allowed, _ = mgr.Evaluate("spiffe://tarak.mesh/ns/default/sa/frontend", "api-service", "GET", "/api/v1/users")
	if !allowed {
		t.Fatal("expected frontend identity to be allowed")
	}

	// Test HTTP evaluation endpoint
	evalReq := map[string]string{
		"sourceIdentity": "spiffe://tarak.mesh/ns/default/sa/frontend",
		"targetService":  "api-service",
		"method":         "GET",
		"path":           "/api/v1/data",
	}
	body, _ := json.Marshal(evalReq)
	req := httptest.NewRequest("POST", "/apis/security.tarak.io/v1/zerotrust/evaluate", bytes.NewReader(body))
	w := httptest.NewRecorder()
	mgr.HandleEvaluate(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp struct {
		Allowed bool   `json:"allowed"`
		Reason  string `json:"reason"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Allowed {
		t.Fatalf("expected allowed true, got false (%s)", resp.Reason)
	}
}
