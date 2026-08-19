// Package meta defines the common metadata types shared across all Tarak API groups.
// These types are structurally compatible with the Kubernetes API object model so that
// existing Kubernetes-style manifests can be used with Tarak without modification.
package meta

import (
	"encoding/json"
	"time"
)

// ─── TypeMeta ────────────────────────────────────────────────────────────────

// TypeMeta describes an individual object in an API response or request with
// strings representing the type of the object and its API schema version.
// Structures that are versioned or persisted should inline TypeMeta.
type TypeMeta struct {
	// Kind is a string value representing the REST resource this object represents.
	Kind string `json:"kind,omitempty"`
	// APIVersion defines the versioned schema of this representation of an object.
	APIVersion string `json:"apiVersion,omitempty"`
}

// ─── Time ────────────────────────────────────────────────────────────────────

// Time is a wrapper around time.Time that provides RFC3339 JSON marshalling and
// represents a point in time without a monotonic clock component.
type Time struct {
	time.Time `protobuf:"-"`
}

// NewTime constructs a Time from a standard time.Time value.
func NewTime(t time.Time) Time { return Time{t} }

// MarshalJSON implements json.Marshaler. A zero Time marshals as null.
func (t Time) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.UTC().Format(time.RFC3339))
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Time) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*t = Time{}
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

// DeepCopyTime returns a deep copy.
func (t *Time) DeepCopyTime() *Time {
	if t == nil {
		return nil
	}
	out := *t
	return &out
}

// ─── ObjectMeta ──────────────────────────────────────────────────────────────

// ObjectMeta is metadata that all persisted resources must have.  Clients who
// wish to create an object must populate the Name (or GenerateName) field and
// Namespace (for namespaced resources).
type ObjectMeta struct {
	// Name is unique within a namespace.  Required for most objects.
	Name string `json:"name,omitempty"`
	// GenerateName is an optional prefix used to generate a unique name.
	GenerateName string `json:"generateName,omitempty"`
	// Namespace defines the space within which the name must be unique.
	// An empty namespace is equivalent to the "default" namespace.
	Namespace string `json:"namespace,omitempty"`
	// SelfLink is a URL representing this object. Deprecated; populated by the server.
	SelfLink string `json:"selfLink,omitempty"`
	// UID is the unique in time and space value for this object.  Populated by the server.
	UID string `json:"uid,omitempty"`
	// ResourceVersion is an opaque value that clients may use to determine when
	// objects have changed.  It is populated by the server on all objects.
	ResourceVersion string `json:"resourceVersion,omitempty"`
	// Generation is a sequence number representing a specific generation of the desired state.
	Generation int64 `json:"generation,omitempty"`
	// CreationTimestamp is a timestamp representing when this object was created.
	CreationTimestamp Time `json:"creationTimestamp,omitempty"`
	// DeletionTimestamp is RFC 3339 date and time at which this resource will be deleted.
	DeletionTimestamp *Time `json:"deletionTimestamp,omitempty"`
	// DeletionGracePeriodSeconds allows the duration in seconds before the object
	// should be deleted.
	DeletionGracePeriodSeconds *int64 `json:"deletionGracePeriodSeconds,omitempty"`
	// Labels are key/value pairs that are attached to objects.
	Labels map[string]string `json:"labels,omitempty"`
	// Annotations are unstructured key/value pairs stored with a resource.
	Annotations map[string]string `json:"annotations,omitempty"`
	// OwnerReferences is a list of objects depended by this object.
	OwnerReferences []OwnerReference `json:"ownerReferences,omitempty"`
	// Finalizers is an ordered list of finalizers that must be empty before the
	// object is removed from the registry.
	Finalizers []string `json:"finalizers,omitempty"`
	// ManagedFields maps workflow-id and version to the set of fields that are
	// managed by that workflow.
	ManagedFields []ManagedFieldsEntry `json:"managedFields,omitempty"`
}

// ─── ListMeta ────────────────────────────────────────────────────────────────

// ListMeta describes metadata that synthetic resources must have, including
// lists and various status objects.
type ListMeta struct {
	// SelfLink is a URL representing this object. Deprecated.
	SelfLink string `json:"selfLink,omitempty"`
	// ResourceVersion denotes the server state at which this list was generated.
	ResourceVersion string `json:"resourceVersion,omitempty"`
	// Continue is used to retrieve the next page of results when the server
	// returns a token with the continue key.
	Continue string `json:"continue,omitempty"`
	// RemainingItemCount is the number of subsequent items in the list which are
	// not included in this list response.
	RemainingItemCount *int64 `json:"remainingItemCount,omitempty"`
}

// ─── OwnerReference ──────────────────────────────────────────────────────────

// OwnerReference contains enough information to let you identify an owning
// object.  An owning object must be in the same namespace as the dependent.
type OwnerReference struct {
	// APIVersion of the referent.
	APIVersion string `json:"apiVersion"`
	// Kind of the referent.
	Kind string `json:"kind"`
	// Name of the referent.
	Name string `json:"name"`
	// UID of the referent.
	UID string `json:"uid"`
	// If true, this reference points to the managing controller.
	Controller *bool `json:"controller,omitempty"`
	// If true, AND if the owner has the "foregroundDeletion" finalizer, then
	// the owner cannot be deleted from the key-value store until this reference
	// is removed.
	BlockOwnerDeletion *bool `json:"blockOwnerDeletion,omitempty"`
}

// ─── ManagedFieldsEntry ──────────────────────────────────────────────────────

// ManagedFieldsEntry is a workflow-id, a FieldSet and the group version of the
// resource that the fieldset applies to.
type ManagedFieldsEntry struct {
	// Manager is an identifier of the workflow managing these fields.
	Manager string `json:"manager,omitempty"`
	// Operation is the type of operation which lead to this ManagedFieldsEntry
	// being created. The only valid values for this field are 'Apply' and 'Update'.
	Operation string `json:"operation,omitempty"`
	// APIVersion defines the version of this resource that this field set applies to.
	APIVersion string `json:"apiVersion,omitempty"`
	// Time is the timestamp of when the ManagedFields Entry was added.
	Time *Time `json:"time,omitempty"`
	// FieldsType is the discriminator for the different fields format and version.
	FieldsType string `json:"fieldsType,omitempty"`
	// FieldsV1 holds the first JSON version format as described in the "FieldsV1" type.
	FieldsV1 *json.RawMessage `json:"fieldsV1,omitempty"`
	// Subresource is the name of the subresource used to update that object.
	Subresource string `json:"subresource,omitempty"`
}

// ─── Status ──────────────────────────────────────────────────────────────────

// Status is a return value for calls that don't return other objects.  It is
// also used to return error information when the status of a request is failure.
type Status struct {
	TypeMeta `json:",inline"`
	// Standard list metadata.
	Metadata ListMeta `json:"metadata,omitempty"`
	// Status of the operation.  One of: "Success" or "Failure".
	Status string `json:"status,omitempty"`
	// A human-readable description of the status of this operation.
	Message string `json:"message,omitempty"`
	// A machine-readable description of why this operation is in the "Failure" status.
	Reason StatusReason `json:"reason,omitempty"`
	// Extended data associated with the reason.
	Details *StatusDetails `json:"details,omitempty"`
	// Suggested HTTP return code for this status, 0 if not set.
	Code int32 `json:"code,omitempty"`
}

// StatusReason is an enumeration of possible failure causes.
type StatusReason string

const (
	StatusReasonUnknown            StatusReason = ""
	StatusReasonNotFound           StatusReason = "NotFound"
	StatusReasonAlreadyExists      StatusReason = "AlreadyExists"
	StatusReasonConflict           StatusReason = "Conflict"
	StatusReasonGone               StatusReason = "Gone"
	StatusReasonInvalid            StatusReason = "Invalid"
	StatusReasonServerTimeout      StatusReason = "ServerTimeout"
	StatusReasonTimeout            StatusReason = "Timeout"
	StatusReasonTooManyRequests    StatusReason = "TooManyRequests"
	StatusReasonBadRequest         StatusReason = "BadRequest"
	StatusReasonMethodNotAllowed   StatusReason = "MethodNotAllowed"
	StatusReasonNotAcceptable      StatusReason = "NotAcceptable"
	StatusReasonForbidden          StatusReason = "Forbidden"
	StatusReasonUnauthorized       StatusReason = "Unauthorized"
	StatusReasonServiceUnavailable StatusReason = "ServiceUnavailable"
	StatusReasonInternalError      StatusReason = "InternalError"
	StatusReasonExpired            StatusReason = "Expired"
)

// StatusDetails is a set of additional properties that MAY be set by the
// server to provide additional information about a response.
type StatusDetails struct {
	// The name attribute of the resource associated with the status StatusReason.
	Name string `json:"name,omitempty"`
	// The group attribute of the resource associated with the status StatusReason.
	Group string `json:"group,omitempty"`
	// The kind attribute of the resource associated with the status StatusReason.
	Kind string `json:"kind,omitempty"`
	// UID of the resource.
	UID string `json:"uid,omitempty"`
	// The Causes array includes more details associated with the StatusReason failure.
	Causes []StatusCause `json:"causes,omitempty"`
	// If specified, the time in seconds before the operation should be retried.
	RetryAfterSeconds int32 `json:"retryAfterSeconds,omitempty"`
}

// StatusCause provides more information about an api.Status failure.
type StatusCause struct {
	// A machine-readable description of the cause of the error.
	Type CauseType `json:"reason,omitempty"`
	// A human-readable description of the cause of the error.
	Message string `json:"message,omitempty"`
	// The field of the resource that has caused this error.
	Field string `json:"field,omitempty"`
}

// CauseType is a machine-readable value providing more detail about what
// caused a request to fail.
type CauseType string

const (
	CauseTypeFieldValueNotFound      CauseType = "FieldValueNotFound"
	CauseTypeFieldValueRequired      CauseType = "FieldValueRequired"
	CauseTypeFieldValueDuplicate     CauseType = "FieldValueDuplicate"
	CauseTypeFieldValueInvalid       CauseType = "FieldValueInvalid"
	CauseTypeFieldValueNotSupported  CauseType = "FieldValueNotSupported"
	CauseTypeUnexpectedServerResponse CauseType = "UnexpectedServerResponse"
	CauseTypeFieldManagerConflict    CauseType = "FieldManagerConflict"
	CauseTypeResourceVersionTooLarge CauseType = "ResourceVersionTooLarge"
)

// ─── Condition ───────────────────────────────────────────────────────────────

// Condition contains details for one aspect of the current state of this API Resource.
type Condition struct {
	// Type of condition in CamelCase or in foo.example.com/CamelCase.
	Type string `json:"type"`
	// Status of the condition: True, False, or Unknown.
	Status ConditionStatus `json:"status"`
	// ObservedGeneration represents the .metadata.generation that the condition was set based upon.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// LastTransitionTime is the last time the condition transitioned from one status to another.
	LastTransitionTime Time `json:"lastTransitionTime"`
	// Reason contains a programmatic identifier indicating the reason for the condition's last transition.
	Reason string `json:"reason"`
	// Message is a human readable message indicating details about the transition.
	Message string `json:"message"`
}

// ConditionStatus defines conditions of resources.
type ConditionStatus string

const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

// ─── Watch Events ─────────────────────────────────────────────────────────────

// WatchEvent describes a change to an object in the watch stream.
type WatchEvent struct {
	// Type of watch event: ADDED, MODIFIED, DELETED, BOOKMARK, or ERROR.
	Type EventType `json:"type"`
	// Object is the object at the time of the event.
	Object json.RawMessage `json:"object"`
}

// EventType represents the type of watch event.
type EventType string

const (
	EventTypeAdded    EventType = "ADDED"
	EventTypeModified EventType = "MODIFIED"
	EventTypeDeleted  EventType = "DELETED"
	EventTypeError    EventType = "ERROR"
	EventTypeBookmark EventType = "BOOKMARK"
)

// ─── ResourceGroupVersionKind ─────────────────────────────────────────────────

// GroupVersionKind unambiguously identifies a kind.
type GroupVersionKind struct {
	Group   string
	Version string
	Kind    string
}

// String returns a slash-separated representation of the GVK.
func (gvk GroupVersionKind) String() string {
	if gvk.Group == "" {
		return gvk.Version + "/" + gvk.Kind
	}
	return gvk.Group + "/" + gvk.Version + "/" + gvk.Kind
}

// GroupVersionResource unambiguously identifies a resource.
type GroupVersionResource struct {
	Group    string
	Version  string
	Resource string
}

// String returns a slash-separated representation of the GVR.
func (gvr GroupVersionResource) String() string {
	if gvr.Group == "" {
		return gvr.Version + "/" + gvr.Resource
	}
	return gvr.Group + "/" + gvr.Version + "/" + gvr.Resource
}

// ─── Patch Types ──────────────────────────────────────────────────────────────

// PatchType is the type of patch being applied to an object.
type PatchType string

const (
	// JSONPatchType is a patch type using RFC6902 JSON Patch.
	JSONPatchType PatchType = "application/json-patch+json"
	// MergePatchType is a patch type using RFC7396 JSON Merge Patch.
	MergePatchType PatchType = "application/merge-patch+json"
	// StrategicMergePatchType is Kubernetes-style strategic merge patch.
	StrategicMergePatchType PatchType = "application/strategic-merge-patch+json"
	// ApplyPatchType is a patch type used for server-side apply.
	ApplyPatchType PatchType = "application/apply-patch+yaml"
)

// ─── DeleteOptions ────────────────────────────────────────────────────────────

// DeleteOptions may be provided when deleting an API object.
type DeleteOptions struct {
	TypeMeta `json:",inline"`
	// The duration in seconds before the object should be deleted.
	GracePeriodSeconds *int64 `json:"gracePeriodSeconds,omitempty"`
	// Preconditions must be fulfilled before a deletion is carried out.
	Preconditions *Preconditions `json:"preconditions,omitempty"`
	// Deprecated: please use the PropagationPolicy, this field will be deprecated in 1.7.
	OrphanDependents *bool `json:"orphanDependents,omitempty"`
	// Whether and how garbage collection will be performed.
	PropagationPolicy *DeletionPropagation `json:"propagationPolicy,omitempty"`
	// When present, indicates that modifications should not be persisted.
	DryRun []string `json:"dryRun,omitempty"`
}

// DeletionPropagation decides if a deletion will propagate to the dependents of
// the object, and how the garbage collector will handle the propagation.
type DeletionPropagation string

const (
	DeletePropagationOrphan     DeletionPropagation = "Orphan"
	DeletePropagationBackground DeletionPropagation = "Background"
	DeletePropagationForeground DeletionPropagation = "Foreground"
)

// Preconditions must be fulfilled before an operation (update, delete, etc.)
// is carried out.
type Preconditions struct {
	// Specifies the target UID.
	UID *string `json:"uid,omitempty"`
	// Specifies the target ResourceVersion.
	ResourceVersion *string `json:"resourceVersion,omitempty"`
}

// ─── ListOptions ──────────────────────────────────────────────────────────────

// ListOptions is the query options to a standard REST list call.
type ListOptions struct {
	TypeMeta `json:",inline"`
	// A selector to restrict the list of returned objects by their labels.
	LabelSelector string `json:"labelSelector,omitempty"`
	// A selector to restrict the list of returned objects by their fields.
	FieldSelector string `json:"fieldSelector,omitempty"`
	// Watch for changes to the described resources and return them as a stream
	// of add, update, and remove notifications.
	Watch bool `json:"watch,omitempty"`
	// allowWatchBookmarks requests watch events with type "BOOKMARK".
	AllowWatchBookmarks bool `json:"allowWatchBookmarks,omitempty"`
	// resourceVersion sets a constraint on what resource versions a request may be served from.
	ResourceVersion string `json:"resourceVersion,omitempty"`
	// ResourceVersionMatch determines how resourceVersion is applied to list calls.
	ResourceVersionMatch ResourceVersionMatch `json:"resourceVersionMatch,omitempty"`
	// TimeoutSeconds specifies the seconds of ClientIP type session sticky time.
	TimeoutSeconds *int64 `json:"timeoutSeconds,omitempty"`
	// Limit is a maximum number of responses to return for a list call.
	Limit int64 `json:"limit,omitempty"`
	// Continue is used to retrieve the next page of results for list calls.
	Continue string `json:"continue,omitempty"`
	// sendInitialEvents controls whether the server will send a stream of
	// initial events for a watch call.
	SendInitialEvents *bool `json:"sendInitialEvents,omitempty"`
}

// ResourceVersionMatch specifies how the ResourceVersion parameter is applied.
type ResourceVersionMatch string

const (
	ResourceVersionMatchNotOlderThan ResourceVersionMatch = "NotOlderThan"
	ResourceVersionMatchExact        ResourceVersionMatch = "Exact"
)

// ─── GetOptions ───────────────────────────────────────────────────────────────

// GetOptions is the standard query options to the standard REST get call.
type GetOptions struct {
	TypeMeta `json:",inline"`
	// ResourceVersion for the object at an exact version.
	ResourceVersion string `json:"resourceVersion,omitempty"`
}

// ─── CreateOptions ────────────────────────────────────────────────────────────

// CreateOptions may be provided when creating an API object.
type CreateOptions struct {
	TypeMeta `json:",inline"`
	// When present, indicates that modifications should not be persisted.
	DryRun []string `json:"dryRun,omitempty"`
	// fieldManager is a name associated with the actor that is making changes.
	FieldManager string `json:"fieldManager,omitempty"`
	// fieldValidation instructs the server on how to handle objects in the request.
	FieldValidation string `json:"fieldValidation,omitempty"`
}

// ─── UpdateOptions ────────────────────────────────────────────────────────────

// UpdateOptions may be provided when updating an API object.
type UpdateOptions struct {
	TypeMeta `json:",inline"`
	// When present, indicates that modifications should not be persisted.
	DryRun []string `json:"dryRun,omitempty"`
	// fieldManager is a name associated with the actor that is making changes.
	FieldManager string `json:"fieldManager,omitempty"`
	// fieldValidation instructs the server on how to handle objects in the request.
	FieldValidation string `json:"fieldValidation,omitempty"`
}

// ─── PatchOptions ─────────────────────────────────────────────────────────────

// PatchOptions may be provided when patching an API object.
type PatchOptions struct {
	TypeMeta `json:",inline"`
	// When present, indicates that modifications should not be persisted.
	DryRun []string `json:"dryRun,omitempty"`
	// Force is going to "force" Apply requests.
	Force *bool `json:"force,omitempty"`
	// fieldManager is a name associated with the actor that is making changes.
	FieldManager string `json:"fieldManager,omitempty"`
	// fieldValidation instructs the server on how to handle objects in the request.
	FieldValidation string `json:"fieldValidation,omitempty"`
}
