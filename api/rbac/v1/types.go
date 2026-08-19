// Package v1 contains API types for the rbac.authorization.k8s.io API group.
package v1

import "github.com/vikukumar/tarak/api/meta"

// ─── PolicyRule ──────────────────────────────────────────────────────────────

// PolicyRule holds information that describes a policy rule, but does not contain
// information about who the rule applies to or which namespace the rule applies to.
type PolicyRule struct {
	// Verbs is a list of Verbs that apply to ALL the ResourceKinds contained in this rule.
	Verbs []string `json:"verbs"`
	// APIGroups is the name of the APIGroup that contains the resources.
	APIGroups []string `json:"apiGroups,omitempty"`
	// Resources is a list of resources this rule applies to.
	Resources []string `json:"resources,omitempty"`
	// ResourceNames is an optional white list of names that the rule applies to.
	ResourceNames []string `json:"resourceNames,omitempty"`
	// NonResourceURLs is a set of partial urls that a user should have access to.
	NonResourceURLs []string `json:"nonResourceURLs,omitempty"`
}

// Subject contains a reference to the object or user identities a role binding applies to.
type Subject struct {
	// Kind of object being referenced.
	Kind string `json:"kind"`
	// APIGroup holds the API group of the referenced subject.
	APIGroup string `json:"apiGroup,omitempty"`
	// Name of the object being referenced.
	Name string `json:"name"`
	// Namespace of the referenced object.
	Namespace string `json:"namespace,omitempty"`
}

// RoleRef contains information that points to the role being used.
type RoleRef struct {
	// APIGroup is the group for the resource being referenced.
	APIGroup string `json:"apiGroup"`
	// Kind is the type of resource being referenced.
	Kind string `json:"kind"`
	// Name is the name of resource being referenced.
	Name string `json:"name"`
}

// ─── Role ────────────────────────────────────────────────────────────────────

// Role is a namespaced, logical grouping of PolicyRules that can be referenced
// as a unit by a RoleBinding.
type Role struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	// Rules holds all the PolicyRules for this Role.
	Rules []PolicyRule `json:"rules,omitempty"`
}

// RoleList is a collection of Roles.
type RoleList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []Role `json:"items"`
}

// ─── RoleBinding ─────────────────────────────────────────────────────────────

// RoleBinding references a role, but does not contain it. It can reference a
// Role in the same namespace or a ClusterRole in the global namespace.
type RoleBinding struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	// Subjects holds references to the objects the role applies to.
	Subjects []Subject `json:"subjects,omitempty"`
	// RoleRef can reference a Role in the current namespace or a ClusterRole in the global namespace.
	RoleRef RoleRef `json:"roleRef"`
}

// RoleBindingList is a collection of RoleBindings.
type RoleBindingList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []RoleBinding `json:"items"`
}

// ─── ClusterRole ─────────────────────────────────────────────────────────────

// ClusterRole is a cluster level, logical grouping of PolicyRules that can be
// referenced as a unit by a ClusterRoleBinding or a RoleBinding (within a namespace).
type ClusterRole struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	// Rules holds all the PolicyRules for this ClusterRole.
	Rules []PolicyRule `json:"rules,omitempty"`
	// AggregationRule is an optional field that describes how to build the Rules for this ClusterRole.
	AggregationRule *AggregationRule `json:"aggregationRule,omitempty"`
}

// AggregationRule describes how to locate ClusterRoles to aggregate into the ClusterRole.
type AggregationRule struct {
	// ClusterRoleSelectors holds a list of selectors which will be used to find ClusterRoles
	// and create the rules.
	ClusterRoleSelectors []meta.LabelSelector `json:"clusterRoleSelectors,omitempty"`
}

// ClusterRoleList is a collection of ClusterRoles.
type ClusterRoleList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []ClusterRole `json:"items"`
}

// ─── ClusterRoleBinding ──────────────────────────────────────────────────────

// ClusterRoleBinding references a ClusterRole, but not contain it. It can reference
// a ClusterRole in the global namespace, and adds who information via Subjects.
type ClusterRoleBinding struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	// Subjects holds references to the objects the role applies to.
	Subjects []Subject `json:"subjects,omitempty"`
	// RoleRef can only reference a ClusterRole in the global namespace.
	RoleRef RoleRef `json:"roleRef"`
}

// ClusterRoleBindingList is a collection of ClusterRoleBindings.
type ClusterRoleBindingList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []ClusterRoleBinding `json:"items"`
}
