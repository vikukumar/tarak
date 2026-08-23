package mesh

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// MultiMeshManager coordinates all mesh tenants and policies.
type MultiMeshManager struct {
	mu                  sync.RWMutex
	log                 *zap.Logger
	meshes              map[string]*Mesh
	services            map[string]map[string]*MeshService         // meshName -> svcName -> MeshService
	externalServices    map[string]map[string]*MeshExternalService // meshName -> extName -> MeshExternalService
	trafficPermissions  map[string]map[string]*MeshTrafficPermission
	passthroughPolicies map[string]map[string]*MeshPassthroughPolicy
	proxyPatches        map[string]map[string]*MeshProxyPatch
	dnsResolver         *DNSResolver
}

// NewMultiMeshManager creates a new MultiMeshManager.
func NewMultiMeshManager(log *zap.Logger) *MultiMeshManager {
	m := &MultiMeshManager{
		log:                 log.Named("multi-mesh"),
		meshes:              make(map[string]*Mesh),
		services:            make(map[string]map[string]*MeshService),
		externalServices:    make(map[string]map[string]*MeshExternalService),
		trafficPermissions:  make(map[string]map[string]*MeshTrafficPermission),
		passthroughPolicies: make(map[string]map[string]*MeshPassthroughPolicy),
		proxyPatches:        make(map[string]map[string]*MeshProxyPatch),
		dnsResolver:         NewDNSResolver(),
	}

	// Bootstrap default mesh
	defaultMesh := &Mesh{
		Name: "default",
		MTLS: MTLSConfig{
			Enabled:     true,
			Mode:        "Strict",
			TrustDomain: "tarak.mesh",
			Backend:     "builtin",
		},
		Passthrough: "Passthrough",
		Metrics:     "Prometheus",
		Tracing:     "OpenTelemetry",
		Logging:     "StructuredJSON",
		CreatedAt:   time.Now(),
	}
	m.meshes["default"] = defaultMesh
	m.services["default"] = make(map[string]*MeshService)
	m.externalServices["default"] = make(map[string]*MeshExternalService)
	m.trafficPermissions["default"] = make(map[string]*MeshTrafficPermission)
	m.passthroughPolicies["default"] = make(map[string]*MeshPassthroughPolicy)
	m.proxyPatches["default"] = make(map[string]*MeshProxyPatch)

	// Pre-seed sample service discovery in default mesh
	m.AutoEnrollWorkload("default", "default", "frontend", 80, "http", []string{"10.244.0.12:80"})
	m.AutoEnrollWorkload("default", "default", "api-service", 8080, "http", []string{"10.244.0.14:8080"})

	// Pre-seed an external service
	m.externalServices["default"]["stripe-api"] = &MeshExternalService{
		Name:        "stripe-api",
		Mesh:        "default",
		Host:        "api.stripe.com",
		Port:        443,
		TLSRequired: true,
		SNI:         "api.stripe.com",
	}

	// Pre-seed traffic permission
	m.trafficPermissions["default"]["allow-frontend-to-api"] = &MeshTrafficPermission{
		Name: "allow-frontend-to-api",
		Mesh: "default",
		From: []PermissionMatch{{Service: "frontend"}},
		To:   []PermissionMatch{{Service: "api-service"}},
		Action: "ALLOW",
	}

	return m
}

// AutoEnrollWorkload registers a workload discovered via tarak.io/mesh annotations or labels.
func (m *MultiMeshManager) AutoEnrollWorkload(meshName, namespace, serviceName string, port int, proto string, endpoints []string) *MeshService {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.meshes[meshName]; !ok {
		m.meshes[meshName] = &Mesh{
			Name: meshName,
			MTLS: MTLSConfig{
				Enabled:     true,
				Mode:        "Strict",
				TrustDomain: fmt.Sprintf("%s.tarak.mesh", meshName),
				Backend:     "builtin",
			},
			Passthrough: "Passthrough",
			Metrics:     "Prometheus",
			Tracing:     "OpenTelemetry",
			Logging:     "StructuredJSON",
			CreatedAt:   time.Now(),
		}
		m.services[meshName] = make(map[string]*MeshService)
		m.externalServices[meshName] = make(map[string]*MeshExternalService)
		m.trafficPermissions[meshName] = make(map[string]*MeshTrafficPermission)
		m.passthroughPolicies[meshName] = make(map[string]*MeshPassthroughPolicy)
		m.proxyPatches[meshName] = make(map[string]*MeshProxyPatch)
	}

	vip, hostnames := m.dnsResolver.RegisterService(meshName, namespace, serviceName)
	spiffeID := fmt.Sprintf("spiffe://%s/ns/%s/sa/%s", m.meshes[meshName].MTLS.TrustDomain, namespace, serviceName)

	svc := &MeshService{
		Name:       serviceName,
		Mesh:       meshName,
		Namespace:  namespace,
		Hostnames:  hostnames,
		VirtualVIP: vip,
		Port:       port,
		Protocol:   proto,
		Endpoints:  endpoints,
		Status:     "Healthy",
		SPIFFEID:   spiffeID,
	}

	m.services[meshName][serviceName] = svc
	m.log.Info("auto-enrolled workload into mesh",
		zap.String("mesh", meshName),
		zap.String("service", serviceName),
		zap.String("vip", vip),
	)
	return svc
}

// ── HTTP Handlers ─────────────────────────────────────────────────────────────

// HandleListMeshes returns all active mesh tenants.
func (m *MultiMeshManager) HandleListMeshes(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*Mesh, 0, len(m.meshes))
	for _, mesh := range m.meshes {
		list = append(list, mesh)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"items": list,
		"total": len(list),
	})
}

// HandleCreateMesh creates a new mesh tenant.
func (m *MultiMeshManager) HandleCreateMesh(w http.ResponseWriter, r *http.Request) {
	var req Mesh
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "missing mesh name", http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	if req.MTLS.TrustDomain == "" {
		req.MTLS.TrustDomain = fmt.Sprintf("%s.tarak.mesh", req.Name)
	}
	if req.MTLS.Mode == "" {
		req.MTLS.Mode = "Strict"
	}
	req.CreatedAt = time.Now()

	m.meshes[req.Name] = &req
	if _, ok := m.services[req.Name]; !ok {
		m.services[req.Name] = make(map[string]*MeshService)
		m.externalServices[req.Name] = make(map[string]*MeshExternalService)
		m.trafficPermissions[req.Name] = make(map[string]*MeshTrafficPermission)
		m.passthroughPolicies[req.Name] = make(map[string]*MeshPassthroughPolicy)
		m.proxyPatches[req.Name] = make(map[string]*MeshProxyPatch)
	}
	m.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(req)
}

// HandleListServices returns discovered services in a mesh.
func (m *MultiMeshManager) HandleListServices(w http.ResponseWriter, r *http.Request) {
	meshName := chi.URLParam(r, "mesh")
	m.mu.RLock()
	defer m.mu.RUnlock()

	svcsMap, ok := m.services[meshName]
	if !ok {
		svcsMap = m.services["default"]
	}

	list := make([]*MeshService, 0, len(svcsMap))
	for _, svc := range svcsMap {
		list = append(list, svc)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"items": list,
		"total": len(list),
		"mesh":  meshName,
	})
}

// HandleListExternalServices returns external non-mesh dependencies.
func (m *MultiMeshManager) HandleListExternalServices(w http.ResponseWriter, r *http.Request) {
	meshName := chi.URLParam(r, "mesh")
	m.mu.RLock()
	defer m.mu.RUnlock()

	extMap, ok := m.externalServices[meshName]
	if !ok {
		extMap = m.externalServices["default"]
	}

	list := make([]*MeshExternalService, 0, len(extMap))
	for _, ext := range extMap {
		list = append(list, ext)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"items": list,
		"total": len(list),
		"mesh":  meshName,
	})
}

// HandleListTrafficPermissions returns Zero-Trust permissions for a mesh.
func (m *MultiMeshManager) HandleListTrafficPermissions(w http.ResponseWriter, r *http.Request) {
	meshName := chi.URLParam(r, "mesh")
	m.mu.RLock()
	defer m.mu.RUnlock()

	permMap, ok := m.trafficPermissions[meshName]
	if !ok {
		permMap = m.trafficPermissions["default"]
	}

	list := make([]*MeshTrafficPermission, 0, len(permMap))
	for _, perm := range permMap {
		list = append(list, perm)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"items": list,
		"total": len(list),
		"mesh":  meshName,
	})
}

// HandleListPassthroughPolicies returns egress passthrough policies.
func (m *MultiMeshManager) HandleListPassthroughPolicies(w http.ResponseWriter, r *http.Request) {
	meshName := chi.URLParam(r, "mesh")
	m.mu.RLock()
	defer m.mu.RUnlock()

	passMap, ok := m.passthroughPolicies[meshName]
	if !ok {
		passMap = m.passthroughPolicies["default"]
	}

	list := make([]*MeshPassthroughPolicy, 0, len(passMap))
	for _, pass := range passMap {
		list = append(list, pass)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"items": list,
		"total": len(list),
		"mesh":  meshName,
	})
}

// HandleListProxyPatches returns proxy customization patches.
func (m *MultiMeshManager) HandleListProxyPatches(w http.ResponseWriter, r *http.Request) {
	meshName := chi.URLParam(r, "mesh")
	m.mu.RLock()
	defer m.mu.RUnlock()

	patchMap, ok := m.proxyPatches[meshName]
	if !ok {
		patchMap = m.proxyPatches["default"]
	}

	list := make([]*MeshProxyPatch, 0, len(patchMap))
	for _, patch := range patchMap {
		list = append(list, patch)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"items": list,
		"total": len(list),
		"mesh":  meshName,
	})
}
