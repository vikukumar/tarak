// Package v1 contains API types for the core API group.
// These types are structurally compatible with the Kubernetes core/v1 API so that
// existing manifests can be applied to a Tarak cluster without modification.
package v1

import (
	"github.com/vikukumar/tarak/api/meta"
)

// ─── Namespace ───────────────────────────────────────────────────────────────

// Namespace provides a scope for Names.
type Namespace struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            NamespaceSpec   `json:"spec,omitempty"`
	Status          NamespaceStatus `json:"status,omitempty"`
}

// NamespaceSpec describes the attributes on a Namespace.
type NamespaceSpec struct {
	// Finalizers is an opaque list of values that must be empty to permanently remove object from storage.
	Finalizers []FinalizerName `json:"finalizers,omitempty"`
}

// NamespaceStatus is information about the current status of a Namespace.
type NamespaceStatus struct {
	// Phase is the current lifecycle phase of the namespace.
	Phase NamespacePhase `json:"phase,omitempty"`
	// Conditions is an array of current conditions.
	Conditions []meta.Condition `json:"conditions,omitempty"`
}

type NamespacePhase string

const (
	NamespaceActive      NamespacePhase = "Active"
	NamespaceTerminating NamespacePhase = "Terminating"
)

type FinalizerName string

const (
	FinalizerKubernetes FinalizerName = "kubernetes"
	FinalizerTarak      FinalizerName = "tarak"
)

// NamespaceList is a list of Namespaces.
type NamespaceList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []Namespace `json:"items"`
}

// ─── Node ────────────────────────────────────────────────────────────────────

// Node is a worker node in the cluster.
type Node struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            NodeSpec   `json:"spec,omitempty"`
	Status          NodeStatus `json:"status,omitempty"`
}

// NodeSpec describes the attributes that a node is created with.
type NodeSpec struct {
	// PodCIDR represents the pod IP range assigned to the node.
	PodCIDR string `json:"podCIDR,omitempty"`
	// PodCIDRs represents the IP ranges assigned to the node for usage by Pods.
	PodCIDRs []string `json:"podCIDRs,omitempty"`
	// ProviderID is the unique identifier of a node in the cloud provider format.
	ProviderID string `json:"providerID,omitempty"`
	// Unschedulable controls node schedulability of new pods.
	Unschedulable bool `json:"unschedulable,omitempty"`
	// Taints represents the taints attached to the node.
	Taints []Taint `json:"taints,omitempty"`
	// ConfigSource specifies the source to get node configuration from.
	ConfigSource *NodeConfigSource `json:"configSource,omitempty"`
}

// NodeConfigSource specifies a source of node configuration.
type NodeConfigSource struct {
	ConfigMap *ConfigMapNodeConfigSource `json:"configMap,omitempty"`
}

// ConfigMapNodeConfigSource contains the information to reference a ConfigMap as a config source.
type ConfigMapNodeConfigSource struct {
	Namespace        string `json:"namespace"`
	Name             string `json:"name"`
	UID              string `json:"uid,omitempty"`
	ResourceVersion  string `json:"resourceVersion,omitempty"`
	KubeletConfigKey string `json:"kubeletConfigKey"`
}

// Taint represents a taint attached to a node.
type Taint struct {
	Key    string      `json:"key"`
	Value  string      `json:"value,omitempty"`
	Effect TaintEffect `json:"effect"`
	// TimeAdded records when the taint was added. Only written for NoExecute taints.
	TimeAdded *meta.Time `json:"timeAdded,omitempty"`
}

// TaintEffect defines the effect of a taint.
type TaintEffect string

const (
	TaintEffectNoSchedule       TaintEffect = "NoSchedule"
	TaintEffectPreferNoSchedule TaintEffect = "PreferNoSchedule"
	TaintEffectNoExecute        TaintEffect = "NoExecute"
)

// NodeStatus is information about the current status of a node.
type NodeStatus struct {
	// Capacity represents the total resources of a node.
	Capacity ResourceList `json:"capacity,omitempty"`
	// Allocatable represents the resources of a node that are available for scheduling.
	Allocatable ResourceList `json:"allocatable,omitempty"`
	// Phase is deprecated.
	Phase NodePhase `json:"phase,omitempty"`
	// Conditions is an array of current observed node conditions.
	Conditions []NodeCondition `json:"conditions,omitempty"`
	// Addresses is the list of addresses reachable to the node.
	Addresses []NodeAddress `json:"addresses,omitempty"`
	// DaemonEndpoints provides the details of the ports opened by daemons running on the Node.
	DaemonEndpoints NodeDaemonEndpoints `json:"daemonEndpoints,omitempty"`
	// NodeInfo is the set of ids/uuids to uniquely identify the node.
	NodeInfo NodeSystemInfo `json:"nodeInfo,omitempty"`
	// Images is the list of container images on this node.
	Images []ContainerImage `json:"images,omitempty"`
	// VolumesInUse is the list of unique volumes currently mounted by any pod on the node.
	VolumesInUse []UniqueVolumeName `json:"volumesInUse,omitempty"`
	// VolumesAttached is the list of volumes that are attached to the current node.
	VolumesAttached []AttachedVolume `json:"volumesAttached,omitempty"`
	// Config is the corresponding effective node configuration.
	Config *NodeConfigStatus `json:"config,omitempty"`
}

type NodePhase string

const (
	NodePending    NodePhase = "Pending"
	NodeRunning    NodePhase = "Running"
	NodeTerminated NodePhase = "Terminated"
)

// NodeCondition contains condition information for a node.
type NodeCondition struct {
	Type               NodeConditionType    `json:"type"`
	Status             meta.ConditionStatus `json:"status"`
	LastHeartbeatTime  meta.Time            `json:"lastHeartbeatTime,omitempty"`
	LastTransitionTime meta.Time            `json:"lastTransitionTime,omitempty"`
	Reason             string               `json:"reason,omitempty"`
	Message            string               `json:"message,omitempty"`
}

type NodeConditionType string

const (
	NodeReady              NodeConditionType = "Ready"
	NodeMemoryPressure     NodeConditionType = "MemoryPressure"
	NodeDiskPressure       NodeConditionType = "DiskPressure"
	NodePIDPressure        NodeConditionType = "PIDPressure"
	NodeNetworkUnavailable NodeConditionType = "NetworkUnavailable"
)

// NodeAddress contains information for the node's address.
type NodeAddress struct {
	Type    NodeAddressType `json:"type"`
	Address string          `json:"address"`
}

type NodeAddressType string

const (
	NodeHostName    NodeAddressType = "Hostname"
	NodeExternalIP  NodeAddressType = "ExternalIP"
	NodeInternalIP  NodeAddressType = "InternalIP"
	NodeExternalDNS NodeAddressType = "ExternalDNS"
	NodeInternalDNS NodeAddressType = "InternalDNS"
)

// NodeDaemonEndpoints lists ports opened by daemons running on the Node.
type NodeDaemonEndpoints struct {
	KubeletEndpoint DaemonEndpoint `json:"kubeletEndpoint,omitempty"`
}

// DaemonEndpoint contains information about a single Daemon endpoint.
type DaemonEndpoint struct {
	Port int32 `json:"Port"`
}

// NodeSystemInfo is a set of ids/uuids to uniquely identify the node.
type NodeSystemInfo struct {
	MachineID               string `json:"machineID"`
	SystemUUID              string `json:"systemUUID"`
	BootID                  string `json:"bootID"`
	KernelVersion           string `json:"kernelVersion"`
	OSImage                 string `json:"osImage"`
	ContainerRuntimeVersion string `json:"containerRuntimeVersion"`
	KubeletVersion          string `json:"kubeletVersion"`
	KubeProxyVersion        string `json:"kubeProxyVersion"`
	OperatingSystem         string `json:"operatingSystem"`
	Architecture            string `json:"architecture"`
}

// ContainerImage describes a container image present on a machine.
type ContainerImage struct {
	Names     []string `json:"names,omitempty"`
	SizeBytes int64    `json:"sizeBytes,omitempty"`
}

// UniqueVolumeName defines the name of attached volume.
type UniqueVolumeName string

// AttachedVolume describes a volume attached to a node.
type AttachedVolume struct {
	Name       UniqueVolumeName `json:"name"`
	DevicePath string           `json:"devicePath"`
}

// NodeConfigStatus describes the status of the config assigned by Node.Spec.ConfigSource.
type NodeConfigStatus struct {
	Assigned      *NodeConfigSource `json:"assigned,omitempty"`
	Active        *NodeConfigSource `json:"active,omitempty"`
	LastKnownGood *NodeConfigSource `json:"lastKnownGood,omitempty"`
	Error         string            `json:"error,omitempty"`
}

// NodeList is a list of Nodes.
type NodeList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []Node `json:"items"`
}

// ─── Pod ─────────────────────────────────────────────────────────────────────

// Pod is a collection of containers that can run on a host.
type Pod struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            PodSpec   `json:"spec,omitempty"`
	Status          PodStatus `json:"status,omitempty"`
}

// PodSpec describes the specification of the desired behaviour of the pod.
type PodSpec struct {
	Volumes                       []Volume                   `json:"volumes,omitempty"`
	InitContainers                []Container                `json:"initContainers,omitempty"`
	Containers                    []Container                `json:"containers"`
	EphemeralContainers           []EphemeralContainer       `json:"ephemeralContainers,omitempty"`
	RestartPolicy                 RestartPolicy              `json:"restartPolicy,omitempty"`
	TerminationGracePeriodSeconds *int64                     `json:"terminationGracePeriodSeconds,omitempty"`
	ActiveDeadlineSeconds         *int64                     `json:"activeDeadlineSeconds,omitempty"`
	DNSPolicy                     DNSPolicy                  `json:"dnsPolicy,omitempty"`
	NodeSelector                  map[string]string          `json:"nodeSelector,omitempty"`
	ServiceAccountName            string                     `json:"serviceAccountName,omitempty"`
	AutomountServiceAccountToken  *bool                      `json:"automountServiceAccountToken,omitempty"`
	NodeName                      string                     `json:"nodeName,omitempty"`
	HostNetwork                   bool                       `json:"hostNetwork,omitempty"`
	HostPID                       bool                       `json:"hostPID,omitempty"`
	HostIPC                       bool                       `json:"hostIPC,omitempty"`
	ShareProcessNamespace         *bool                      `json:"shareProcessNamespace,omitempty"`
	SecurityContext               *PodSecurityContext        `json:"securityContext,omitempty"`
	ImagePullSecrets              []LocalObjectReference     `json:"imagePullSecrets,omitempty"`
	Hostname                      string                     `json:"hostname,omitempty"`
	Subdomain                     string                     `json:"subdomain,omitempty"`
	Affinity                      *Affinity                  `json:"affinity,omitempty"`
	SchedulerName                 string                     `json:"schedulerName,omitempty"`
	Tolerations                   []Toleration               `json:"tolerations,omitempty"`
	HostAliases                   []HostAlias                `json:"hostAliases,omitempty"`
	PriorityClassName             string                     `json:"priorityClassName,omitempty"`
	Priority                      *int32                     `json:"priority,omitempty"`
	DNSConfig                     *PodDNSConfig              `json:"dnsConfig,omitempty"`
	ReadinessGates                []PodReadinessGate         `json:"readinessGates,omitempty"`
	RuntimeClassName              *string                    `json:"runtimeClassName,omitempty"`
	EnableServiceLinks            *bool                      `json:"enableServiceLinks,omitempty"`
	PreemptionPolicy              *PreemptionPolicy          `json:"preemptionPolicy,omitempty"`
	Overhead                      ResourceList               `json:"overhead,omitempty"`
	TopologySpreadConstraints     []TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	SetHostnameAsFQDN             *bool                      `json:"setHostnameAsFQDN,omitempty"`
	OS                            *PodOS                     `json:"os,omitempty"`
	HostUsers                     *bool                      `json:"hostUsers,omitempty"`
	SchedulingGates               []PodSchedulingGate        `json:"schedulingGates,omitempty"`
	ResourceClaims                []PodResourceClaim         `json:"resourceClaims,omitempty"`
}

type RestartPolicy string

const (
	RestartPolicyAlways    RestartPolicy = "Always"
	RestartPolicyOnFailure RestartPolicy = "OnFailure"
	RestartPolicyNever     RestartPolicy = "Never"
)

type DNSPolicy string

const (
	DNSClusterFirstWithHostNet DNSPolicy = "ClusterFirstWithHostNet"
	DNSClusterFirst            DNSPolicy = "ClusterFirst"
	DNSDefault                 DNSPolicy = "Default"
	DNSNone                    DNSPolicy = "None"
)

type PreemptionPolicy string

const (
	PreemptNever         PreemptionPolicy = "Never"
	PreemptLowerPriority PreemptionPolicy = "PreemptLowerPriority"
)

// PodOS defines the OS parameters for a pod.
type PodOS struct {
	Name OSName `json:"name"`
}

type OSName string

const (
	Linux   OSName = "linux"
	Windows OSName = "windows"
)

// PodSchedulingGate is associated to a Pod to guard its scheduling.
type PodSchedulingGate struct {
	Name string `json:"name"`
}

// PodResourceClaim references which ResourceClaim to use for a given volume.
type PodResourceClaim struct {
	Name   string      `json:"name"`
	Source ClaimSource `json:"source,omitempty"`
}

// ClaimSource describes a reference to a ResourceClaim.
type ClaimSource struct {
	ResourceClaimName         *string `json:"resourceClaimName,omitempty"`
	ResourceClaimTemplateName *string `json:"resourceClaimTemplateName,omitempty"`
}

// Container represents a container in a pod.
type Container struct {
	Name                     string                   `json:"name"`
	Image                    string                   `json:"image,omitempty"`
	Command                  []string                 `json:"command,omitempty"`
	Args                     []string                 `json:"args,omitempty"`
	WorkingDir               string                   `json:"workingDir,omitempty"`
	Ports                    []ContainerPort          `json:"ports,omitempty"`
	EnvFrom                  []EnvFromSource          `json:"envFrom,omitempty"`
	Env                      []EnvVar                 `json:"env,omitempty"`
	Resources                ResourceRequirements     `json:"resources,omitempty"`
	VolumeMounts             []VolumeMount            `json:"volumeMounts,omitempty"`
	VolumeDevices            []VolumeDevice           `json:"volumeDevices,omitempty"`
	LivenessProbe            *Probe                   `json:"livenessProbe,omitempty"`
	ReadinessProbe           *Probe                   `json:"readinessProbe,omitempty"`
	StartupProbe             *Probe                   `json:"startupProbe,omitempty"`
	Lifecycle                *Lifecycle               `json:"lifecycle,omitempty"`
	TerminationMessagePath   string                   `json:"terminationMessagePath,omitempty"`
	TerminationMessagePolicy TerminationMessagePolicy `json:"terminationMessagePolicy,omitempty"`
	ImagePullPolicy          PullPolicy               `json:"imagePullPolicy,omitempty"`
	SecurityContext          *SecurityContext         `json:"securityContext,omitempty"`
	Stdin                    bool                     `json:"stdin,omitempty"`
	StdinOnce                bool                     `json:"stdinOnce,omitempty"`
	TTY                      bool                     `json:"tty,omitempty"`
	ResizePolicy             []ContainerResizePolicy  `json:"resizePolicy,omitempty"`
	RestartPolicy            *ContainerRestartPolicy  `json:"restartPolicy,omitempty"`
}

type TerminationMessagePolicy string

const (
	TerminationMessageReadFile              TerminationMessagePolicy = "File"
	TerminationMessageFallbackToLogsOnError TerminationMessagePolicy = "FallbackToLogsOnError"
)

type PullPolicy string

const (
	PullAlways       PullPolicy = "Always"
	PullNever        PullPolicy = "Never"
	PullIfNotPresent PullPolicy = "IfNotPresent"
)

// ContainerResizePolicy represents a list of resource resize policies.
type ContainerResizePolicy struct {
	ResourceName  ResourceName        `json:"resourceName"`
	RestartPolicy ResizeRestartPolicy `json:"restartPolicy"`
}

type ResizeRestartPolicy string

const (
	NotRequired      ResizeRestartPolicy = "NotRequired"
	RestartContainer ResizeRestartPolicy = "RestartContainer"
)

type ContainerRestartPolicy string

const (
	ContainerRestartPolicyAlways ContainerRestartPolicy = "Always"
)

// EphemeralContainer is a temporary container that may be added to an existing pod
// for user-initiated activities.
type EphemeralContainer struct {
	EphemeralContainerCommon `json:",inline"`
	TargetContainerName      string `json:"targetContainerName,omitempty"`
}

// EphemeralContainerCommon is a copy of all fields in Container to be inlined
// in EphemeralContainer.
type EphemeralContainerCommon struct {
	Name                     string                   `json:"name"`
	Image                    string                   `json:"image,omitempty"`
	Command                  []string                 `json:"command,omitempty"`
	Args                     []string                 `json:"args,omitempty"`
	WorkingDir               string                   `json:"workingDir,omitempty"`
	Ports                    []ContainerPort          `json:"ports,omitempty"`
	EnvFrom                  []EnvFromSource          `json:"envFrom,omitempty"`
	Env                      []EnvVar                 `json:"env,omitempty"`
	Resources                ResourceRequirements     `json:"resources,omitempty"`
	VolumeMounts             []VolumeMount            `json:"volumeMounts,omitempty"`
	VolumeDevices            []VolumeDevice           `json:"volumeDevices,omitempty"`
	LivenessProbe            *Probe                   `json:"livenessProbe,omitempty"`
	ReadinessProbe           *Probe                   `json:"readinessProbe,omitempty"`
	StartupProbe             *Probe                   `json:"startupProbe,omitempty"`
	Lifecycle                *Lifecycle               `json:"lifecycle,omitempty"`
	TerminationMessagePath   string                   `json:"terminationMessagePath,omitempty"`
	TerminationMessagePolicy TerminationMessagePolicy `json:"terminationMessagePolicy,omitempty"`
	ImagePullPolicy          PullPolicy               `json:"imagePullPolicy,omitempty"`
	SecurityContext          *SecurityContext         `json:"securityContext,omitempty"`
	Stdin                    bool                     `json:"stdin,omitempty"`
	StdinOnce                bool                     `json:"stdinOnce,omitempty"`
	TTY                      bool                     `json:"tty,omitempty"`
}

// ContainerPort represents a network port in a single container.
type ContainerPort struct {
	Name          string   `json:"name,omitempty"`
	HostPort      int32    `json:"hostPort,omitempty"`
	ContainerPort int32    `json:"containerPort"`
	Protocol      Protocol `json:"protocol,omitempty"`
	HostIP        string   `json:"hostIP,omitempty"`
}

type Protocol string

const (
	ProtocolTCP  Protocol = "TCP"
	ProtocolUDP  Protocol = "UDP"
	ProtocolSCTP Protocol = "SCTP"
)

// EnvVar represents an environment variable present in a Container.
type EnvVar struct {
	Name      string        `json:"name"`
	Value     string        `json:"value,omitempty"`
	ValueFrom *EnvVarSource `json:"valueFrom,omitempty"`
}

// EnvVarSource represents a source for the value of an EnvVar.
type EnvVarSource struct {
	FieldRef         *ObjectFieldSelector   `json:"fieldRef,omitempty"`
	ResourceFieldRef *ResourceFieldSelector `json:"resourceFieldRef,omitempty"`
	ConfigMapKeyRef  *ConfigMapKeySelector  `json:"configMapKeyRef,omitempty"`
	SecretKeyRef     *SecretKeySelector     `json:"secretKeyRef,omitempty"`
}

// ObjectFieldSelector selects an APIVersioned field of an object.
type ObjectFieldSelector struct {
	APIVersion string `json:"apiVersion,omitempty"`
	FieldPath  string `json:"fieldPath"`
}

// ResourceFieldSelector represents container resources (cpu, memory) and their output format.
type ResourceFieldSelector struct {
	ContainerName string `json:"containerName,omitempty"`
	Resource      string `json:"resource"`
	Divisor       string `json:"divisor,omitempty"`
}

// ConfigMapKeySelector selects a key of a ConfigMap.
type ConfigMapKeySelector struct {
	LocalObjectReference `json:",inline"`
	Key                  string `json:"key"`
	Optional             *bool  `json:"optional,omitempty"`
}

// SecretKeySelector selects a key of a Secret.
type SecretKeySelector struct {
	LocalObjectReference `json:",inline"`
	Key                  string `json:"key"`
	Optional             *bool  `json:"optional,omitempty"`
}

// EnvFromSource represents the source of a set of ConfigMaps or Secrets.
type EnvFromSource struct {
	Prefix       string              `json:"prefix,omitempty"`
	ConfigMapRef *ConfigMapEnvSource `json:"configMapRef,omitempty"`
	SecretRef    *SecretEnvSource    `json:"secretRef,omitempty"`
}

// ConfigMapEnvSource selects a ConfigMap to populate the environment variables with.
type ConfigMapEnvSource struct {
	LocalObjectReference `json:",inline"`
	Optional             *bool `json:"optional,omitempty"`
}

// SecretEnvSource selects a Secret to populate the environment variables with.
type SecretEnvSource struct {
	LocalObjectReference `json:",inline"`
	Optional             *bool `json:"optional,omitempty"`
}

// ResourceRequirements describes the compute resource requirements.
type ResourceRequirements struct {
	Limits   ResourceList    `json:"limits,omitempty"`
	Requests ResourceList    `json:"requests,omitempty"`
	Claims   []ResourceClaim `json:"claims,omitempty"`
}

// ResourceClaim references one entry in PodSpec.ResourceClaims.
type ResourceClaim struct {
	Name string `json:"name"`
}

// ResourceList is a set of (resource name, quantity) pairs.
type ResourceList map[ResourceName]string

// ResourceName is the name of a compute resource.
type ResourceName string

const (
	ResourceCPU              ResourceName = "cpu"
	ResourceMemory           ResourceName = "memory"
	ResourceStorage          ResourceName = "storage"
	ResourceEphemeralStorage ResourceName = "ephemeral-storage"
	ResourceHugePagesPrefix  ResourceName = "hugepages-"
)

// VolumeMount describes a mounting of a Volume within a container.
type VolumeMount struct {
	Name             string                `json:"name"`
	ReadOnly         bool                  `json:"readOnly,omitempty"`
	MountPath        string                `json:"mountPath"`
	SubPath          string                `json:"subPath,omitempty"`
	MountPropagation *MountPropagationMode `json:"mountPropagation,omitempty"`
	SubPathExpr      string                `json:"subPathExpr,omitempty"`
}

type MountPropagationMode string

const (
	MountPropagationNone            MountPropagationMode = "None"
	MountPropagationHostToContainer MountPropagationMode = "HostToContainer"
	MountPropagationBidirectional   MountPropagationMode = "Bidirectional"
)

// VolumeDevice describes a mapping of a raw block device within a container.
type VolumeDevice struct {
	Name       string `json:"name"`
	DevicePath string `json:"devicePath"`
}

// Probe describes a health check to be performed against a container.
type Probe struct {
	ProbeHandler                  `json:",inline"`
	InitialDelaySeconds           int32  `json:"initialDelaySeconds,omitempty"`
	TimeoutSeconds                int32  `json:"timeoutSeconds,omitempty"`
	PeriodSeconds                 int32  `json:"periodSeconds,omitempty"`
	SuccessThreshold              int32  `json:"successThreshold,omitempty"`
	FailureThreshold              int32  `json:"failureThreshold,omitempty"`
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`
}

// ProbeHandler defines a specific action that should be taken in a probe.
type ProbeHandler struct {
	Exec      *ExecAction      `json:"exec,omitempty"`
	HTTPGet   *HTTPGetAction   `json:"httpGet,omitempty"`
	TCPSocket *TCPSocketAction `json:"tcpSocket,omitempty"`
	GRPC      *GRPCAction      `json:"grpc,omitempty"`
}

// ExecAction describes a "run in container" action.
type ExecAction struct {
	Command []string `json:"command,omitempty"`
}

// HTTPGetAction describes an action based on HTTP Get requests.
type HTTPGetAction struct {
	Path        string       `json:"path,omitempty"`
	Port        IntOrString  `json:"port"`
	Host        string       `json:"host,omitempty"`
	Scheme      URIScheme    `json:"scheme,omitempty"`
	HTTPHeaders []HTTPHeader `json:"httpHeaders,omitempty"`
}

// HTTPHeader describes a custom header to be used in HTTP probes.
type HTTPHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type URIScheme string

const (
	URISchemeHTTP  URIScheme = "HTTP"
	URISchemeHTTPS URIScheme = "HTTPS"
)

// TCPSocketAction describes an action based on opening a socket.
type TCPSocketAction struct {
	Port IntOrString `json:"port"`
	Host string      `json:"host,omitempty"`
}

// GRPCAction describes an action involving a GRPC port.
type GRPCAction struct {
	Port    int32   `json:"port"`
	Service *string `json:"service"`
}

// IntOrString is a type that can hold an int32 or a string.
type IntOrString struct {
	Type   IntOrStringType `json:"type,omitempty"`
	IntVal int32           `json:"intVal,omitempty"`
	StrVal string          `json:"strVal,omitempty"`
}

type IntOrStringType int

const (
	Int    IntOrStringType = 0
	String IntOrStringType = 1
)

// Lifecycle describes actions that the management system should take in response
// to container lifecycle events.
type Lifecycle struct {
	PostStart *LifecycleHandler `json:"postStart,omitempty"`
	PreStop   *LifecycleHandler `json:"preStop,omitempty"`
}

// LifecycleHandler defines a specific action that should be taken in a lifecycle hook.
type LifecycleHandler struct {
	Exec      *ExecAction      `json:"exec,omitempty"`
	HTTPGet   *HTTPGetAction   `json:"httpGet,omitempty"`
	TCPSocket *TCPSocketAction `json:"tcpSocket,omitempty"`
	Sleep     *SleepAction     `json:"sleep,omitempty"`
}

// SleepAction describes a "sleep" action.
type SleepAction struct {
	Seconds int64 `json:"seconds"`
}

// PodSecurityContext holds pod-level security attributes.
type PodSecurityContext struct {
	SELinuxOptions      *SELinuxOptions                `json:"seLinuxOptions,omitempty"`
	WindowsOptions      *WindowsSecurityContextOptions `json:"windowsOptions,omitempty"`
	RunAsUser           *int64                         `json:"runAsUser,omitempty"`
	RunAsGroup          *int64                         `json:"runAsGroup,omitempty"`
	RunAsNonRoot        *bool                          `json:"runAsNonRoot,omitempty"`
	SupplementalGroups  []int64                        `json:"supplementalGroups,omitempty"`
	FSGroup             *int64                         `json:"fsGroup,omitempty"`
	Sysctls             []Sysctl                       `json:"sysctls,omitempty"`
	FSGroupChangePolicy *PodFSGroupChangePolicy        `json:"fsGroupChangePolicy,omitempty"`
	SeccompProfile      *SeccompProfile                `json:"seccompProfile,omitempty"`
}

// SecurityContext holds security configuration that will be applied to a container.
type SecurityContext struct {
	Capabilities             *Capabilities                  `json:"capabilities,omitempty"`
	Privileged               *bool                          `json:"privileged,omitempty"`
	SELinuxOptions           *SELinuxOptions                `json:"seLinuxOptions,omitempty"`
	WindowsOptions           *WindowsSecurityContextOptions `json:"windowsOptions,omitempty"`
	RunAsUser                *int64                         `json:"runAsUser,omitempty"`
	RunAsGroup               *int64                         `json:"runAsGroup,omitempty"`
	RunAsNonRoot             *bool                          `json:"runAsNonRoot,omitempty"`
	ReadOnlyRootFilesystem   *bool                          `json:"readOnlyRootFilesystem,omitempty"`
	AllowPrivilegeEscalation *bool                          `json:"allowPrivilegeEscalation,omitempty"`
	ProcMount                *ProcMountType                 `json:"procMount,omitempty"`
	SeccompProfile           *SeccompProfile                `json:"seccompProfile,omitempty"`
}

type ProcMountType string

const (
	DefaultProcMount  ProcMountType = "Default"
	UnmaskedProcMount ProcMountType = "Unmasked"
)

// Capabilities adds and removes POSIX capabilities from running containers.
type Capabilities struct {
	Add  []Capability `json:"add,omitempty"`
	Drop []Capability `json:"drop,omitempty"`
}

type Capability string

// SELinuxOptions are the labels to be applied to the container.
type SELinuxOptions struct {
	User  string `json:"user,omitempty"`
	Role  string `json:"role,omitempty"`
	Type  string `json:"type,omitempty"`
	Level string `json:"level,omitempty"`
}

// WindowsSecurityContextOptions contain Windows-specific options.
type WindowsSecurityContextOptions struct {
	GMSACredentialSpecName *string `json:"gmsaCredentialSpecName,omitempty"`
	GMSACredentialSpec     *string `json:"gmsaCredentialSpec,omitempty"`
	RunAsUserName          *string `json:"runAsUserName,omitempty"`
	HostProcess            *bool   `json:"hostProcess,omitempty"`
}

// Sysctl defines a kernel parameter to be set.
type Sysctl struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PodFSGroupChangePolicy string

const (
	FSGroupChangeOnRootMismatch PodFSGroupChangePolicy = "OnRootMismatch"
	FSGroupChangeAlways         PodFSGroupChangePolicy = "Always"
)

// SeccompProfile defines a pod/container's seccomp profile settings.
type SeccompProfile struct {
	Type             SeccompProfileType `json:"type"`
	LocalhostProfile *string            `json:"localhostProfile,omitempty"`
}

type SeccompProfileType string

const (
	SeccompProfileTypeUnconfined     SeccompProfileType = "Unconfined"
	SeccompProfileTypeRuntimeDefault SeccompProfileType = "RuntimeDefault"
	SeccompProfileTypeLocalhost      SeccompProfileType = "Localhost"
)

// Volume represents a named volume in a pod that may be accessed by any container.
type Volume struct {
	Name         string `json:"name"`
	VolumeSource `json:",inline"`
}

// VolumeSource represents the location and type of the mounted volume.
type VolumeSource struct {
	HostPath              *HostPathVolumeSource              `json:"hostPath,omitempty"`
	EmptyDir              *EmptyDirVolumeSource              `json:"emptyDir,omitempty"`
	GCEPersistentDisk     *GCEPersistentDiskVolumeSource     `json:"gcePersistentDisk,omitempty"`
	AWSElasticBlockStore  *AWSElasticBlockStoreVolumeSource  `json:"awsElasticBlockStore,omitempty"`
	ConfigMap             *ConfigMapVolumeSource             `json:"configMap,omitempty"`
	Secret                *SecretVolumeSource                `json:"secret,omitempty"`
	NFS                   *NFSVolumeSource                   `json:"nfs,omitempty"`
	PersistentVolumeClaim *PersistentVolumeClaimVolumeSource `json:"persistentVolumeClaim,omitempty"`
	Projected             *ProjectedVolumeSource             `json:"projected,omitempty"`
	CSI                   *CSIVolumeSource                   `json:"csi,omitempty"`
	Ephemeral             *EphemeralVolumeSource             `json:"ephemeral,omitempty"`
}

// HostPathVolumeSource represents a host path mounted into a Pod.
type HostPathVolumeSource struct {
	Path string        `json:"path"`
	Type *HostPathType `json:"type,omitempty"`
}

type HostPathType string

const (
	HostPathUnset             HostPathType = ""
	HostPathDirectoryOrCreate HostPathType = "DirectoryOrCreate"
	HostPathDirectory         HostPathType = "Directory"
	HostPathFileOrCreate      HostPathType = "FileOrCreate"
	HostPathFile              HostPathType = "File"
	HostPathSocket            HostPathType = "Socket"
	HostPathCharDev           HostPathType = "CharDevice"
	HostPathBlockDev          HostPathType = "BlockDevice"
)

// EmptyDirVolumeSource represents an empty directory for a pod.
type EmptyDirVolumeSource struct {
	Medium    StorageMedium `json:"medium,omitempty"`
	SizeLimit *string       `json:"sizeLimit,omitempty"`
}

type StorageMedium string

const (
	StorageMediumDefault   StorageMedium = ""
	StorageMediumMemory    StorageMedium = "Memory"
	StorageMediumHugePages StorageMedium = "HugePages"
)

// GCEPersistentDiskVolumeSource represents a GCE Persistent Disk.
type GCEPersistentDiskVolumeSource struct {
	PDName    string `json:"pdName"`
	FSType    string `json:"fsType,omitempty"`
	Partition int32  `json:"partition,omitempty"`
	ReadOnly  bool   `json:"readOnly,omitempty"`
}

// AWSElasticBlockStoreVolumeSource represents an AWS EBS disk.
type AWSElasticBlockStoreVolumeSource struct {
	VolumeID  string `json:"volumeID"`
	FSType    string `json:"fsType,omitempty"`
	Partition int32  `json:"partition,omitempty"`
	ReadOnly  bool   `json:"readOnly,omitempty"`
}

// ConfigMapVolumeSource adapts a ConfigMap into a volume.
type ConfigMapVolumeSource struct {
	LocalObjectReference `json:",inline"`
	Items                []KeyToPath `json:"items,omitempty"`
	DefaultMode          *int32      `json:"defaultMode,omitempty"`
	Optional             *bool       `json:"optional,omitempty"`
}

// SecretVolumeSource adapts a Secret into a volume.
type SecretVolumeSource struct {
	SecretName  string      `json:"secretName,omitempty"`
	Items       []KeyToPath `json:"items,omitempty"`
	DefaultMode *int32      `json:"defaultMode,omitempty"`
	Optional    *bool       `json:"optional,omitempty"`
}

// NFSVolumeSource represents an NFS mount.
type NFSVolumeSource struct {
	Server   string `json:"server"`
	Path     string `json:"path"`
	ReadOnly bool   `json:"readOnly,omitempty"`
}

// PersistentVolumeClaimVolumeSource references the user's PVC in the same namespace.
type PersistentVolumeClaimVolumeSource struct {
	ClaimName string `json:"claimName"`
	ReadOnly  bool   `json:"readOnly,omitempty"`
}

// ProjectedVolumeSource represents a projected volume.
type ProjectedVolumeSource struct {
	Sources     []VolumeProjection `json:"sources,omitempty"`
	DefaultMode *int32             `json:"defaultMode,omitempty"`
}

// VolumeProjection may be projected along with other supported volume types.
type VolumeProjection struct {
	Secret              *SecretProjection              `json:"secret,omitempty"`
	DownwardAPI         *DownwardAPIProjection         `json:"downwardAPI,omitempty"`
	ConfigMap           *ConfigMapProjection           `json:"configMap,omitempty"`
	ServiceAccountToken *ServiceAccountTokenProjection `json:"serviceAccountToken,omitempty"`
	ClusterTrustBundle  *ClusterTrustBundleProjection  `json:"clusterTrustBundle,omitempty"`
}

// SecretProjection adapts a secret into a projected volume.
type SecretProjection struct {
	LocalObjectReference `json:",inline"`
	Items                []KeyToPath `json:"items,omitempty"`
	Optional             *bool       `json:"optional,omitempty"`
}

// DownwardAPIProjection represents downward API info for a projected volume.
type DownwardAPIProjection struct {
	Items []DownwardAPIVolumeFile `json:"items,omitempty"`
}

// DownwardAPIVolumeFile represents information to project a pod field into a file.
type DownwardAPIVolumeFile struct {
	Path             string                 `json:"path"`
	FieldRef         *ObjectFieldSelector   `json:"fieldRef,omitempty"`
	ResourceFieldRef *ResourceFieldSelector `json:"resourceFieldRef,omitempty"`
	Mode             *int32                 `json:"mode,omitempty"`
}

// ConfigMapProjection adapts a configmap into a projected volume.
type ConfigMapProjection struct {
	LocalObjectReference `json:",inline"`
	Items                []KeyToPath `json:"items,omitempty"`
	Optional             *bool       `json:"optional,omitempty"`
}

// ServiceAccountTokenProjection represents a projected service account token volume.
type ServiceAccountTokenProjection struct {
	Audience          string `json:"audience,omitempty"`
	ExpirationSeconds *int64 `json:"expirationSeconds,omitempty"`
	Path              string `json:"path"`
}

// ClusterTrustBundleProjection describes how to select a set of ClusterTrustBundle objects.
type ClusterTrustBundleProjection struct {
	Name          *string             `json:"name,omitempty"`
	SignerName    *string             `json:"signerName,omitempty"`
	LabelSelector *meta.LabelSelector `json:"labelSelector,omitempty"`
	Optional      *bool               `json:"optional,omitempty"`
	Path          string              `json:"path"`
}

// CSIVolumeSource represents a CSI volume.
type CSIVolumeSource struct {
	Driver               string                `json:"driver"`
	ReadOnly             *bool                 `json:"readOnly,omitempty"`
	FSType               *string               `json:"fsType,omitempty"`
	VolumeAttributes     map[string]string     `json:"volumeAttributes,omitempty"`
	NodePublishSecretRef *LocalObjectReference `json:"nodePublishSecretRef,omitempty"`
}

// EphemeralVolumeSource represents an ephemeral volume to be attached using a VolumeClaimTemplate.
type EphemeralVolumeSource struct {
	VolumeClaimTemplate *PersistentVolumeClaimTemplate `json:"volumeClaimTemplate,omitempty"`
}

// PersistentVolumeClaimTemplate is used to produce PersistentVolumeClaim objects
// as part of an EphemeralVolumeSource.
type PersistentVolumeClaimTemplate struct {
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            PersistentVolumeClaimSpec `json:"spec"`
}

// KeyToPath maps a string key to a path within a volume.
type KeyToPath struct {
	Key  string `json:"key"`
	Path string `json:"path"`
	Mode *int32 `json:"mode,omitempty"`
}

// LocalObjectReference contains enough information to let you locate the referenced object.
type LocalObjectReference struct {
	Name string `json:"name,omitempty"`
}

// Affinity defines the scheduling preferences for a pod.
type Affinity struct {
	NodeAffinity    *NodeAffinity    `json:"nodeAffinity,omitempty"`
	PodAffinity     *PodAffinity     `json:"podAffinity,omitempty"`
	PodAntiAffinity *PodAntiAffinity `json:"podAntiAffinity,omitempty"`
}

// NodeAffinity defines scheduling constraints based on node labels.
type NodeAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution  *NodeSelector             `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
	PreferredDuringSchedulingIgnoredDuringExecution []PreferredSchedulingTerm `json:"preferredDuringSchedulingIgnoredDuringExecution,omitempty"`
}

// NodeSelector represents a union of multiple NodeSelectorTerms.
type NodeSelector struct {
	NodeSelectorTerms []NodeSelectorTerm `json:"nodeSelectorTerms"`
}

// NodeSelectorTerm matches a set of node requirements.
type NodeSelectorTerm struct {
	MatchExpressions []NodeSelectorRequirement `json:"matchExpressions,omitempty"`
	MatchFields      []NodeSelectorRequirement `json:"matchFields,omitempty"`
}

// NodeSelectorRequirement is a selector that contains values, a key, and an operator.
type NodeSelectorRequirement struct {
	Key      string               `json:"key"`
	Operator NodeSelectorOperator `json:"operator"`
	Values   []string             `json:"values,omitempty"`
}

type NodeSelectorOperator string

const (
	NodeSelectorOpIn           NodeSelectorOperator = "In"
	NodeSelectorOpNotIn        NodeSelectorOperator = "NotIn"
	NodeSelectorOpExists       NodeSelectorOperator = "Exists"
	NodeSelectorOpDoesNotExist NodeSelectorOperator = "DoesNotExist"
	NodeSelectorOpGt           NodeSelectorOperator = "Gt"
	NodeSelectorOpLt           NodeSelectorOperator = "Lt"
)

// PreferredSchedulingTerm represents a preferred scheduling term.
type PreferredSchedulingTerm struct {
	Weight     int32            `json:"weight"`
	Preference NodeSelectorTerm `json:"preference"`
}

// PodAffinity specifies pod affinity scheduling rules.
type PodAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution  []PodAffinityTerm         `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
	PreferredDuringSchedulingIgnoredDuringExecution []WeightedPodAffinityTerm `json:"preferredDuringSchedulingIgnoredDuringExecution,omitempty"`
}

// PodAntiAffinity specifies pod anti-affinity scheduling rules.
type PodAntiAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution  []PodAffinityTerm         `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
	PreferredDuringSchedulingIgnoredDuringExecution []WeightedPodAffinityTerm `json:"preferredDuringSchedulingIgnoredDuringExecution,omitempty"`
}

// PodAffinityTerm defines a set of pods that this pod should be co-located with.
type PodAffinityTerm struct {
	LabelSelector     *meta.LabelSelector `json:"labelSelector,omitempty"`
	Namespaces        []string            `json:"namespaces,omitempty"`
	TopologyKey       string              `json:"topologyKey"`
	NamespaceSelector *meta.LabelSelector `json:"namespaceSelector,omitempty"`
	MatchLabelKeys    []string            `json:"matchLabelKeys,omitempty"`
	MismatchLabelKeys []string            `json:"mismatchLabelKeys,omitempty"`
}

// WeightedPodAffinityTerm represents the weights of all the matched WeightedPodAffinityTerm fields.
type WeightedPodAffinityTerm struct {
	Weight          int32           `json:"weight"`
	PodAffinityTerm PodAffinityTerm `json:"podAffinityTerm"`
}

// Toleration allows the pod to be scheduled on nodes with matching taints.
type Toleration struct {
	Key               string             `json:"key,omitempty"`
	Operator          TolerationOperator `json:"operator,omitempty"`
	Value             string             `json:"value,omitempty"`
	Effect            TaintEffect        `json:"effect,omitempty"`
	TolerationSeconds *int64             `json:"tolerationSeconds,omitempty"`
}

type TolerationOperator string

const (
	TolerationOpExists TolerationOperator = "Exists"
	TolerationOpEqual  TolerationOperator = "Equal"
)

// HostAlias holds the mapping between IP and hostnames.
type HostAlias struct {
	IP        string   `json:"ip"`
	Hostnames []string `json:"hostnames,omitempty"`
}

// PodDNSConfig defines the DNS parameters of a pod.
type PodDNSConfig struct {
	Nameservers []string             `json:"nameservers,omitempty"`
	Searches    []string             `json:"searches,omitempty"`
	Options     []PodDNSConfigOption `json:"options,omitempty"`
}

// PodDNSConfigOption defines DNS resolver options.
type PodDNSConfigOption struct {
	Name  string  `json:"name,omitempty"`
	Value *string `json:"value,omitempty"`
}

// PodReadinessGate contains the reference to a pod condition.
type PodReadinessGate struct {
	ConditionType PodConditionType `json:"conditionType"`
}

type PodConditionType string

const (
	ContainersReady           PodConditionType = "ContainersReady"
	PodInitialized            PodConditionType = "Initialized"
	PodReady                  PodConditionType = "Ready"
	PodScheduled              PodConditionType = "PodScheduled"
	DisruptionTarget          PodConditionType = "DisruptionTarget"
	PodReadyToStartContainers PodConditionType = "PodReadyToStartContainers"
)

// TopologySpreadConstraint specifies how to spread matching pods among the given topology.
type TopologySpreadConstraint struct {
	MaxSkew            int32                         `json:"maxSkew"`
	TopologyKey        string                        `json:"topologyKey"`
	WhenUnsatisfiable  UnsatisfiableConstraintAction `json:"whenUnsatisfiable"`
	LabelSelector      *meta.LabelSelector           `json:"labelSelector,omitempty"`
	MinDomains         *int32                        `json:"minDomains,omitempty"`
	NodeAffinityPolicy *NodeInclusionPolicy          `json:"nodeAffinityPolicy,omitempty"`
	NodeTaintsPolicy   *NodeInclusionPolicy          `json:"nodeTaintsPolicy,omitempty"`
	MatchLabelKeys     []string                      `json:"matchLabelKeys,omitempty"`
}

type UnsatisfiableConstraintAction string

const (
	DoNotSchedule  UnsatisfiableConstraintAction = "DoNotSchedule"
	ScheduleAnyway UnsatisfiableConstraintAction = "ScheduleAnyway"
)

type NodeInclusionPolicy string

const (
	NodeInclusionPolicyIgnore NodeInclusionPolicy = "Ignore"
	NodeInclusionPolicyHonor  NodeInclusionPolicy = "Honor"
)

// PodStatus represents information about the status of a pod.
type PodStatus struct {
	Phase                      PodPhase                 `json:"phase,omitempty"`
	Conditions                 []PodCondition           `json:"conditions,omitempty"`
	Message                    string                   `json:"message,omitempty"`
	Reason                     string                   `json:"reason,omitempty"`
	NominatedNodeName          string                   `json:"nominatedNodeName,omitempty"`
	HostIP                     string                   `json:"hostIP,omitempty"`
	HostIPs                    []HostIP                 `json:"hostIPs,omitempty"`
	PodIP                      string                   `json:"podIP,omitempty"`
	PodIPs                     []PodIP                  `json:"podIPs,omitempty"`
	StartTime                  *meta.Time               `json:"startTime,omitempty"`
	InitContainerStatuses      []ContainerStatus        `json:"initContainerStatuses,omitempty"`
	ContainerStatuses          []ContainerStatus        `json:"containerStatuses,omitempty"`
	QOSClass                   PodQOSClass              `json:"qosClass,omitempty"`
	EphemeralContainerStatuses []ContainerStatus        `json:"ephemeralContainerStatuses,omitempty"`
	Resize                     PodResizeStatus          `json:"resize,omitempty"`
	ResourceClaimStatuses      []PodResourceClaimStatus `json:"resourceClaimStatuses,omitempty"`
}

type PodPhase string

const (
	PodPending   PodPhase = "Pending"
	PodRunning   PodPhase = "Running"
	PodSucceeded PodPhase = "Succeeded"
	PodFailed    PodPhase = "Failed"
	PodUnknown   PodPhase = "Unknown"
)

type PodQOSClass string

const (
	PodQOSGuaranteed PodQOSClass = "Guaranteed"
	PodQOSBurstable  PodQOSClass = "Burstable"
	PodQOSBestEffort PodQOSClass = "BestEffort"
)

type PodResizeStatus string

const (
	PodResizeStatusProposed   PodResizeStatus = "Proposed"
	PodResizeStatusInProgress PodResizeStatus = "InProgress"
	PodResizeStatusDeferred   PodResizeStatus = "Deferred"
	PodResizeStatusInfeasible PodResizeStatus = "Infeasible"
)

// PodCondition contains details for the current condition of this pod.
type PodCondition struct {
	Type               PodConditionType     `json:"type"`
	Status             meta.ConditionStatus `json:"status"`
	LastProbeTime      meta.Time            `json:"lastProbeTime,omitempty"`
	LastTransitionTime meta.Time            `json:"lastTransitionTime,omitempty"`
	Reason             string               `json:"reason,omitempty"`
	Message            string               `json:"message,omitempty"`
}

// HostIP represents the IP address of the host.
type HostIP struct {
	IP string `json:"ip"`
}

// PodIP represents the IP address of the pod.
type PodIP struct {
	IP string `json:"ip"`
}

// ContainerStatus contains details for the current status of a container.
type ContainerStatus struct {
	Name                 string                `json:"name"`
	State                ContainerState        `json:"state,omitempty"`
	LastTerminationState ContainerState        `json:"lastState,omitempty"`
	Ready                bool                  `json:"ready"`
	RestartCount         int32                 `json:"restartCount"`
	Image                string                `json:"image"`
	ImageID              string                `json:"imageID"`
	ContainerID          string                `json:"containerID,omitempty"`
	Started              *bool                 `json:"started,omitempty"`
	AllocatedResources   ResourceList          `json:"allocatedResources,omitempty"`
	Resources            *ResourceRequirements `json:"resources,omitempty"`
	VolumeMounts         []VolumeMountStatus   `json:"volumeMounts,omitempty"`
}

// VolumeMountStatus shows status of volume mounts.
type VolumeMountStatus struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
	ReadOnly  bool   `json:"readOnly,omitempty"`
}

// ContainerState holds a possible state of container.
type ContainerState struct {
	Waiting    *ContainerStateWaiting    `json:"waiting,omitempty"`
	Running    *ContainerStateRunning    `json:"running,omitempty"`
	Terminated *ContainerStateTerminated `json:"terminated,omitempty"`
}

// ContainerStateWaiting is a waiting state of a container.
type ContainerStateWaiting struct {
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

// ContainerStateRunning is a running state of a container.
type ContainerStateRunning struct {
	StartedAt meta.Time `json:"startedAt,omitempty"`
}

// ContainerStateTerminated is a terminated state of a container.
type ContainerStateTerminated struct {
	ExitCode    int32     `json:"exitCode"`
	Signal      int32     `json:"signal,omitempty"`
	Reason      string    `json:"reason,omitempty"`
	Message     string    `json:"message,omitempty"`
	StartedAt   meta.Time `json:"startedAt,omitempty"`
	FinishedAt  meta.Time `json:"finishedAt,omitempty"`
	ContainerID string    `json:"containerID,omitempty"`
}

// PodResourceClaimStatus is stored in the PodStatus for each PodResourceClaim which references a ResourceClaimTemplate.
type PodResourceClaimStatus struct {
	Name              string  `json:"name"`
	ResourceClaimName *string `json:"resourceClaimName,omitempty"`
}

// PodList is a list of Pods.
type PodList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []Pod `json:"items"`
}

// ─── Service ─────────────────────────────────────────────────────────────────

// Service is a named abstraction of software service.
type Service struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            ServiceSpec   `json:"spec,omitempty"`
	Status          ServiceStatus `json:"status,omitempty"`
}

// ServiceSpec describes the attributes that a user creates on a service.
type ServiceSpec struct {
	Ports                         []ServicePort                 `json:"ports,omitempty"`
	Selector                      map[string]string             `json:"selector,omitempty"`
	ClusterIP                     string                        `json:"clusterIP,omitempty"`
	ClusterIPs                    []string                      `json:"clusterIPs,omitempty"`
	Type                          ServiceType                   `json:"type,omitempty"`
	ExternalIPs                   []string                      `json:"externalIPs,omitempty"`
	SessionAffinity               ServiceAffinity               `json:"sessionAffinity,omitempty"`
	LoadBalancerIP                string                        `json:"loadBalancerIP,omitempty"`
	LoadBalancerSourceRanges      []string                      `json:"loadBalancerSourceRanges,omitempty"`
	ExternalName                  string                        `json:"externalName,omitempty"`
	ExternalTrafficPolicy         ServiceExternalTrafficPolicy  `json:"externalTrafficPolicy,omitempty"`
	HealthCheckNodePort           int32                         `json:"healthCheckNodePort,omitempty"`
	PublishNotReadyAddresses      bool                          `json:"publishNotReadyAddresses,omitempty"`
	SessionAffinityConfig         *SessionAffinityConfig        `json:"sessionAffinityConfig,omitempty"`
	IPFamilies                    []IPFamily                    `json:"ipFamilies,omitempty"`
	IPFamilyPolicy                *IPFamilyPolicy               `json:"ipFamilyPolicy,omitempty"`
	AllocateLoadBalancerNodePorts *bool                         `json:"allocateLoadBalancerNodePorts,omitempty"`
	LoadBalancerClass             *string                       `json:"loadBalancerClass,omitempty"`
	InternalTrafficPolicy         *ServiceInternalTrafficPolicy `json:"internalTrafficPolicy,omitempty"`
}

type ServiceType string

const (
	ServiceTypeClusterIP    ServiceType = "ClusterIP"
	ServiceTypeNodePort     ServiceType = "NodePort"
	ServiceTypeLoadBalancer ServiceType = "LoadBalancer"
	ServiceTypeExternalName ServiceType = "ExternalName"
)

type ServiceAffinity string

const (
	ServiceAffinityClientIP ServiceAffinity = "ClientIP"
	ServiceAffinityNone     ServiceAffinity = "None"
)

type ServiceExternalTrafficPolicy string

const (
	ServiceExternalTrafficPolicyLocal   ServiceExternalTrafficPolicy = "Local"
	ServiceExternalTrafficPolicyCluster ServiceExternalTrafficPolicy = "Cluster"
)

type ServiceInternalTrafficPolicy string

const (
	ServiceInternalTrafficPolicyLocal   ServiceInternalTrafficPolicy = "Local"
	ServiceInternalTrafficPolicyCluster ServiceInternalTrafficPolicy = "Cluster"
)

type IPFamily string

const (
	IPv4Protocol IPFamily = "IPv4"
	IPv6Protocol IPFamily = "IPv6"
)

type IPFamilyPolicy string

const (
	IPFamilyPolicySingleStack      IPFamilyPolicy = "SingleStack"
	IPFamilyPolicyPreferDualStack  IPFamilyPolicy = "PreferDualStack"
	IPFamilyPolicyRequireDualStack IPFamilyPolicy = "RequireDualStack"
)

// ServicePort contains information on service's port.
type ServicePort struct {
	Name        string      `json:"name,omitempty"`
	Protocol    Protocol    `json:"protocol,omitempty"`
	AppProtocol *string     `json:"appProtocol,omitempty"`
	Port        int32       `json:"port"`
	TargetPort  IntOrString `json:"targetPort,omitempty"`
	NodePort    int32       `json:"nodePort,omitempty"`
}

// SessionAffinityConfig contains the configurations of session affinity.
type SessionAffinityConfig struct {
	ClientIP *ClientIPConfig `json:"clientIP,omitempty"`
}

// ClientIPConfig represents the configurations of Client IP based session affinity.
type ClientIPConfig struct {
	TimeoutSeconds *int32 `json:"timeoutSeconds,omitempty"`
}

// ServiceStatus represents the current status of a service.
type ServiceStatus struct {
	LoadBalancer LoadBalancerStatus `json:"loadBalancer,omitempty"`
	Conditions   []meta.Condition   `json:"conditions,omitempty"`
}

// LoadBalancerStatus represents the status of a load-balancer.
type LoadBalancerStatus struct {
	Ingress []LoadBalancerIngress `json:"ingress,omitempty"`
}

// LoadBalancerIngress represents the status of a load-balancer ingress point.
type LoadBalancerIngress struct {
	IP       string              `json:"ip,omitempty"`
	Hostname string              `json:"hostname,omitempty"`
	IPMode   *LoadBalancerIPMode `json:"ipMode,omitempty"`
	Ports    []PortStatus        `json:"ports,omitempty"`
}

type LoadBalancerIPMode string

const (
	LoadBalancerIPModeVIP   LoadBalancerIPMode = "VIP"
	LoadBalancerIPModeProxy LoadBalancerIPMode = "Proxy"
)

// PortStatus represents the error condition of a service port.
type PortStatus struct {
	Port     int32    `json:"port"`
	Protocol Protocol `json:"protocol"`
	Error    *string  `json:"error,omitempty"`
}

// ServiceList is a list of Services.
type ServiceList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []Service `json:"items"`
}

// ─── ConfigMap ───────────────────────────────────────────────────────────────

// ConfigMap holds configuration data for pods to consume.
type ConfigMap struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	// Immutable, if set to true, ensures that data stored in the ConfigMap cannot be updated.
	Immutable *bool `json:"immutable,omitempty"`
	// Data contains the configuration data.
	Data map[string]string `json:"data,omitempty"`
	// BinaryData contains the binary data.
	BinaryData map[string][]byte `json:"binaryData,omitempty"`
}

// ConfigMapList is a list of ConfigMaps.
type ConfigMapList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []ConfigMap `json:"items"`
}

// ─── Secret ──────────────────────────────────────────────────────────────────

// Secret holds secret data of a certain type.
type Secret struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	// Immutable, if set to true, ensures that data stored in the Secret cannot be updated.
	Immutable *bool `json:"immutable,omitempty"`
	// Data contains the secret data. Each value in the map must be base64 encoded.
	Data map[string][]byte `json:"data,omitempty"`
	// StringData allows specifying non-binary secret data in string form.
	StringData map[string]string `json:"stringData,omitempty"`
	// Type used to facilitate programmatic handling of secret data.
	Type SecretType `json:"type,omitempty"`
}

type SecretType string

const (
	SecretTypeOpaque              SecretType = "Opaque"
	SecretTypeServiceAccountToken SecretType = "kubernetes.io/service-account-token"
	SecretTypeDockercfg           SecretType = "kubernetes.io/dockercfg"
	SecretTypeDockerConfigJSON    SecretType = "kubernetes.io/dockerconfigjson"
	SecretTypeBasicAuth           SecretType = "kubernetes.io/basic-auth"
	SecretTypeSSHAuth             SecretType = "kubernetes.io/ssh-auth"
	SecretTypeTLS                 SecretType = "kubernetes.io/tls"
	SecretTypeBootstrapToken      SecretType = "bootstrap.kubernetes.io/token"
)

// SecretList is a list of Secrets.
type SecretList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []Secret `json:"items"`
}

// ─── ServiceAccount ──────────────────────────────────────────────────────────

// ServiceAccount binds together: a name, understood by users, and a set of
// pods which have access to an optional set of secrets.
type ServiceAccount struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	// Secrets is a list of the secrets in the same namespace that pods running using this ServiceAccount are allowed to use.
	Secrets []ObjectReference `json:"secrets,omitempty"`
	// ImagePullSecrets is a list of references to secrets in the same namespace to use for pulling any images.
	ImagePullSecrets []LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// AutomountServiceAccountToken indicates whether pods running as this service account should have an API token automatically mounted.
	AutomountServiceAccountToken *bool `json:"automountServiceAccountToken,omitempty"`
}

// ObjectReference contains enough information to let you inspect or modify the referred object.
type ObjectReference struct {
	Kind            string `json:"kind,omitempty"`
	Namespace       string `json:"namespace,omitempty"`
	Name            string `json:"name,omitempty"`
	UID             string `json:"uid,omitempty"`
	APIVersion      string `json:"apiVersion,omitempty"`
	ResourceVersion string `json:"resourceVersion,omitempty"`
	FieldPath       string `json:"fieldPath,omitempty"`
}

// ServiceAccountList is a list of ServiceAccount objects.
type ServiceAccountList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []ServiceAccount `json:"items"`
}

// ─── PersistentVolume ────────────────────────────────────────────────────────

// PersistentVolume is a storage resource provisioned by an administrator.
type PersistentVolume struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            PersistentVolumeSpec   `json:"spec,omitempty"`
	Status          PersistentVolumeStatus `json:"status,omitempty"`
}

// PersistentVolumeSpec describes the specification of a persistent volume.
type PersistentVolumeSpec struct {
	Capacity                      ResourceList `json:"capacity,omitempty"`
	PersistentVolumeSource        `json:",inline"`
	AccessModes                   []PersistentVolumeAccessMode  `json:"accessModes,omitempty"`
	ClaimRef                      *ObjectReference              `json:"claimRef,omitempty"`
	PersistentVolumeReclaimPolicy PersistentVolumeReclaimPolicy `json:"persistentVolumeReclaimPolicy,omitempty"`
	StorageClassName              string                        `json:"storageClassName,omitempty"`
	MountOptions                  []string                      `json:"mountOptions,omitempty"`
	VolumeMode                    *PersistentVolumeMode         `json:"volumeMode,omitempty"`
	NodeAffinity                  *VolumeNodeAffinity           `json:"nodeAffinity,omitempty"`
}

// PersistentVolumeSource is similar to VolumeSource but meant for the administrator who creates PVs.
type PersistentVolumeSource struct {
	HostPath *HostPathVolumeSource      `json:"hostPath,omitempty"`
	NFS      *NFSVolumeSource           `json:"nfs,omitempty"`
	CSI      *CSIPersistentVolumeSource `json:"csi,omitempty"`
	Local    *LocalVolumeSource         `json:"local,omitempty"`
}

// CSIPersistentVolumeSource represents storage that is managed by an external CSI volume driver.
type CSIPersistentVolumeSource struct {
	Driver                     string            `json:"driver"`
	VolumeHandle               string            `json:"volumeHandle"`
	ReadOnly                   bool              `json:"readOnly,omitempty"`
	FSType                     string            `json:"fsType,omitempty"`
	VolumeAttributes           map[string]string `json:"volumeAttributes,omitempty"`
	ControllerPublishSecretRef *SecretReference  `json:"controllerPublishSecretRef,omitempty"`
	NodeStageSecretRef         *SecretReference  `json:"nodeStageSecretRef,omitempty"`
	NodePublishSecretRef       *SecretReference  `json:"nodePublishSecretRef,omitempty"`
	ControllerExpandSecretRef  *SecretReference  `json:"controllerExpandSecretRef,omitempty"`
	NodeExpandSecretRef        *SecretReference  `json:"nodeExpandSecretRef,omitempty"`
}

// SecretReference represents a Secret Reference.
type SecretReference struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// LocalVolumeSource represents directly-attached storage with node affinity.
type LocalVolumeSource struct {
	Path   string  `json:"path"`
	FSType *string `json:"fsType,omitempty"`
}

// VolumeNodeAffinity defines constraints that limit what nodes this volume can be accessed from.
type VolumeNodeAffinity struct {
	Required *NodeSelector `json:"required,omitempty"`
}

type PersistentVolumeAccessMode string

const (
	ReadWriteOnce    PersistentVolumeAccessMode = "ReadWriteOnce"
	ReadOnlyMany     PersistentVolumeAccessMode = "ReadOnlyMany"
	ReadWriteMany    PersistentVolumeAccessMode = "ReadWriteMany"
	ReadWriteOncePod PersistentVolumeAccessMode = "ReadWriteOncePod"
)

type PersistentVolumeReclaimPolicy string

const (
	PersistentVolumeReclaimRecycle PersistentVolumeReclaimPolicy = "Recycle"
	PersistentVolumeReclaimDelete  PersistentVolumeReclaimPolicy = "Delete"
	PersistentVolumeReclaimRetain  PersistentVolumeReclaimPolicy = "Retain"
)

type PersistentVolumeMode string

const (
	PersistentVolumeBlock      PersistentVolumeMode = "Block"
	PersistentVolumeFilesystem PersistentVolumeMode = "Filesystem"
)

// PersistentVolumeStatus represents the status of PV.
type PersistentVolumeStatus struct {
	Phase   PersistentVolumePhase `json:"phase,omitempty"`
	Message string                `json:"message,omitempty"`
	Reason  string                `json:"reason,omitempty"`
}

type PersistentVolumePhase string

const (
	VolumePending   PersistentVolumePhase = "Pending"
	VolumeAvailable PersistentVolumePhase = "Available"
	VolumeBound     PersistentVolumePhase = "Bound"
	VolumeReleased  PersistentVolumePhase = "Released"
	VolumeFailed    PersistentVolumePhase = "Failed"
)

// PersistentVolumeList is a list of PersistentVolume items.
type PersistentVolumeList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []PersistentVolume `json:"items"`
}

// ─── PersistentVolumeClaim ───────────────────────────────────────────────────

// PersistentVolumeClaim is a user's request for a volume.
type PersistentVolumeClaim struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            PersistentVolumeClaimSpec   `json:"spec,omitempty"`
	Status          PersistentVolumeClaimStatus `json:"status,omitempty"`
}

// PersistentVolumeClaimSpec describes the common attributes of storage devices.
type PersistentVolumeClaimSpec struct {
	AccessModes               []PersistentVolumeAccessMode `json:"accessModes,omitempty"`
	Selector                  *meta.LabelSelector          `json:"selector,omitempty"`
	Resources                 VolumeResourceRequirements   `json:"resources,omitempty"`
	VolumeName                string                       `json:"volumeName,omitempty"`
	StorageClassName          *string                      `json:"storageClassName,omitempty"`
	VolumeMode                *PersistentVolumeMode        `json:"volumeMode,omitempty"`
	DataSource                *TypedLocalObjectReference   `json:"dataSource,omitempty"`
	DataSourceRef             *TypedObjectReference        `json:"dataSourceRef,omitempty"`
	VolumeAttributesClassName *string                      `json:"volumeAttributesClassName,omitempty"`
}

// VolumeResourceRequirements describes the storage resource requirements for a volume.
type VolumeResourceRequirements struct {
	Limits   ResourceList `json:"limits,omitempty"`
	Requests ResourceList `json:"requests,omitempty"`
}

// TypedLocalObjectReference contains enough information to let you locate the typed referenced object inside the same namespace.
type TypedLocalObjectReference struct {
	APIGroup *string `json:"apiGroup"`
	Kind     string  `json:"kind"`
	Name     string  `json:"name"`
}

// TypedObjectReference contains enough information to let you locate the typed referenced object.
type TypedObjectReference struct {
	APIGroup  *string `json:"apiGroup"`
	Kind      string  `json:"kind"`
	Namespace *string `json:"namespace,omitempty"`
	Name      string  `json:"name"`
}

// PersistentVolumeClaimStatus represents the status of a PVC.
type PersistentVolumeClaimStatus struct {
	Phase                            ClaimPhase                           `json:"phase,omitempty"`
	AccessModes                      []PersistentVolumeAccessMode         `json:"accessModes,omitempty"`
	Capacity                         ResourceList                         `json:"capacity,omitempty"`
	Conditions                       []PersistentVolumeClaimCondition     `json:"conditions,omitempty"`
	AllocatedResources               ResourceList                         `json:"allocatedResources,omitempty"`
	AllocatedResourceStatuses        map[ResourceName]ClaimResourceStatus `json:"allocatedResourceStatuses,omitempty"`
	CurrentVolumeAttributesClassName *string                              `json:"currentVolumeAttributesClassName,omitempty"`
	ModifyVolumeStatus               *ModifyVolumeStatus                  `json:"modifyVolumeStatus,omitempty"`
}

type ClaimPhase string

const (
	ClaimPending ClaimPhase = "Pending"
	ClaimBound   ClaimPhase = "Bound"
	ClaimLost    ClaimPhase = "Lost"
)

type ClaimResourceStatus string

const (
	PersistentVolumeClaimNodeResizePending     ClaimResourceStatus = "NodeResizePending"
	PersistentVolumeClaimNodeResizeInProgress  ClaimResourceStatus = "NodeResizeInProgress"
	PersistentVolumeClaimControllerResizeError ClaimResourceStatus = "ControllerResizeError"
	PersistentVolumeClaimNodeResizeError       ClaimResourceStatus = "NodeResizeError"
)

// PersistentVolumeClaimCondition contains details about the state of a PVC.
type PersistentVolumeClaimCondition struct {
	Type               PersistentVolumeClaimConditionType `json:"type"`
	Status             meta.ConditionStatus               `json:"status"`
	LastProbeTime      meta.Time                          `json:"lastProbeTime,omitempty"`
	LastTransitionTime meta.Time                          `json:"lastTransitionTime,omitempty"`
	Reason             string                             `json:"reason,omitempty"`
	Message            string                             `json:"message,omitempty"`
}

type PersistentVolumeClaimConditionType string

const (
	PersistentVolumeClaimResizing                         PersistentVolumeClaimConditionType = "Resizing"
	PersistentVolumeClaimFileSystemResizePending          PersistentVolumeClaimConditionType = "FileSystemResizePending"
	PersistentVolumeClaimConditionControllerResizeError   PersistentVolumeClaimConditionType = "ControllerResizeError"
	PersistentVolumeClaimConditionNodeResizeError         PersistentVolumeClaimConditionType = "NodeResizeError"
)

// ModifyVolumeStatus represents the status object of ControllerModifyVolume operation.
type ModifyVolumeStatus struct {
	TargetVolumeAttributesClassName string                                  `json:"targetVolumeAttributesClassName,omitempty"`
	Status                          PersistentVolumeClaimModifyVolumeStatus `json:"status"`
}

type PersistentVolumeClaimModifyVolumeStatus string

const (
	PersistentVolumeClaimModifyVolumePending    PersistentVolumeClaimModifyVolumeStatus = "Pending"
	PersistentVolumeClaimModifyVolumeInProgress PersistentVolumeClaimModifyVolumeStatus = "InProgress"
	PersistentVolumeClaimModifyVolumeInfeasible PersistentVolumeClaimModifyVolumeStatus = "Infeasible"
)

// PersistentVolumeClaimList is a list of PVCs.
type PersistentVolumeClaimList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []PersistentVolumeClaim `json:"items"`
}

// ─── ResourceQuota ───────────────────────────────────────────────────────────

// ResourceQuota sets aggregate quota restrictions enforced per namespace.
type ResourceQuota struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            ResourceQuotaSpec   `json:"spec,omitempty"`
	Status          ResourceQuotaStatus `json:"status,omitempty"`
}

// ResourceQuotaSpec defines the desired hard limits to enforce for Quota.
type ResourceQuotaSpec struct {
	Hard          ResourceList         `json:"hard,omitempty"`
	Scopes        []ResourceQuotaScope `json:"scopes,omitempty"`
	ScopeSelector *ScopeSelector       `json:"scopeSelector,omitempty"`
}

type ResourceQuotaScope string

const (
	ResourceQuotaScopeTerminating               ResourceQuotaScope = "Terminating"
	ResourceQuotaScopeNotTerminating            ResourceQuotaScope = "NotTerminating"
	ResourceQuotaScopeBestEffort                ResourceQuotaScope = "BestEffort"
	ResourceQuotaScopeNotBestEffort             ResourceQuotaScope = "NotBestEffort"
	ResourceQuotaScopePriorityClass             ResourceQuotaScope = "PriorityClass"
	ResourceQuotaScopeCrossNamespacePodAffinity ResourceQuotaScope = "CrossNamespacePodAffinity"
)

// ScopeSelector contains a selector for resource quota scope.
type ScopeSelector struct {
	MatchExpressions []ScopedResourceSelectorRequirement `json:"matchExpressions,omitempty"`
}

// ScopedResourceSelectorRequirement is a selector that contains values, a scope name, and an operator.
type ScopedResourceSelectorRequirement struct {
	ScopeName ResourceQuotaScope    `json:"scopeName"`
	Operator  ScopeSelectorOperator `json:"operator"`
	Values    []string              `json:"values,omitempty"`
}

type ScopeSelectorOperator string

const (
	ScopeSelectorOpIn           ScopeSelectorOperator = "In"
	ScopeSelectorOpNotIn        ScopeSelectorOperator = "NotIn"
	ScopeSelectorOpExists       ScopeSelectorOperator = "Exists"
	ScopeSelectorOpDoesNotExist ScopeSelectorOperator = "DoesNotExist"
)

// ResourceQuotaStatus defines the enforced hard limits and observed use.
type ResourceQuotaStatus struct {
	Hard ResourceList `json:"hard,omitempty"`
	Used ResourceList `json:"used,omitempty"`
}

// ResourceQuotaList is a list of ResourceQuota items.
type ResourceQuotaList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []ResourceQuota `json:"items"`
}

// ─── LimitRange ──────────────────────────────────────────────────────────────

// LimitRange sets resource limits for pods, containers, and volumes.
type LimitRange struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            LimitRangeSpec `json:"spec,omitempty"`
}

// LimitRangeSpec defines a min/max usage limit for resources that match on kind.
type LimitRangeSpec struct {
	Limits []LimitRangeItem `json:"limits"`
}

// LimitRangeItem defines a min/max usage limit for any resource that matches on kind.
type LimitRangeItem struct {
	Type                 LimitType    `json:"type"`
	Max                  ResourceList `json:"max,omitempty"`
	Min                  ResourceList `json:"min,omitempty"`
	Default              ResourceList `json:"default,omitempty"`
	DefaultRequest       ResourceList `json:"defaultRequest,omitempty"`
	MaxLimitRequestRatio ResourceList `json:"maxLimitRequestRatio,omitempty"`
}

type LimitType string

const (
	LimitTypePod                   LimitType = "Pod"
	LimitTypeContainer             LimitType = "Container"
	LimitTypePersistentVolumeClaim LimitType = "PersistentVolumeClaim"
)

// LimitRangeList is a list of LimitRange items.
type LimitRangeList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []LimitRange `json:"items"`
}

// ─── Endpoints ───────────────────────────────────────────────────────────────

// Endpoints is a collection of endpoints that implement the actual service.
type Endpoints struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	// Subsets is the set of all endpoints.
	Subsets []EndpointSubset `json:"subsets,omitempty"`
}

// EndpointSubset is a group of addresses with a common set of ports.
type EndpointSubset struct {
	Addresses         []EndpointAddress `json:"addresses,omitempty"`
	NotReadyAddresses []EndpointAddress `json:"notReadyAddresses,omitempty"`
	Ports             []EndpointPort    `json:"ports,omitempty"`
}

// EndpointAddress is a tuple that describes single IP address.
type EndpointAddress struct {
	IP        string           `json:"ip"`
	Hostname  string           `json:"hostname,omitempty"`
	NodeName  *string          `json:"nodeName,omitempty"`
	TargetRef *ObjectReference `json:"targetRef,omitempty"`
}

// EndpointPort is a tuple that describes a single port.
type EndpointPort struct {
	Name        string   `json:"name,omitempty"`
	Port        int32    `json:"port"`
	Protocol    Protocol `json:"protocol,omitempty"`
	AppProtocol *string  `json:"appProtocol,omitempty"`
}

// EndpointsList is a list of endpoints.
type EndpointsList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []Endpoints `json:"items"`
}

// ─── Event ───────────────────────────────────────────────────────────────────

// Event is a report of an event somewhere in the cluster.
type Event struct {
	meta.TypeMeta      `json:",inline"`
	meta.ObjectMeta    `json:"metadata,omitempty"`
	InvolvedObject     ObjectReference  `json:"involvedObject"`
	Reason             string           `json:"reason,omitempty"`
	Message            string           `json:"message,omitempty"`
	Source             EventSource      `json:"source,omitempty"`
	FirstTimestamp     meta.Time        `json:"firstTimestamp,omitempty"`
	LastTimestamp      meta.Time        `json:"lastTimestamp,omitempty"`
	Count              int32            `json:"count,omitempty"`
	Type               string           `json:"type,omitempty"`
	EventTime          meta.Time        `json:"eventTime,omitempty"`
	Series             *EventSeries     `json:"series,omitempty"`
	Action             string           `json:"action,omitempty"`
	Related            *ObjectReference `json:"related,omitempty"`
	ReportingComponent string           `json:"reportingComponent,omitempty"`
	ReportingInstance  string           `json:"reportingInstance,omitempty"`
}

// EventSource contains information for an event.
type EventSource struct {
	Component string `json:"component,omitempty"`
	Host      string `json:"host,omitempty"`
}

// EventSeries contains information about a series of events that all share the same characteristics.
type EventSeries struct {
	Count            int32     `json:"count,omitempty"`
	LastObservedTime meta.Time `json:"lastObservedTime,omitempty"`
}

// EventList is a list of Events.
type EventList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []Event `json:"items"`
}
