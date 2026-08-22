package ui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUIHandler_ServesDashboard(t *testing.T) {
	h := Handler()

	// Test GET /dashboard/
	req := httptest.NewRequest("GET", "/dashboard/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "TARAK") && !strings.Contains(body, "root") {
		t.Fatalf("expected response body to contain TARAK HTML, got: %s", body)
	}
}
