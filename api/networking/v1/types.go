// Package v1 contains API types for the networking.k8s.io API group.
package v1

import (
	"github.com/vikukumar/tarak/api/meta"
	corev1 "github.com/vikukumar/tarak/api/core/v1"
)

// ─── NetworkPolicy ───────────────────────────────────────────────────────────

// NetworkPolicy describes what network traffic is allowed for a set of Pods.
type NetworkPolicy struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            NetworkPolicySpec   `json:"spec,omitempty"`
	Status          NetworkPolicyStatus `json:"status,omitempty"`
}

// NetworkPolicySpec provides the specification of a NetworkPolicy.
type NetworkPolicySpec struct {
	// Selects the pods to which this NetworkPolicy object applies.
	PodSelector meta.LabelSelector `json:"podSelector"`
	// Ingress is a list of ingress rules to be applied to the selected pods.
	Ingress []NetworkPolicyIngressRule `json:"ingress,omitempty"`
	// Egress is a list of egress rules to be applied to the selected pods.
	Egress []NetworkPolicyEgressRule `json:"egress,omitempty"`
	// PolicyTypes indicates whether the NetworkPolicy defines Ingress, Egress, or both.
	PolicyTypes []PolicyType `json:"policyTypes,omitempty"`
}

type PolicyType string

const (
	PolicyTypeIngress PolicyType = "Ingress"
	PolicyTypeEgress  PolicyType = "Egress"
)

// NetworkPolicyIngressRule describes a particular set of traffic that is allowed to the pods
// matched by a NetworkPolicySpec's podSelector.
type NetworkPolicyIngressRule struct {
	Ports []NetworkPolicyPort `json:"ports,omitempty"`
	From  []NetworkPolicyPeer `json:"from,omitempty"`
}

// NetworkPolicyEgressRule describes a particular set of traffic that is allowed out of pods
// matched by a NetworkPolicySpec's podSelector.
type NetworkPolicyEgressRule struct {
	Ports []NetworkPolicyPort `json:"ports,omitempty"`
	To    []NetworkPolicyPeer `json:"to,omitempty"`
}

// NetworkPolicyPort describes a port to allow traffic on.
type NetworkPolicyPort struct {
	Protocol *corev1.Protocol    `json:"protocol,omitempty"`
	Port     *corev1.IntOrString `json:"port,omitempty"`
	EndPort  *int32              `json:"endPort,omitempty"`
}

// NetworkPolicyPeer describes a peer to allow traffic from/to.
type NetworkPolicyPeer struct {
	PodSelector       *meta.LabelSelector `json:"podSelector,omitempty"`
	NamespaceSelector *meta.LabelSelector `json:"namespaceSelector,omitempty"`
	IPBlock           *IPBlock            `json:"ipBlock,omitempty"`
}

// IPBlock describes a particular CIDR (Ex. "192.168.1.0/24","2001:db8::/64") that is allowed
// to the pods matched by a NetworkPolicySpec's podSelector.
type IPBlock struct {
	CIDR   string   `json:"cidr"`
	Except []string `json:"except,omitempty"`
}

// NetworkPolicyStatus describe the current state of the NetworkPolicy.
type NetworkPolicyStatus struct {
	Conditions []meta.Condition `json:"conditions,omitempty"`
}

// NetworkPolicyList is a list of NetworkPolicy objects.
type NetworkPolicyList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []NetworkPolicy `json:"items"`
}

// ─── Ingress ─────────────────────────────────────────────────────────────────

// Ingress is a collection of rules that allow inbound connections to reach the endpoints
// defined by a backend.
type Ingress struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            IngressSpec   `json:"spec,omitempty"`
	Status          IngressStatus `json:"status,omitempty"`
}

// IngressSpec describes the Ingress the user wishes to exist.
type IngressSpec struct {
	// IngressClassName is the name of an IngressClass cluster resource.
	IngressClassName *string `json:"ingressClassName,omitempty"`
	// DefaultBackend is the backend that should handle requests that don't match any rule.
	DefaultBackend *IngressBackend `json:"defaultBackend,omitempty"`
	// TLS configuration. Currently the Ingress only supports a single TLS port, 443.
	TLS []IngressTLS `json:"tls,omitempty"`
	// Rules is a list of host rules used to configure the Ingress.
	Rules []IngressRule `json:"rules,omitempty"`
}

// IngressTLS describes the transport layer security associated with an Ingress.
type IngressTLS struct {
	// Hosts are a list of hosts included in the TLS certificate.
	Hosts []string `json:"hosts,omitempty"`
	// SecretName is the name of the secret used to terminate TLS traffic.
	SecretName string `json:"secretName,omitempty"`
}

// IngressRule represents the rules mapping the paths under a specified host to the related backend services.
type IngressRule struct {
	// Host is the fully qualified domain name of a network host.
	Host string `json:"host,omitempty"`
	// IngressRuleValue represents a rule to route requests for this IngressRule.
	IngressRuleValue `json:",inline"`
}

// IngressRuleValue represents a rule to apply against incoming requests.
type IngressRuleValue struct {
	HTTP *HTTPIngressRuleValue `json:"http,omitempty"`
}

// HTTPIngressRuleValue is a list of http selectors pointing to backends.
type HTTPIngressRuleValue struct {
	Paths []HTTPIngressPath `json:"paths"`
}

// HTTPIngressPath associates a path with a backend.
type HTTPIngressPath struct {
	// Path is matched against the path of an incoming request.
	Path string `json:"path,omitempty"`
	// PathType determines the interpretation of the Path matching.
	PathType *PathType `json:"pathType"`
	// Backend defines the referenced service endpoint to which the traffic will be forwarded to.
	Backend IngressBackend `json:"backend"`
}

type PathType string

const (
	PathTypeExact                  PathType = "Exact"
	PathTypePrefix                 PathType = "Prefix"
	PathTypeImplementationSpecific PathType = "ImplementationSpecific"
)

// IngressBackend describes all endpoints for a given service and port.
type IngressBackend struct {
	// Service references a Service as a Backend.
	Service *IngressServiceBackend `json:"service,omitempty"`
	// Resource is an ObjectRef to another Kubernetes resource in the namespace of the Ingress object.
	Resource *corev1.TypedLocalObjectReference `json:"resource,omitempty"`
}

// IngressServiceBackend references a Kubernetes Service as a Backend.
type IngressServiceBackend struct {
	// Name is the referenced service.
	Name string `json:"name"`
	// Port of the referenced service.
	Port ServiceBackendPort `json:"port"`
}

// ServiceBackendPort is the service port being referenced.
type ServiceBackendPort struct {
	// Name is the name of the port on the Service.
	Name string `json:"name,omitempty"`
	// Number is the numerical port number (e.g. 80) on the Service.
	Number int32 `json:"number,omitempty"`
}

// IngressStatus describe the current state of the Ingress.
type IngressStatus struct {
	// LoadBalancer contains the current status of the load-balancer.
	LoadBalancer IngressLoadBalancerStatus `json:"loadBalancer,omitempty"`
}

// IngressLoadBalancerStatus represents the status of a load-balancer.
type IngressLoadBalancerStatus struct {
	// Ingress is a list containing ingress points for the load-balancer.
	Ingress []IngressLoadBalancerIngress `json:"ingress,omitempty"`
}

// IngressLoadBalancerIngress represents the status of a load-balancer ingress point.
type IngressLoadBalancerIngress struct {
	IP       string              `json:"ip,omitempty"`
	Hostname string              `json:"hostname,omitempty"`
	Ports    []IngressPortStatus `json:"ports,omitempty"`
}

// IngressPortStatus represents the error condition of a service port.
type IngressPortStatus struct {
	Port     int32           `json:"port"`
	Protocol corev1.Protocol `json:"protocol"`
	Error    *string         `json:"error,omitempty"`
}

// IngressList is a collection of Ingress.
type IngressList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []Ingress `json:"items"`
}

// ─── IngressClass ────────────────────────────────────────────────────────────

// IngressClass represents the class of the Ingress.
type IngressClass struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            IngressClassSpec `json:"spec,omitempty"`
}

// IngressClassSpec provides information about the class of an Ingress.
type IngressClassSpec struct {
	Controller string                           `json:"controller,omitempty"`
	Parameters *IngressClassParametersReference `json:"parameters,omitempty"`
}

// IngressClassParametersReference identifies an API object.
type IngressClassParametersReference struct {
	APIGroup  *string `json:"apiGroup,omitempty"`
	Kind      string  `json:"kind"`
	Name      string  `json:"name"`
	Scope     *string `json:"scope,omitempty"`
	Namespace *string `json:"namespace,omitempty"`
}

// IngressClassList is a collection of IngressClasses.
type IngressClassList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []IngressClass `json:"items"`
}
