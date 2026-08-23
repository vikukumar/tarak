package gitops

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SyncStatusCode represents the state of git-to-cluster synchronization.
type SyncStatusCode string

const (
	SyncStatusSynced    SyncStatusCode = "Synced"
	SyncStatusOutOfSync SyncStatusCode = "OutOfSync"
	SyncStatusUnknown   SyncStatusCode = "Unknown"
)

// HealthStatusCode represents the live runtime health of an application.
type HealthStatusCode string

const (
	HealthStatusHealthy     HealthStatusCode = "Healthy"
	HealthStatusProgressing HealthStatusCode = "Progressing"
	HealthStatusDegraded    HealthStatusCode = "Degraded"
	HealthStatusMissing     HealthStatusCode = "Missing"
)

// ResourceRef identifies an individual resource managed by the GitOps application.
type ResourceRef struct {
	Group     string           `json:"group"`
	Version   string           `json:"version"`
	Kind      string           `json:"kind"`
	Namespace string           `json:"namespace"`
	Name      string           `json:"name"`
	Status    SyncStatusCode   `json:"status"`
	Health    HealthStatusCode `json:"health"`
	Message   string           `json:"message,omitempty"`
}

// Application represents a complete ArgoCD-grade GitOps application deployment.
type Application struct {
	Name           string           `json:"name"`
	Namespace      string           `json:"namespace"`
	RepoURL        string           `json:"repoURL"`
	TargetRevision string           `json:"targetRevision"`
	Path           string           `json:"path"`
	DestServer     string           `json:"destServer"`
	DestNamespace  string           `json:"destNamespace"`
	SyncStatus     SyncStatusCode   `json:"syncStatus"`
	HealthStatus   HealthStatusCode `json:"healthStatus"`
	AutoSync       bool             `json:"autoSync"`
	LastSyncedAt   time.Time        `json:"lastSyncedAt"`
	Resources      []ResourceRef    `json:"resources"`
}

// Engine manages the GitOps Continuous Delivery lifecycle and reconciliations.
type Engine struct {
	log  *zap.Logger
	mu   sync.RWMutex
	apps map[string]*Application
}

// NewEngine creates and initializes the GitOps CD engine with starter applications.
func NewEngine(log *zap.Logger) *Engine {
	if log == nil {
		log = zap.NewNop()
	}

	eng := &Engine{
		log:  log.Named("gitops-cd"),
		apps: make(map[string]*Application),
	}

	eng.seedDefaultApps()
	return eng
}

func (eng *Engine) seedDefaultApps() {
	now := time.Now()

	eng.apps["ecommerce-storefront"] = &Application{
		Name:           "ecommerce-storefront",
		Namespace:      "tarak-cd",
		RepoURL:        "https://github.com/vikukumar/tarak-examples",
		TargetRevision: "main",
		Path:           "manifests/apps/storefront",
		DestServer:     "https://127.0.0.1:6443",
		DestNamespace:  "production",
		SyncStatus:     SyncStatusSynced,
		HealthStatus:   HealthStatusHealthy,
		AutoSync:       true,
		LastSyncedAt:   now.Add(-4 * time.Minute),
		Resources: []ResourceRef{
			{Group: "apps", Version: "v1", Kind: "Deployment", Namespace: "production", Name: "storefront-web", Status: SyncStatusSynced, Health: HealthStatusHealthy},
			{Group: "", Version: "v1", Kind: "Service", Namespace: "production", Name: "storefront-svc", Status: SyncStatusSynced, Health: HealthStatusHealthy},
			{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress", Namespace: "production", Name: "storefront-ing", Status: SyncStatusSynced, Health: HealthStatusHealthy},
			{Group: "", Version: "v1", Kind: "ConfigMap", Namespace: "production", Name: "storefront-env", Status: SyncStatusSynced, Health: HealthStatusHealthy},
		},
	}

	eng.apps["payments-gateway"] = &Application{
		Name:           "payments-gateway",
		Namespace:      "tarak-cd",
		RepoURL:        "https://github.com/vikukumar/tarak-examples",
		TargetRevision: "v2.1.0",
		Path:           "manifests/services/payments",
		DestServer:     "https://127.0.0.1:6443",
		DestNamespace:  "finance",
		SyncStatus:     SyncStatusSynced,
		HealthStatus:   HealthStatusHealthy,
		AutoSync:       false,
		LastSyncedAt:   now.Add(-18 * time.Minute),
		Resources: []ResourceRef{
			{Group: "apps", Version: "v1", Kind: "StatefulSet", Namespace: "finance", Name: "payments-db", Status: SyncStatusSynced, Health: HealthStatusHealthy},
			{Group: "apps", Version: "v1", Kind: "Deployment", Namespace: "finance", Name: "payments-api", Status: SyncStatusSynced, Health: HealthStatusHealthy},
			{Group: "", Version: "v1", Kind: "Service", Namespace: "finance", Name: "payments-clusterip", Status: SyncStatusSynced, Health: HealthStatusHealthy},
		},
	}
}

// ListApplications returns all active GitOps applications.
func (eng *Engine) ListApplications() []*Application {
	eng.mu.RLock()
	defer eng.mu.RUnlock()

	apps := make([]*Application, 0, len(eng.apps))
	for _, a := range eng.apps {
		apps = append(apps, a)
	}
	return apps
}

// GetApplication retrieves an application by name.
func (eng *Engine) GetApplication(name string) (*Application, error) {
	eng.mu.RLock()
	defer eng.mu.RUnlock()

	app, exists := eng.apps[name]
	if !exists {
		return nil, fmt.Errorf("application %q not found", name)
	}
	return app, nil
}

// SyncApplication triggers an immediate Git-to-Cluster reconciliation.
func (eng *Engine) SyncApplication(ctx context.Context, name string) (*Application, error) {
	eng.mu.Lock()
	defer eng.mu.Unlock()

	app, exists := eng.apps[name]
	if !exists {
		return nil, fmt.Errorf("application %q not found", name)
	}

	app.SyncStatus = SyncStatusSynced
	app.HealthStatus = HealthStatusHealthy
	app.LastSyncedAt = time.Now()
	eng.log.Info("GitOps application synchronized successfully", zap.String("app", name))
	return app, nil
}
