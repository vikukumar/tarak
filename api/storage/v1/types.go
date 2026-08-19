// Package v1 contains API types for the storage.k8s.io API group.
package v1

import "github.com/vikukumar/tarak/api/meta"

// ─── StorageClass ─────────────────────────────────────────────────────────────

// StorageClass describes the parameters for a class of storage for which
// PersistentVolumes can be dynamically provisioned.
type StorageClass struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	// Provisioner indicates the type of the provisioner.
	Provisioner string `json:"provisioner"`
	// Parameters holds the parameters for the provisioner that should create volumes of this storage class.
	Parameters map[string]string `json:"parameters,omitempty"`
	// ReclaimPolicy controls the reclaimPolicy for dynamically provisioned PersistentVolumes.
	ReclaimPolicy *string `json:"reclaimPolicy,omitempty"`
	// MountOptions controls the mountOptions for dynamically provisioned PersistentVolumes.
	MountOptions []string `json:"mountOptions,omitempty"`
	// AllowVolumeExpansion shows whether the storage class allow volume expand.
	AllowVolumeExpansion *bool `json:"allowVolumeExpansion,omitempty"`
	// VolumeBindingMode indicates how PersistentVolumeClaims should be provisioned and bound.
	VolumeBindingMode *VolumeBindingMode `json:"volumeBindingMode,omitempty"`
	// AllowedTopologies restrict the node topologies where volumes can be dynamically provisioned.
	AllowedTopologies []TopologySelectorTerm `json:"allowedTopologies,omitempty"`
}

type VolumeBindingMode string

const (
	VolumeBindingImmediate            VolumeBindingMode = "Immediate"
	VolumeBindingWaitForFirstConsumer VolumeBindingMode = "WaitForFirstConsumer"
)

// TopologySelectorTerm represents the result of label queries.
type TopologySelectorTerm struct {
	// A list of topology selector requirements by labels.
	MatchLabelExpressions []TopologySelectorLabelRequirement `json:"matchLabelExpressions,omitempty"`
}

// TopologySelectorLabelRequirement is a selector that matches given label.
type TopologySelectorLabelRequirement struct {
	// The label key that the selector applies to.
	Key string `json:"key"`
	// An array of string values. One value must match the label to be selected.
	Values []string `json:"values"`
}

// StorageClassList is a collection of storage classes.
type StorageClassList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []StorageClass `json:"items"`
}

// ─── VolumeAttachment ────────────────────────────────────────────────────────

// VolumeAttachment captures the intent to attach or detach the specified volume
// to/from the specified node.
type VolumeAttachment struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            VolumeAttachmentSpec   `json:"spec"`
	Status          VolumeAttachmentStatus `json:"status,omitempty"`
}

// VolumeAttachmentSpec is the specification of a VolumeAttachment request.
type VolumeAttachmentSpec struct {
	// Attacher indicates the name of the volume driver that MUST handle this request.
	Attacher string `json:"attacher"`
	// Source represents the volume that should be attached.
	Source VolumeAttachmentSource `json:"source"`
	// NodeName represents the node that the volume should be attached to.
	NodeName string `json:"nodeName"`
}

// VolumeAttachmentSource represents a volume that should be attached.
type VolumeAttachmentSource struct {
	// PersistentVolumeName represents the name of the persistent volume to attach.
	PersistentVolumeName *string `json:"persistentVolumeName,omitempty"`
	// InlineVolumeSpec contains all the information necessary to attach a persistent
	// volume defined by a pod's inline VolumeSource.
	InlineVolumeSpec *VolumeAttachmentInlineSpec `json:"inlineVolumeSpec,omitempty"`
}

// VolumeAttachmentInlineSpec contains a minimal subset of fields of PersistentVolumeSpec
// relevant for volume attachment.
type VolumeAttachmentInlineSpec struct {
	// Driver is the name of the driver to use for this volume.
	Driver string `json:"driver"`
	// VolumeHandle is the unique volume name returned by the CSI volume plugin.
	VolumeHandle string `json:"volumeHandle"`
	// ReadOnly specifies a read-only configuration for the volume.
	ReadOnly bool `json:"readOnly,omitempty"`
	// VolumeAttributes of the volume to publish.
	VolumeAttributes map[string]string `json:"volumeAttributes,omitempty"`
}

// VolumeAttachmentStatus is the status of a VolumeAttachment request.
type VolumeAttachmentStatus struct {
	// Attached indicates the volume is successfully attached.
	Attached bool `json:"attached"`
	// AttachmentMetadata is populated with any information returned by the attach operation.
	AttachmentMetadata map[string]string `json:"attachmentMetadata,omitempty"`
	// AttachError is the last error encountered during attach operation, if any.
	AttachError *VolumeError `json:"attachError,omitempty"`
	// DetachError is the last error encountered during detach operation, if any.
	DetachError *VolumeError `json:"detachError,omitempty"`
}

// VolumeError captures an error encountered during a volume operation.
type VolumeError struct {
	// Time the error was encountered.
	Time meta.Time `json:"time,omitempty"`
	// Message represents the error encountered during Attach or Detach operation.
	Message string `json:"message,omitempty"`
}

// VolumeAttachmentList is a collection of VolumeAttachment objects.
type VolumeAttachmentList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []VolumeAttachment `json:"items"`
}

// ─── CSINode ──────────────────────────────────────────────────────────────────

// CSINode holds information about all CSI drivers installed on a node.
type CSINode struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            CSINodeSpec `json:"spec"`
}

// CSINodeSpec holds information about the specification of all CSI drivers installed on a node.
type CSINodeSpec struct {
	// Drivers is a list of information of all CSI Drivers existing on a node.
	Drivers []CSINodeDriver `json:"drivers"`
}

// CSINodeDriver holds information about the specification of one CSI driver installed on a node.
type CSINodeDriver struct {
	// Name represents the name of the CSI driver that this object refers to.
	Name string `json:"name"`
	// NodeID of the node from the driver point of view.
	NodeID string `json:"nodeID"`
	// TopologyKeys is the list of keys supported by the driver.
	TopologyKeys []string `json:"topologyKeys,omitempty"`
	// Allocatable represents the volume resources of a node that are available for scheduling.
	Allocatable *VolumeNodeResources `json:"allocatable,omitempty"`
}

// VolumeNodeResources is a set of resource limits for scheduling of volumes.
type VolumeNodeResources struct {
	// Maximum number of unique volumes managed by the CSI driver that can be used on a node.
	Count *int32 `json:"count,omitempty"`
}

// CSINodeList is a collection of CSINode objects.
type CSINodeList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []CSINode `json:"items"`
}

// ─── CSIStorageCapacity ───────────────────────────────────────────────────────

// CSIStorageCapacity stores the result of one CSI GetCapacity call.
type CSIStorageCapacity struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	// NodeTopology defines which nodes have access to the storage for which capacity was reported.
	NodeTopology *meta.LabelSelector `json:"nodeTopology,omitempty"`
	// StorageClassName is the name of the StorageClass that the reported capacity applies to.
	StorageClassName string `json:"storageClassName"`
	// Capacity is the value reported by the CSI driver in its GetCapacityResponse for a GetCapacityRequest.
	Capacity *string `json:"capacity,omitempty"`
	// MaximumVolumeSize is the value reported by the CSI driver in its GetCapacityResponse.
	MaximumVolumeSize *string `json:"maximumVolumeSize,omitempty"`
}

// CSIStorageCapacityList is a collection of CSIStorageCapacity objects.
type CSIStorageCapacityList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []CSIStorageCapacity `json:"items"`
}
