package auth

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestSSOManager_ProvidersAndTokens(t *testing.T) {
	log, _ := zap.NewDevelopment()
	mgr := NewManager([]byte("test-jwt-secret-key-32b-long!!"), log)

	providers := mgr.ListProviders()
	if len(providers) < 3 {
		t.Fatalf("expected at least 3 default SSO providers, got %d", len(providers))
	}

	tok, err := mgr.IssueToken("alice", "github", []string{"developers"}, []string{"developer"}, 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	if tok.Token == "" {
		t.Fatal("issued token is empty")
	}

	prof, err := mgr.ValidateToken(tok.Token)
	if err != nil {
		t.Fatalf("failed to validate valid token: %v", err)
	}

	if prof.Username != "alice" {
		t.Errorf("expected username alice, got %s", prof.Username)
	}
}
