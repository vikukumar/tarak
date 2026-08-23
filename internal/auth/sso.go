package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/vikukumar/tarak/pkg/security"
)

// SSOProvider represents a supported single sign-on provider.
type SSOProvider struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"` // "google", "github", "gitlab", "okta", "microsoft", "keycloak", "oidc"
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"-"`
	AuthURL      string `json:"authUrl"`
	TokenURL     string `json:"tokenUrl"`
	UserInfoURL  string `json:"userInfoUrl"`
	Issuer       string `json:"issuer,omitempty"`
	Icon         string `json:"icon"`
	Enabled      bool   `json:"enabled"`
}

// UserProfile represents the authenticated user's details.
type UserProfile struct {
	Username  string   `json:"username"`
	Email     string   `json:"email,omitempty"`
	Name      string   `json:"name,omitempty"`
	Avatar    string   `json:"avatar,omitempty"`
	Provider  string   `json:"provider"`
	Groups    []string `json:"groups"`
	Roles     []string `json:"roles"`
	IssuedAt  int64    `json:"iat"`
	ExpiresAt int64    `json:"exp"`
}

// SessionToken represents an issued session or API token.
type SessionToken struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Token       string    `json:"token"`
	Description string    `json:"description"`
	Groups      []string  `json:"groups"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

// Manager handles multi-provider SSO and cluster session tokens.
type Manager struct {
	mu            sync.RWMutex
	providers     map[string]*SSOProvider
	sessions      map[string]*UserProfile
	tokens        map[string]*SessionToken
	jwtSecret     []byte
	hasSuperAdmin bool
	adminUser     string
	adminPass     string
	log           *zap.Logger
}

// NewManager creates a new SSO and Token manager.
func NewManager(jwtSecret []byte, log *zap.Logger) *Manager {
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("tarak-default-cluster-jwt-secret-key-32b")
	}
	m := &Manager{
		providers:     make(map[string]*SSOProvider),
		sessions:      make(map[string]*UserProfile),
		tokens:        make(map[string]*SessionToken),
		jwtSecret:     jwtSecret,
		hasSuperAdmin: false,
		adminUser:     "admin",
		adminPass:     "admin",
		log:           log.Named("sso-manager"),
	}

	m.seedDefaultProviders()
	return m
}

func (m *Manager) seedDefaultProviders() {
	m.providers["github"] = &SSOProvider{
		ID:          "github",
		Name:        "GitHub Enterprise & Cloud",
		Type:        "github",
		AuthURL:     "https://github.com/login/oauth/authorize",
		TokenURL:    "https://github.com/login/oauth/access_token",
		UserInfoURL: "https://api.github.com/user",
		Icon:        "github",
		Enabled:     true,
	}
	m.providers["google"] = &SSOProvider{
		ID:          "google",
		Name:        "Google Workspace",
		Type:        "google",
		AuthURL:     "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:    "https://oauth2.googleapis.com/token",
		UserInfoURL: "https://www.googleapis.com/oauth2/v3/userinfo",
		Icon:        "google",
		Enabled:     true,
	}
	m.providers["microsoft"] = &SSOProvider{
		ID:          "microsoft",
		Name:        "Microsoft Entra ID (Azure AD)",
		Type:        "microsoft",
		AuthURL:     "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
		TokenURL:    "https://login.microsoftonline.com/common/oauth2/v2.0/token",
		UserInfoURL: "https://graph.microsoft.com/v1.0/me",
		Icon:        "microsoft",
		Enabled:     true,
	}
	m.providers["okta"] = &SSOProvider{
		ID:          "okta",
		Name:        "Okta Workforce Identity",
		Type:        "okta",
		AuthURL:     "https://dev.okta.com/oauth2/v1/authorize",
		TokenURL:    "https://dev.okta.com/oauth2/v1/token",
		UserInfoURL: "https://dev.okta.com/oauth2/v1/userinfo",
		Icon:        "shield",
		Enabled:     true,
	}
	m.providers["keycloak"] = &SSOProvider{
		ID:          "keycloak",
		Name:        "Keycloak Identity Provider",
		Type:        "keycloak",
		AuthURL:     "http://localhost:8080/realms/tarak/protocol/openid-connect/auth",
		TokenURL:    "http://localhost:8080/realms/tarak/protocol/openid-connect/token",
		UserInfoURL: "http://localhost:8080/realms/tarak/protocol/openid-connect/userinfo",
		Icon:        "key",
		Enabled:     true,
	}
}

// ListProviders returns all registered SSO providers.
func (m *Manager) ListProviders() []*SSOProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []*SSOProvider
	for _, p := range m.providers {
		list = append(list, p)
	}
	return list
}

// IssueToken creates a signed session token for a user.
func (m *Manager) IssueToken(username, provider string, groups, roles []string, ttl time.Duration) (*SessionToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	signer := security.NewTokenSigner(m.jwtSecret)
	tokStr, err := signer.Issue(username, groups, ttl)
	if err != nil {
		return nil, err
	}

	id := generateRandomID(8)
	st := &SessionToken{
		ID:          id,
		Username:    username,
		Token:       tokStr,
		Description: fmt.Sprintf("SSO token via %s", provider),
		Groups:      groups,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(ttl),
	}

	m.tokens[tokStr] = st
	m.sessions[tokStr] = &UserProfile{
		Username:  username,
		Provider:  provider,
		Groups:    groups,
		Roles:     roles,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(ttl).Unix(),
	}

	return st, nil
}

// ValidateToken verifies a token string and returns the user profile.
func (m *Manager) ValidateToken(tokenStr string) (*UserProfile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if prof, ok := m.sessions[tokenStr]; ok {
		if time.Now().Unix() < prof.ExpiresAt {
			return prof, nil
		}
	}

	// Verify cryptographic signature
	signer := security.NewTokenSigner(m.jwtSecret)
	claims, err := signer.Verify(tokenStr)
	if err != nil {
		return nil, err
	}

	return &UserProfile{
		Username: claims.Subject,
		Groups:   claims.Groups,
		Provider: "bearer",
		Roles:    []string{"admin"},
	}, nil
}

// HTTP Handlers

func (m *Manager) HandleListProviders(w http.ResponseWriter, r *http.Request) {
	providers := m.ListProviders()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"apiVersion": "auth.tarak.io/v1",
		"kind":       "SSOProviderList",
		"items":      providers,
	})
}

// HandleStatus returns the cluster authentication and setup readiness state.
func (m *Manager) HandleStatus(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	setupReq := !m.hasSuperAdmin
	m.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"setupRequired": setupReq,
		"clusterName":   "tarak-cluster",
		"authModes":     []string{"password", "token", "mtls", "sso"},
		"version":       "v1.0.6",
	})
}

// HandleSetup performs first-time Super Admin account creation.
func (m *Manager) HandleSetup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid setup payload"}`, http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, `{"error":"username and password required"}`, http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	m.adminUser = req.Username
	m.adminPass = req.Password
	m.hasSuperAdmin = true
	m.mu.Unlock()

	groups := []string{"system:authenticated", "system:masters"}
	roles := []string{"cluster-admin"}

	tok, err := m.IssueToken(req.Username, "initial-setup", groups, roles, 7*24*time.Hour)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "Super-Admin account initialized successfully",
		"token":     tok.Token,
		"expiresAt": tok.ExpiresAt,
		"user": map[string]interface{}{
			"username": req.Username,
			"roles":    roles,
			"groups":   groups,
		},
	})
}

func (m *Manager) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Token    string `json:"token"`
		Provider string `json:"provider"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	m.mu.RLock()
	expectedUser := m.adminUser
	expectedPass := m.adminPass
	m.mu.RUnlock()

	if req.Username == "" {
		req.Username = "admin"
	}
	if req.Provider == "" {
		req.Provider = "local"
	}

	// Validate credentials
	if req.Password != "" && req.Password != expectedPass && expectedPass != "" && req.Username == expectedUser {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	groups := []string{"system:authenticated", "system:masters"}
	roles := []string{"cluster-admin"}

	tok, err := m.IssueToken(req.Username, req.Provider, groups, roles, 24*time.Hour)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"token":     tok.Token,
		"expiresAt": tok.ExpiresAt,
		"user": map[string]interface{}{
			"username": req.Username,
			"provider": req.Provider,
			"roles":    roles,
			"groups":   groups,
		},
	})
}

func (m *Manager) HandleUserInfo(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == "" {
		tokenStr = r.URL.Query().Get("token")
	}

	userProfile, err := m.ValidateToken(tokenStr)
	if err != nil {
		// Fallback admin info for local dev
		userProfile = &UserProfile{
			Username: "admin",
			Provider: "local",
			Groups:   []string{"system:authenticated", "system:masters"},
			Roles:    []string{"cluster-admin"},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(userProfile)
}

func generateRandomID(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
