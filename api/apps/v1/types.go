// Package v1 contains API types for the apps API group.
package v1

import (
	corev1 "github.com/vikukumar/tarak/api/core/v1"
	"github.com/vikukumar/tarak/api/meta"
)

// ─── Deployment ──────────────────────────────────────────────────────────────

// Deployment enables declarative updates for Pods and ReplicaSets.
type Deployment struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            DeploymentSpec   `json:"spec,omitempty"`
	Status          DeploymentStatus `json:"status,omitempty"`
}

// DeploymentSpec is the specification of the desired behaviour of the Deployment.
type DeploymentSpec struct {
	// Replicas is the desired number of replicas of the given Template.
	Replicas *int32 `json:"replicas,omitempty"`
	// Selector is a label query over pods that should match the replica count.
	Selector *meta.LabelSelector `json:"selector"`
	// Template describes the pods that will be created.
	Template PodTemplateSpec `json:"template"`
	// Strategy describes how to replace existing pods with new ones.
	Strategy DeploymentStrategy `json:"strategy,omitempty"`
	// MinReadySeconds is the minimum number of seconds for which a newly created pod should be ready.
	MinReadySeconds int32 `json:"minReadySeconds,omitempty"`
	// RevisionHistoryLimit is the number of old ReplicaSets to retain to allow rollback.
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty"`
	// Paused indicates that the deployment is paused.
	Paused bool `json:"paused,omitempty"`
	// ProgressDeadlineSeconds is the maximum time in seconds for a deployment to make progress.
	ProgressDeadlineSeconds *int32 `json:"progressDeadlineSeconds,omitempty"`
}

// DeploymentStrategy describes how to replace existing pods with new ones.
type DeploymentStrategy struct {
	// Type of deployment. Can be "Recreate" or "RollingUpdate".
	Type DeploymentStrategyType `json:"type,omitempty"`
	// Rolling update config params. Present only if DeploymentStrategyType = RollingUpdate.
	RollingUpdate *RollingUpdateDeployment `json:"rollingUpdate,omitempty"`
}

type DeploymentStrategyType string

const (
	RecreateDeploymentStrategyType      DeploymentStrategyType = "Recreate"
	RollingUpdateDeploymentStrategyType DeploymentStrategyType = "RollingUpdate"
)

// RollingUpdateDeployment specifies the parameters to be used for the rolling update deployment.
type RollingUpdateDeployment struct {
	// MaxUnavailable is the maximum number of pods that can be unavailable during the update.
	MaxUnavailable *corev1.IntOrString `json:"maxUnavailable,omitempty"`
	// MaxSurge is the maximum number of pods that can be scheduled above the desired number of pods.
	MaxSurge *corev1.IntOrString `json:"maxSurge,omitempty"`
}

// DeploymentStatus is the most recently observed status of the Deployment.
type DeploymentStatus struct {
	// ObservedGeneration is the generation most recently observed by the deployment controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Replicas is the total number of non-terminated pods targeted by this deployment.
	Replicas int32 `json:"replicas,omitempty"`
	// UpdatedReplicas is the total number of non-terminated pods targeted by this deployment that have the desired template spec.
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty"`
	// ReadyReplicas is the number of pods targeted by this Deployment with a Ready Condition.
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
	// AvailableReplicas is the total number of available pods (ready for at least minReadySeconds).
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`
	// UnavailableReplicas is the total number of unavailable pods.
	UnavailableReplicas int32 `json:"unavailableReplicas,omitempty"`
	// Conditions represent the latest available observations of a deployment's current state.
	Conditions []DeploymentCondition `json:"conditions,omitempty"`
	// CollisionCount is the count of hash collisions for the Deployment.
	CollisionCount *int32 `json:"collisionCount,omitempty"`
}

// DeploymentCondition describes the state of a deployment at a certain point.
type DeploymentCondition struct {
	Type               DeploymentConditionType `json:"type"`
	Status             meta.ConditionStatus    `json:"status"`
	LastUpdateTime     meta.Time               `json:"lastUpdateTime,omitempty"`
	LastTransitionTime meta.Time               `json:"lastTransitionTime,omitempty"`
	Reason             string                  `json:"reason,omitempty"`
	Message            string                  `json:"message,omitempty"`
}

type DeploymentConditionType string

const (
	DeploymentAvailable      DeploymentConditionType = "Available"
	DeploymentProgressing    DeploymentConditionType = "Progressing"
	DeploymentReplicaFailure DeploymentConditionType = "ReplicaFailure"
)

// DeploymentList is a list of Deployments.
type DeploymentList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []Deployment `json:"items"`
}

// ─── ReplicaSet ──────────────────────────────────────────────────────────────

// ReplicaSet ensures a specified number of pod "replicas" are running at any given time.
type ReplicaSet struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            ReplicaSetSpec   `json:"spec,omitempty"`
	Status          ReplicaSetStatus `json:"status,omitempty"`
}

// ReplicaSetSpec is the specification of a ReplicaSet.
type ReplicaSetSpec struct {
	// Replicas is the number of desired replicas.
	Replicas *int32 `json:"replicas,omitempty"`
	// MinReadySeconds is the minimum number of seconds for which a newly created pod should be ready.
	MinReadySeconds int32 `json:"minReadySeconds,omitempty"`
	// Selector is a label query over pods that should match the replica count.
	Selector *meta.LabelSelector `json:"selector"`
	// Template is the object that describes the pod that will be created.
	Template PodTemplateSpec `json:"template,omitempty"`
}

// ReplicaSetStatus represents the current status of a ReplicaSet.
type ReplicaSetStatus struct {
	// Replicas is the most recently observed number of replicas.
	Replicas int32 `json:"replicas"`
	// FullyLabeledReplicas is the number of pods that have labels matching the labels of the pod template.
	FullyLabeledReplicas int32 `json:"fullyLabeledReplicas,omitempty"`
	// ReadyReplicas is the number of pods targeted by this ReplicaSet with a Ready Condition.
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
	// AvailableReplicas is the number of available replicas.
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`
	// ObservedGeneration reflects the generation of the most recently observed ReplicaSet.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions is a list of ReplicaSet conditions.
	Conditions []ReplicaSetCondition `json:"conditions,omitempty"`
}

// ReplicaSetCondition describes the state of a replica set at a certain point.
type ReplicaSetCondition struct {
	Type               ReplicaSetConditionType `json:"type"`
	Status             meta.ConditionStatus    `json:"status"`
	LastTransitionTime meta.Time               `json:"lastTransitionTime,omitempty"`
	Reason             string                  `json:"reason,omitempty"`
	Message            string                  `json:"message,omitempty"`
}

type ReplicaSetConditionType string

const (
	ReplicaSetReplicaFailure ReplicaSetConditionType = "ReplicaFailure"
)

// ReplicaSetList is a list of ReplicaSets.
type ReplicaSetList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []ReplicaSet `json:"items"`
}

// ─── StatefulSet ─────────────────────────────────────────────────────────────

// StatefulSet represents a set of pods with consistent identities.
type StatefulSet struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            StatefulSetSpec   `json:"spec,omitempty"`
	Status          StatefulSetStatus `json:"status,omitempty"`
}

// StatefulSetSpec is the specification of a StatefulSet.
type StatefulSetSpec struct {
	// Replicas is the desired number of replicas of the given Template.
	Replicas *int32 `json:"replicas,omitempty"`
	// Selector is a label query over pods that should match the replica count.
	Selector *meta.LabelSelector `json:"selector"`
	// Template is the object that describes the pod that will be created.
	Template PodTemplateSpec `json:"template"`
	// VolumeClaimTemplates is a list of claims that pods are allowed to reference.
	VolumeClaimTemplates []corev1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`
	// ServiceName is the name of the service that governs this StatefulSet.
	ServiceName string `json:"serviceName"`
	// PodManagementPolicy controls how pods are created during initial scale up.
	PodManagementPolicy PodManagementPolicyType `json:"podManagementPolicy,omitempty"`
	// UpdateStrategy indicates the StatefulSetUpdateStrategy that will be employed.
	UpdateStrategy StatefulSetUpdateStrategy `json:"updateStrategy,omitempty"`
	// RevisionHistoryLimit is the maximum number of revisions that will be maintained in the StatefulSet's revision history.
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty"`
	// MinReadySeconds is the minimum number of seconds for which a newly created pod should be ready.
	MinReadySeconds int32 `json:"minReadySeconds,omitempty"`
	// PersistentVolumeClaimRetentionPolicy describes the lifecycle of PVCs created from VolumeClaimTemplates.
	PersistentVolumeClaimRetentionPolicy *StatefulSetPersistentVolumeClaimRetentionPolicy `json:"persistentVolumeClaimRetentionPolicy,omitempty"`
	// Ordinals controls the numbering of replica indices in a StatefulSet.
	Ordinals *StatefulSetOrdinals `json:"ordinals,omitempty"`
}

type PodManagementPolicyType string

const (
	OrderedReadyPodManagement PodManagementPolicyType = "OrderedReady"
	ParallelPodManagement     PodManagementPolicyType = "Parallel"
)

// StatefulSetUpdateStrategy indicates the strategy that the StatefulSet controller will use to perform updates.
type StatefulSetUpdateStrategy struct {
	Type          StatefulSetUpdateStrategyType     `json:"type,omitempty"`
	RollingUpdate *RollingUpdateStatefulSetStrategy `json:"rollingUpdate,omitempty"`
}

type StatefulSetUpdateStrategyType string

const (
	RollingUpdateStatefulSetStrategyType StatefulSetUpdateStrategyType = "RollingUpdate"
	OnDeleteStatefulSetStrategyType      StatefulSetUpdateStrategyType = "OnDelete"
)

// RollingUpdateStatefulSetStrategy is used to communicate parameter for RollingUpdateStatefulSetStrategyType.
type RollingUpdateStatefulSetStrategy struct {
	Partition      *int32              `json:"partition,omitempty"`
	MaxUnavailable *corev1.IntOrString `json:"maxUnavailable,omitempty"`
}

// StatefulSetPersistentVolumeClaimRetentionPolicy describes the policy used for PVCs created from the StatefulSet VolumeClaimTemplates.
type StatefulSetPersistentVolumeClaimRetentionPolicy struct {
	WhenDeleted PersistentVolumeClaimRetentionPolicyType `json:"whenDeleted,omitempty"`
	WhenScaled  PersistentVolumeClaimRetentionPolicyType `json:"whenScaled,omitempty"`
}

type PersistentVolumeClaimRetentionPolicyType string

const (
	RetainPersistentVolumeClaimRetentionPolicyType PersistentVolumeClaimRetentionPolicyType = "Retain"
	DeletePersistentVolumeClaimRetentionPolicyType PersistentVolumeClaimRetentionPolicyType = "Delete"
)

// StatefulSetOrdinals describes the numbering of replica indices in a StatefulSet.
type StatefulSetOrdinals struct {
	Start int32 `json:"start,omitempty"`
}

// StatefulSetStatus represents the current state of a StatefulSet.
type StatefulSetStatus struct {
	// ObservedGeneration is the most recent generation observed for this StatefulSet.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Replicas is the number of Pods created by the StatefulSet controller.
	Replicas int32 `json:"replicas"`
	// ReadyReplicas is the number of pods created for this StatefulSet with a Ready Condition.
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
	// CurrentReplicas is the number of Pods created by the StatefulSet controller from the StatefulSet version indicated by currentRevision.
	CurrentReplicas int32 `json:"currentReplicas,omitempty"`
	// UpdatedReplicas is the number of Pods created by the StatefulSet controller from the StatefulSet version indicated by updateRevision.
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty"`
	// CurrentRevision, if not empty, indicates the version of the StatefulSet used to generate Pods in the sequence [0,currentReplicas).
	CurrentRevision string `json:"currentRevision,omitempty"`
	// UpdateRevision, if not empty, indicates the version of the StatefulSet used to generate Pods in the sequence [replicas-updatedReplicas,replicas).
	UpdateRevision string `json:"updateRevision,omitempty"`
	// CollisionCount is the count of hash collisions for the StatefulSet.
	CollisionCount *int32 `json:"collisionCount,omitempty"`
	// Conditions is a list of StatefulSet conditions.
	Conditions []StatefulSetCondition `json:"conditions,omitempty"`
	// AvailableReplicas is the total number of available pods (ready for at least minReadySeconds).
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`
}

// StatefulSetCondition describes the state of a statefulset at a certain point.
type StatefulSetCondition struct {
	Type               StatefulSetConditionType `json:"type"`
	Status             meta.ConditionStatus     `json:"status"`
	LastTransitionTime meta.Time                `json:"lastTransitionTime,omitempty"`
	Reason             string                   `json:"reason,omitempty"`
	Message            string                   `json:"message,omitempty"`
}

type StatefulSetConditionType string

// StatefulSetList is a list of StatefulSets.
type StatefulSetList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []StatefulSet `json:"items"`
}

// ─── DaemonSet ───────────────────────────────────────────────────────────────

// DaemonSet represents the configuration of a daemon set.
type DaemonSet struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            DaemonSetSpec   `json:"spec,omitempty"`
	Status          DaemonSetStatus `json:"status,omitempty"`
}

// DaemonSetSpec is the specification of a daemon set.
type DaemonSetSpec struct {
	// Selector is a label query over pods that are managed by the daemon set.
	Selector *meta.LabelSelector `json:"selector"`
	// Template is an object that describes the pod that will be created.
	Template PodTemplateSpec `json:"template"`
	// UpdateStrategy is an update strategy for the DaemonSet.
	UpdateStrategy DaemonSetUpdateStrategy `json:"updateStrategy,omitempty"`
	// MinReadySeconds is the minimum number of seconds for which a newly created DaemonSet pod should be ready.
	MinReadySeconds int32 `json:"minReadySeconds,omitempty"`
	// RevisionHistoryLimit is the number of old history to retain to allow rollback.
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty"`
}

// DaemonSetUpdateStrategy is a struct used to control the update strategy for a DaemonSet.
type DaemonSetUpdateStrategy struct {
	// Type of daemon set update.
	Type DaemonSetUpdateStrategyType `json:"type,omitempty"`
	// Rolling update config params. Present only if type = "RollingUpdate".
	RollingUpdate *RollingUpdateDaemonSet `json:"rollingUpdate,omitempty"`
}

type DaemonSetUpdateStrategyType string

const (
	RollingUpdateDaemonSetStrategyType DaemonSetUpdateStrategyType = "RollingUpdate"
	OnDeleteDaemonSetStrategyType      DaemonSetUpdateStrategyType = "OnDelete"
)

// RollingUpdateDaemonSet specifies the parameters for rolling update daemonset.
type RollingUpdateDaemonSet struct {
	// MaxUnavailable is the maximum number of DaemonSet pods that can be unavailable during the update.
	MaxUnavailable *corev1.IntOrString `json:"maxUnavailable,omitempty"`
	// MaxSurge is the maximum number of nodes with an existing available DaemonSet pod that can have an updated DaemonSet pod during during an update.
	MaxSurge *corev1.IntOrString `json:"maxSurge,omitempty"`
}

// DaemonSetStatus is the most recently observed status of the DaemonSet.
type DaemonSetStatus struct {
	// CurrentNumberScheduled is the number of nodes that are running at least 1 daemon pod and are supposed to run the daemon pod.
	CurrentNumberScheduled int32 `json:"currentNumberScheduled"`
	// NumberMisscheduled is the number of nodes that are running the daemon pod, but are not supposed to run the daemon pod.
	NumberMisscheduled int32 `json:"numberMisscheduled"`
	// DesiredNumberScheduled is the total number of nodes that should be running the daemon pod.
	DesiredNumberScheduled int32 `json:"desiredNumberScheduled"`
	// NumberReady is the number of nodes that should be running the daemon pod and have one or more of the daemon pod running and ready.
	NumberReady int32 `json:"numberReady"`
	// ObservedGeneration is the most recent generation observed by the daemon set controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// UpdatedNumberScheduled is the total number of nodes that are running updated daemon pod.
	UpdatedNumberScheduled int32 `json:"updatedNumberScheduled,omitempty"`
	// NumberAvailable is the number of nodes that should be running the daemon pod and have one or more of the daemon pod running and available.
	NumberAvailable int32 `json:"numberAvailable,omitempty"`
	// NumberUnavailable is the number of nodes that should be running the daemon pod and have none of the daemon pod running and available.
	NumberUnavailable int32 `json:"numberUnavailable,omitempty"`
	// CollisionCount is the count of hash collisions for the DaemonSet.
	CollisionCount *int32 `json:"collisionCount,omitempty"`
	// Conditions represent the latest available observations of a DaemonSet's current state.
	Conditions []DaemonSetCondition `json:"conditions,omitempty"`
}

// DaemonSetCondition describes the state of a DaemonSet at a certain point.
type DaemonSetCondition struct {
	Type               DaemonSetConditionType `json:"type"`
	Status             meta.ConditionStatus   `json:"status"`
	LastTransitionTime meta.Time              `json:"lastTransitionTime,omitempty"`
	Reason             string                 `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
}

type DaemonSetConditionType string

// DaemonSetList is a list of DaemonSets.
type DaemonSetList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []DaemonSet `json:"items"`
}

// ─── PodTemplateSpec ─────────────────────────────────────────────────────────

// PodTemplateSpec describes the data a pod should have when created from a template.
type PodTemplateSpec struct {
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            corev1.PodSpec `json:"spec,omitempty"`
}

// ─── ControllerRevision ──────────────────────────────────────────────────────

// ControllerRevision implements an immutable snapshot of state data.
type ControllerRevision struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	// Data is the serialized representation of the state.
	Data []byte `json:"data,omitempty"`
	// Revision indicates the revision of the state represented by Data.
	Revision int64 `json:"revision"`
}

// ControllerRevisionList is a resource containing a list of ControllerRevision objects.
type ControllerRevisionList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []ControllerRevision `json:"items"`
}
