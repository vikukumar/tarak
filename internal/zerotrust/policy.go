package zerotrust

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// ZeroTrustRule defines a cryptographic rule for allowing traffic between identities.
type ZeroTrustRule struct {
	SourceIdentity string   `json:"sourceIdentity"` // e.g. spiffe://tarak.mesh/ns/default/sa/frontend
	TargetService  string   `json:"targetService"`  // e.g. payment-service
	AllowedMethods []string `json:"allowedMethods"` // e.g. GET, POST
	AllowedPaths   []string `json:"allowedPaths"`   // e.g. /api/v1/pay
	Action         string   `json:"action"`         // "ALLOW" or "DENY"
}

// Policy is a collection of Zero-Trust security rules.
type Policy struct {
	Name        string          `json:"name"`
	Namespace   string          `json:"namespace"`
	DefaultDeny bool            `json:"defaultDeny"`
	Rules       []ZeroTrustRule `json:"rules"`
}

// Manager manages Zero-Trust network policies and attestation.
type Manager struct {
	mu       sync.RWMutex
	log      *zap.Logger
	policies map[string]*Policy // key: ns/name
}

// NewManager creates a new Zero-Trust policy manager.
func NewManager(log *zap.Logger) *Manager {
	return &Manager{
		log:      log.Named("zerotrust"),
		policies: make(map[string]*Policy),
	}
}

// RegisterPolicy registers a Zero-Trust policy in the manager.
func (m *Manager) RegisterPolicy(p *Policy) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p.Namespace == "" {
		p.Namespace = "default"
	}
	key := fmt.Sprintf("%s/%s", p.Namespace, p.Name)
	m.policies[key] = p
}

// Evaluate checks if a source identity is permitted to call the target service/method.
func (m *Manager) Evaluate(sourceIdentity, targetService, method, path string) (bool, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, pol := range m.policies {
		for _, rule := range pol.Rules {
			if (rule.SourceIdentity == "*" || rule.SourceIdentity == sourceIdentity) &&
				(rule.TargetService == "*" || rule.TargetService == targetService) {

				// Check method
				methodMatch := false
				for _, m := range rule.AllowedMethods {
					if m == "*" || m == method {
						methodMatch = true
						break
					}
				}

				if methodMatch && rule.Action == "ALLOW" {
					return true, fmt.Sprintf("allowed by rule in policy '%s'", pol.Name)
				}
			}
		}
	}

	return false, "blocked by default-deny zero-trust policy"
}

// HandleListPolicies lists all active zero-trust policies.
func (m *Manager) HandleListPolicies(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*Policy, 0, len(m.policies))
	for _, p := range m.policies {
		list = append(list, p)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"items": list,
		"total": len(list),
	})
}

// HandleCreatePolicy creates or updates a zero-trust policy.
func (m *Manager) HandleCreatePolicy(w http.ResponseWriter, r *http.Request) {
	var req Policy
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Namespace == "" {
		req.Namespace = "default"
	}
	if req.Name == "" {
		http.Error(w, "missing policy name", http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	key := fmt.Sprintf("%s/%s", req.Namespace, req.Name)
	m.policies[key] = &req
	m.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(req)
}

// HandleEvaluate evaluates an access request via HTTP API.
func (m *Manager) HandleEvaluate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SourceIdentity string `json:"sourceIdentity"`
		TargetService  string `json:"targetService"`
		Method         string `json:"method"`
		Path           string `json:"path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	allowed, reason := m.Evaluate(req.SourceIdentity, req.TargetService, req.Method, req.Path)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"allowed": allowed,
		"reason":  reason,
	})
}

// HandleDeletePolicy deletes a policy.
func (m *Manager) HandleDeletePolicy(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	key := fmt.Sprintf("%s/%s", ns, name)

	m.mu.Lock()
	delete(m.policies, key)
	m.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}
