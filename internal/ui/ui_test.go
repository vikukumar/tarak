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

	// 2. Test Direct Icon PNG
	reqIcon := httptest.NewRequest("GET", "/assets/icon.png", nil)
	recIcon := httptest.NewRecorder()
	handler.ServeHTTP(recIcon, reqIcon)

	if recIcon.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /assets/icon.png, got %d", recIcon.Code)
	}
	if recIcon.Header().Get("Content-Type") != "image/png" {
		t.Errorf("expected image/png for icon.png, got %s", recIcon.Header().Get("Content-Type"))
	}

	// 3. Test Horizontal Logo via /dashboard/assets/...
	reqLogo := httptest.NewRequest("GET", "/dashboard/assets/tarak_logo_horizontal.png", nil)
	recLogo := httptest.NewRecorder()
	handler.ServeHTTP(recLogo, reqLogo)

	if recLogo.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /dashboard/assets/tarak_logo_horizontal.png, got %d", recLogo.Code)
	}
	if recLogo.Header().Get("Content-Type") != "image/png" {
		t.Errorf("expected image/png for tarak_logo_horizontal.png, got %s", recLogo.Header().Get("Content-Type"))
	}

	// 4. Test Vertical Logo via /assets/...
	reqVLogo := httptest.NewRequest("GET", "/assets/tarak_logo_vertical.png", nil)
	recVLogo := httptest.NewRecorder()
	handler.ServeHTTP(recVLogo, reqVLogo)

	if recVLogo.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /assets/tarak_logo_vertical.png, got %d", recVLogo.Code)
	}
	if recVLogo.Header().Get("Content-Type") != "image/png" {
		t.Errorf("expected image/png for tarak_logo_vertical.png, got %s", recVLogo.Header().Get("Content-Type"))
	}

	// 5. Test Favicon ICO
	reqFav := httptest.NewRequest("GET", "/favicon.ico", nil)
	recFav := httptest.NewRecorder()
	handler.ServeHTTP(recFav, reqFav)

	if recFav.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /favicon.ico, got %d", recFav.Code)
	}
	if recFav.Header().Get("Content-Type") != "image/x-icon" {
		t.Errorf("expected image/x-icon for favicon.ico, got %s", recFav.Header().Get("Content-Type"))
	}
}
