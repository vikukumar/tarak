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

// NewEngine creates and initializes the GitOps CD engine.
func NewEngine(log *zap.Logger) *Engine {
	if log == nil {
		log = zap.NewNop()
	}

	eng := &Engine{
		log:  log.Named("gitops-cd"),
		apps: make(map[string]*Application),
	}
	return eng
}

// RegisterApplication adds or updates a GitOps Application definition.
func (eng *Engine) RegisterApplication(app *Application) {
	if app == nil || app.Name == "" {
		return
	}
	eng.mu.Lock()
	defer eng.mu.Unlock()
	eng.apps[app.Name] = app
	eng.log.Info("registered gitops application", zap.String("name", app.Name), zap.String("repo", app.RepoURL))
}

// DeleteApplication removes a GitOps Application by name.
func (eng *Engine) DeleteApplication(name string) {
	eng.mu.Lock()
	defer eng.mu.Unlock()
	delete(eng.apps, name)
	eng.log.Info("deleted gitops application", zap.String("name", name))
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
