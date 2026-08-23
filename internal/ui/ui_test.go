package ui_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vikukumar/tarak/internal/ui"
)

func TestUIHandler_AssetResolution(t *testing.T) {
	handler := ui.Handler()

	// 1. Test HTML Dashboard Route
	req := httptest.NewRequest("GET", "/dashboard/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /dashboard/, got %d", rec.Code)
	}
	cType := rec.Header().Get("Content-Type")
	if !strings.Contains(cType, "text/html") {
		t.Errorf("expected text/html for /dashboard/, got %s", cType)
	}

	// 2. Test Asset Chunk resolution via /dashboard/_next/static/...
	reqJS := httptest.NewRequest("GET", "/dashboard/_next/static/chunks/310vm2bl3xxpt.js", nil)
	recJS := httptest.NewRecorder()
	handler.ServeHTTP(recJS, reqJS)

	// Even if specific random hash changes, the content type must never be text/html
	if recJS.Code == http.StatusOK {
		jsType := recJS.Header().Get("Content-Type")
		if strings.Contains(jsType, "text/html") {
			t.Errorf("JS chunk was returned as text/html! Expected application/javascript, got %s", jsType)
		}
	}
}
