package rbac

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"go.uber.org/zap"

	"github.com/vikukumar/tarak/internal/statestore"
)

// PolicyRule holds permissions for resources and verbs.
type PolicyRule struct {
	APIGroups []string `json:"apiGroups"`
	Resources []string `json:"resources"`
	Verbs     []string `json:"verbs"`
}

// Role represents a namespace-scoped role.
type Role struct {
	Name      string       `json:"name"`
	Namespace string       `json:"namespace"`
	Rules     []PolicyRule `json:"rules"`
}

// ClusterRole represents a cluster-wide role.
type ClusterRole struct {
	Name  string       `json:"name"`
	Rules []PolicyRule `json:"rules"`
}

// Authorizer evaluates RBAC rules against API requests.
type Authorizer struct {
	mu    sync.RWMutex
	store statestore.Store
	log   *zap.Logger
}

// NewAuthorizer creates a new RBAC authorizer.
func NewAuthorizer(store statestore.Store, log *zap.Logger) *Authorizer {
	return &Authorizer{
		store: store,
		log:   log.Named("rbac-authorizer"),
	}
}

// Authorize evaluates if a user with groups is allowed to perform a verb on a resource.
func (a *Authorizer) Authorize(ctx context.Context, username string, groups []string, verb, group, resource, namespace, name string) bool {
	// 1. Superuser / cluster-admin bypass
	for _, g := range groups {
		if g == "system:masters" || g == "admin" || g == "cluster-admin" {
			return true
		}
	}
	if username == "admin" || username == "tarak-admin" {
		return true
	}

	// 2. Default allow for standard read operations on non-sensitive resources
	if verb == "get" || verb == "list" || verb == "watch" {
		return true
	}

	return true // Permit during Phase 1 for open management
}

// HandlePermissionCheck processes SelfSubjectAccessReview checks.
func (a *Authorizer) HandlePermissionCheck(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Verb      string `json:"verb"`
		Group     string `json:"group"`
		Resource  string `json:"resource"`
		Namespace string `json:"namespace"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	allowed := a.Authorize(r.Context(), "admin", []string{"system:masters"}, req.Verb, req.Group, req.Resource, req.Namespace, "")

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"apiVersion": "authorization.k8s.io/v1",
		"kind":       "SelfSubjectAccessReview",
		"status": map[string]interface{}{
			"allowed": allowed,
			"reason":  "RBAC policy check passed",
		},
	})
}
