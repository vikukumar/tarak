// Package statestore implements Tarak's distributed state layer.
//
// Phase 1 uses BoltDB (bbolt) as the storage engine. The interface is designed
// so that a Raft-replicated or custom WAL-based engine can replace BoltDB in
// later phases without changing callers.
//
// All objects are stored as JSON-encoded Envelopes keyed by ResourceKey.
// A global, monotonically increasing revision counter (stored in BoltDB itself)
// provides optimistic concurrency and MVCC semantics for watch streams.
package statestore

import (
	"encoding/json"
	"fmt"
	"time"
)

// ─── Envelope ─────────────────────────────────────────────────────────────────

// Envelope wraps every stored object with versioning and lifecycle metadata.
// It is the unit of storage; the Object field contains the raw JSON of the resource.
type Envelope struct {
	// ResourceVersion is the monotonically increasing revision at which this object
	// was last written.  It is returned to API clients as metadata.resourceVersion.
	ResourceVersion int64 `json:"rv"`

	// Generation is incremented on every spec change (status changes do not bump it).
	Generation int64 `json:"gen"`

	// UID is the unique identifier of the object, stable across renames.
	UID string `json:"uid"`

	// CreatedAt is the wall-clock time when the object was first persisted.
	CreatedAt time.Time `json:"cat"`

	// UpdatedAt is the wall-clock time of the most recent write.
	UpdatedAt time.Time `json:"uat"`

	// DeletionTimestamp is set when a graceful delete has been initiated.
	DeletionTimestamp *time.Time `json:"dts,omitempty"`

	// Finalizers is a copy of the object's finalizers at the time of last write,
	// stored here for fast GC decisions without deserialising Object.
	Finalizers []string `json:"fin,omitempty"`

	// Labels is a denormalised copy of the object's labels for index maintenance.
	Labels map[string]string `json:"lbl,omitempty"`

	// Namespace is a denormalised copy to avoid deserialising Object for namespace queries.
	Namespace string `json:"ns,omitempty"`

	// Name is a denormalised copy of metadata.name.
	Name string `json:"nm,omitempty"`

	// Object is the raw JSON of the full resource, ready for API responses.
	// resourceVersion inside Object is always synchronised with this envelope's ResourceVersion.
	Object json.RawMessage `json:"obj"`
}

// ─── ResourceKey ──────────────────────────────────────────────────────────────

// ResourceKey uniquely identifies a stored resource.
//
//	Group     – API group, empty for core ("", "apps", "batch", …)
//	Version   – API version ("v1", "v1beta1", …)
//	Resource  – plural resource name ("pods", "deployments", …)
//	Namespace – empty for cluster-scoped resources
//	Name      – resource name
type ResourceKey struct {
	Group     string
	Version   string
	Resource  string
	Namespace string
	Name      string
}

// BucketPath returns the BoltDB bucket path for this resource type:
//
//	"<group>/<version>/<resource>"
//
// The empty core group is represented as "_core" to avoid empty bucket names.
func (k ResourceKey) BucketPath() string {
	g := k.Group
	if g == "" {
		g = "_core"
	}
	return g + "/" + k.Version + "/" + k.Resource
}

// StorageKey returns the BoltDB key within the type bucket:
//
//	"<namespace>/<name>" for namespaced resources
//	"<name>"             for cluster-scoped resources
func (k ResourceKey) StorageKey() string {
	if k.Namespace != "" {
		return k.Namespace + "/" + k.Name
	}
	return k.Name
}

// PrefixForType returns a storage key prefix that iterates all objects of a type.
// For namespaced resources of a specific namespace: "<namespace>/".
// For all objects of a type regardless of namespace: "".
func (k ResourceKey) PrefixForNamespace() string {
	if k.Namespace != "" {
		return k.Namespace + "/"
	}
	return ""
}

// Validate checks that the key has the minimum required fields.
func (k ResourceKey) Validate() error {
	if k.Version == "" {
		return fmt.Errorf("resource key missing Version")
	}
	if k.Resource == "" {
		return fmt.Errorf("resource key missing Resource")
	}
	return nil
}

// ValidateWithName checks that the key includes a name.
func (k ResourceKey) ValidateWithName() error {
	if err := k.Validate(); err != nil {
		return err
	}
	if k.Name == "" {
		return fmt.Errorf("resource key missing Name")
	}
	return nil
}

// String returns a human-readable representation of the key.
func (k ResourceKey) String() string {
	g := k.Group
	if g == "" {
		g = "core"
	}
	if k.Namespace != "" {
		return fmt.Sprintf("%s/%s/%s/%s/%s", g, k.Version, k.Resource, k.Namespace, k.Name)
	}
	return fmt.Sprintf("%s/%s/%s/%s", g, k.Version, k.Resource, k.Name)
}
