package mesh

import "time"

// Mesh defines a multi-tenant service mesh boundary.
type Mesh struct {
	Name        string     `json:"name"`
	MTLS        MTLSConfig `json:"mtls"`
	Passthrough string     `json:"passthrough"` // "Passthrough" or "DenyAll"
	Metrics     string     `json:"metrics"`     // "Prometheus"
	Tracing     string     `json:"tracing"`     // "OpenTelemetry"
	Logging     string     `json:"logging"`     // "StructuredJSON"
	CreatedAt   time.Time  `json:"createdAt"`
}

// MTLSConfig configures mutual TLS and Zero-Trust identity for a mesh.
type MTLSConfig struct {
	Enabled     bool   `json:"enabled"`
	Mode        string `json:"mode"` // "Strict" or "Permissive"
	TrustDomain string `json:"trustDomain"`
	Backend     string `json:"backend"` // "builtin" or "provided"
}

// MeshService represents a discovered internal service in the mesh.
type MeshService struct {
	Name        string   `json:"name"`
	Mesh        string   `json:"mesh"`
	Namespace   string   `json:"namespace"`
	Hostnames   []string `json:"hostnames"` // e.g. ["order-service.default.mesh", "order-service.mesh"]
	VirtualVIP  string   `json:"virtualVIP"`
	Port        int      `json:"port"`
	Protocol    string   `json:"protocol"` // "http", "grpc", "tcp"
	Endpoints   []string `json:"endpoints"`
	Status      string   `json:"status"` // "Healthy", "Degraded"
	SPIFFEID    string   `json:"spiffeId"`
}

// MeshExternalService represents an external non-mesh dependency (e.g. Stripe, AWS RDS).
type MeshExternalService struct {
	Name        string `json:"name"`
	Mesh        string `json:"mesh"`
	Host        string `json:"host"` // e.g. "api.stripe.com"
	Port        int    `json:"port"`
	TLSRequired bool   `json:"tlsRequired"`
	SNI         string `json:"sni,omitempty"`
}

// MeshTrafficPermission defines Zero-Trust authorization between services in a mesh.
type MeshTrafficPermission struct {
	Name        string            `json:"name"`
	Mesh        string            `json:"mesh"`
	From        []PermissionMatch `json:"from"`
	To          []PermissionMatch `json:"to"`
	Action      string            `json:"action"` // "ALLOW" or "DENY"
}

// PermissionMatch selector for traffic permission rules.
type PermissionMatch struct {
	Service string            `json:"service"`
	Tags    map[string]string `json:"tags,omitempty"`
}

// MeshPassthroughPolicy configures outbound egress routing to external networks.
type MeshPassthroughPolicy struct {
	Name        string   `json:"name"`
	Mesh        string   `json:"mesh"`
	AllowedCIDRs []string `json:"allowedCIDRs"` // e.g. ["0.0.0.0/0", "192.168.1.0/24"]
	AllowedHosts []string `json:"allowedHosts"` // e.g. ["*.github.com", "*.amazonaws.com"]
}

// MeshProxyPatch allows fine-tuning the underlying proxy behavior.
type MeshProxyPatch struct {
	Name         string            `json:"name"`
	Mesh         string            `json:"mesh"`
	Target       string            `json:"target"` // e.g. "all" or specific service name
	ConnectTimeoutMs int           `json:"connectTimeoutMs,omitempty"`
	IdleTimeoutMs    int           `json:"idleTimeoutMs,omitempty"`
	HTTP2Enabled     bool          `json:"http2Enabled"`
	CustomHeaders    map[string]string `json:"customHeaders,omitempty"`
}
