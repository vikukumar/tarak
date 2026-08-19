// Package v1 contains API types for Tarak-native custom resources and security extensions.
package v1

import (
	"github.com/vikukumar/tarak/api/meta"
)

// ─── TarakSecurityPolicy (security.tarak.io/v1) ────────────────────────────────

// TarakSecurityPolicy is a cluster-level or namespaced security enforcement policy
// that provides zero-trust container isolation, kernel hardening, secret encryption,
// and automated network containment for workloads.
type TarakSecurityPolicy struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   TarakSecurityPolicySpec   `json:"spec,omitempty"`
	Status TarakSecurityPolicyStatus `json:"status,omitempty"`
}

// TarakSecurityPolicySpec defines the desired security configuration.
type TarakSecurityPolicySpec struct {
	// Privileged controls whether privileged containers are permitted. Default: false.
	Privileged bool `json:"privileged"`

	// AllowPrivilegeEscalation controls whether a process can set the no_new_privs flag. Default: false.
	AllowPrivilegeEscalation bool `json:"allowPrivilegeEscalation"`

	// ReadOnlyRootFilesystem requires containers to run with a read-only root filesystem. Default: true.
	ReadOnlyRootFilesystem bool `json:"readOnlyRootFilesystem"`

	// RunAsNonRoot requires the container to run with a non-zero UID. Default: true.
	RunAsNonRoot bool `json:"runAsNonRoot"`

	// RunAsUser is the default UID to enforce if not specified in the pod.
	RunAsUser *int64 `json:"runAsUser,omitempty"`

	// AllowedCapabilities is the list of Linux capabilities that may be added to a container.
	AllowedCapabilities []string `json:"allowedCapabilities,omitempty"`

	// RequiredDropCapabilities is the list of Linux capabilities that must be dropped.
	RequiredDropCapabilities []string `json:"requiredDropCapabilities,omitempty"`

	// EnforceEncryptionAtRest requires all attached Secrets and ConfigMaps to use hardware-backed AES-256-GCM.
	EnforceEncryptionAtRest bool `json:"enforceEncryptionAtRest"`

	// NetworkIsolation enforces strict default-deny egress and ingress for matching workloads.
	NetworkIsolation bool `json:"networkIsolation"`

	// AllowedEgressCIDRs specifies permitted outbound IP ranges when NetworkIsolation is active.
	AllowedEgressCIDRs []string `json:"allowedEgressCIDRs,omitempty"`

	// TargetSelector determines which namespaces or pods this policy applies to.
	TargetSelector *meta.LabelSelector `json:"targetSelector,omitempty"`
}

// TarakSecurityPolicyStatus represents the observed enforcement status.
type TarakSecurityPolicyStatus struct {
	// Enforced indicates whether the security policy is actively enforced across all nodes.
	Enforced bool `json:"enforced"`

	// ActiveViolations is the number of active pods that violate this security policy.
	ActiveViolations int `json:"activeViolations"`

	// Conditions represents the latest available observations of the policy's state.
	Conditions []meta.Condition `json:"conditions,omitempty"`
}

// TarakSecurityPolicyList is a list of TarakSecurityPolicy objects.
type TarakSecurityPolicyList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []TarakSecurityPolicy `json:"items"`
}

// ─── TarakApplication (apps.tarak.io/v1) ───────────────────────────────────────

// TarakApplication is a declarative, all-in-one application specification that
// encapsulates container workload, auto-TLS ingress, service routing, scaling,
// and security policy binding into a single native resource.
type TarakApplication struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   TarakApplicationSpec   `json:"spec,omitempty"`
	Status TarakApplicationStatus `json:"status,omitempty"`
}

// TarakApplicationSpec defines the desired state of a Tarak application.
type TarakApplicationSpec struct {
	// Image is the container image repository and tag (e.g. "ghcr.io/org/app:v1.0").
	Image string `json:"image"`

	// Replicas is the desired number of running instances. Default: 1.
	Replicas int32 `json:"replicas"`

	// Port is the primary container port exposed by the application.
	Port int `json:"port"`

	// Domain is the public or internal hostname for automatic ingress and TLS (e.g. "api.mycluster.io").
	Domain string `json:"domain,omitempty"`

	// AutoTLS automatically provisions and rotates Let's Encrypt or Tarak Root CA TLS certificates.
	AutoTLS bool `json:"autoTLS"`

	// Env is a map of environment variables injected into the container.
	Env map[string]string `json:"env,omitempty"`

	// SecurityPolicyRef is the name of a TarakSecurityPolicy to enforce on this application.
	SecurityPolicyRef string `json:"securityPolicyRef,omitempty"`

	// Storage specifies persistent volume claim settings if persistence is required.
	Storage *TarakAppStorage `json:"storage,omitempty"`
}

// TarakAppStorage specifies storage configuration for an application.
type TarakAppStorage struct {
	MountPath string `json:"mountPath"`
	Size      string `json:"size"`
	ClassName string `json:"className,omitempty"`
}

// TarakApplicationStatus defines the observed runtime state of a TarakApplication.
type TarakApplicationStatus struct {
	// Phase is the current application lifecycle stage: Pending, Running, Updating, Degraded.
	Phase string `json:"phase"`

	// ReadyReplicas is the number of replicas currently healthy and serving traffic.
	ReadyReplicas int32 `json:"readyReplicas"`

	// Endpoint is the accessible URL for the application (e.g. "https://api.mycluster.io").
	Endpoint string `json:"endpoint,omitempty"`

	// InternalClusterIP is the assigned service cluster IP within the overlay network.
	InternalClusterIP string `json:"internalClusterIP,omitempty"`

	// Conditions describe the detailed status history.
	Conditions []meta.Condition `json:"conditions,omitempty"`
}

// TarakApplicationList is a list of TarakApplication objects.
type TarakApplicationList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []TarakApplication `json:"items"`
}

// ─── CustomResourceDefinition (apiextensions.k8s.io/v1 & apiextensions.tarak.io/v1) ─

// CustomResourceDefinition represents a dynamic user-defined resource schema in the cluster.
type CustomResourceDefinition struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomResourceDefinitionSpec   `json:"spec"`
	Status CustomResourceDefinitionStatus `json:"status,omitempty"`
}

// CustomResourceDefinitionSpec describes how a user resource is defined.
type CustomResourceDefinitionSpec struct {
	Group string                          `json:"group"`
	Names CustomResourceDefinitionNames   `json:"names"`
	Scope string                          `json:"scope"` // "Namespaced" or "Cluster"
	Versions []CustomResourceDefinitionVersion `json:"versions"`
}

// CustomResourceDefinitionNames indicates the names to serve this resource.
type CustomResourceDefinitionNames struct {
	Plural     string   `json:"plural"`
	Singular   string   `json:"singular,omitempty"`
	ShortNames []string `json:"shortNames,omitempty"`
	Kind       string   `json:"kind"`
	ListKind   string   `json:"listKind,omitempty"`
}

// CustomResourceDefinitionVersion describes a version for CRD.
type CustomResourceDefinitionVersion struct {
	Name    string `json:"name"`
	Served  bool   `json:"served"`
	Storage bool   `json:"storage"`
}

// CustomResourceDefinitionStatus indicates the state of the CRD.
type CustomResourceDefinitionStatus struct {
	Conditions []meta.Condition `json:"conditions,omitempty"`
}

// CustomResourceDefinitionList is a list of CustomResourceDefinitions.
type CustomResourceDefinitionList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []CustomResourceDefinition `json:"items"`
}
