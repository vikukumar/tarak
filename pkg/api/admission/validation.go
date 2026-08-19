// Package admission implements server-side validation of API objects.
//
// Every CREATE and UPDATE request passes through this layer before the
// object reaches the state store.  Validation checks:
//
//   - Required fields are present (apiVersion, kind, metadata.name)
//   - Names match RFC 1123 DNS label constraints
//   - Namespaces are valid (no slashes, no reserved characters)
//   - Immutable fields are not changed on update (UID, creationTimestamp)
//   - Resource-specific constraints (e.g., container image not empty)
//
// Admission does NOT check whether a namespace exists — referential integrity
// is deferred to Phase 7.  This keeps Phase 1 simple while still ensuring
// objects stored in the state store are structurally valid.
package admission

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ─── Error types ──────────────────────────────────────────────────────────────

// ValidationError is returned when an object fails admission validation.
type ValidationError struct {
	// Field is the JSON path of the invalid field (e.g., "metadata.name").
	Field string
	// Reason describes why the field is invalid.
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid field %q: %s", e.Field, e.Reason)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []*ValidationError

func (ve ValidationErrors) Error() string {
	msgs := make([]string, 0, len(ve))
	for _, e := range ve {
		msgs = append(msgs, e.Error())
	}
	return strings.Join(msgs, "; ")
}

// ─── Patterns ─────────────────────────────────────────────────────────────────

var (
	// dns1123Label matches a Kubernetes DNS label segment (RFC 1123).
	dns1123Label = regexp.MustCompile(`^[a-z0-9]([a-z0-9\-]{0,61}[a-z0-9])?$`)
	// dns1123Subdomain matches a Kubernetes DNS subdomain (dots allowed).
	dns1123Subdomain = regexp.MustCompile(`^[a-z0-9]([a-z0-9\-\.]{0,251}[a-z0-9])?$`)
)

// ─── Object structure ─────────────────────────────────────────────────────────

// genericObject is the minimal shape we expect from every API object.
type genericObject struct {
	APIVersion string         `json:"apiVersion"`
	Kind       string         `json:"kind"`
	Metadata   objectMetadata `json:"metadata"`
}

type objectMetadata struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Labels            map[string]string `json:"labels"`
	Annotations       map[string]string `json:"annotations"`
	UID               string            `json:"uid"`
	ResourceVersion   string            `json:"resourceVersion"`
	GenerateName      string            `json:"generateName"`
}

// ─── Validator ────────────────────────────────────────────────────────────────

// Validator performs admission validation on API objects.
type Validator struct{}

// New returns a new Validator.
func New() *Validator {
	return &Validator{}
}

// ValidateCreate validates an object submitted for creation.
// The rawObj is the full JSON body of the request.
func (v *Validator) ValidateCreate(kind string, rawObj []byte) error {
	obj, errs := parseAndValidateCommon(rawObj)
	if len(errs) > 0 {
		return errs
	}
	errs = append(errs, validateKindSpecific(obj.Kind, rawObj)...)
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// ValidateUpdate validates an object submitted for update.
// existingRaw is the current object from the state store.
func (v *Validator) ValidateUpdate(kind string, rawObj, existingRaw []byte) error {
	obj, errs := parseAndValidateCommon(rawObj)
	if len(errs) > 0 {
		return errs
	}

	// Validate immutable fields against the existing object.
	var existing genericObject
	if err := json.Unmarshal(existingRaw, &existing); err == nil {
		if obj.Kind != existing.Kind {
			errs = append(errs, &ValidationError{Field: "kind", Reason: "field is immutable"})
		}
		if obj.Metadata.Name != existing.Metadata.Name {
			errs = append(errs, &ValidationError{Field: "metadata.name", Reason: "field is immutable"})
		}
		if obj.Metadata.Namespace != existing.Metadata.Namespace {
			errs = append(errs, &ValidationError{Field: "metadata.namespace", Reason: "field is immutable"})
		}
	}

	errs = append(errs, validateKindSpecific(obj.Kind, rawObj)...)
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// ─── Common validation ────────────────────────────────────────────────────────

func parseAndValidateCommon(rawObj []byte) (genericObject, ValidationErrors) {
	var obj genericObject
	var errs ValidationErrors

	if err := json.Unmarshal(rawObj, &obj); err != nil {
		errs = append(errs, &ValidationError{Field: "<root>", Reason: "invalid JSON: " + err.Error()})
		return obj, errs
	}

	// apiVersion.
	if obj.APIVersion == "" {
		errs = append(errs, &ValidationError{Field: "apiVersion", Reason: "required field missing"})
	}

	// kind.
	if obj.Kind == "" {
		errs = append(errs, &ValidationError{Field: "kind", Reason: "required field missing"})
	}

	// metadata.name or metadata.generateName.
	if obj.Metadata.Name == "" && obj.Metadata.GenerateName == "" {
		errs = append(errs, &ValidationError{
			Field:  "metadata.name",
			Reason: "name or generateName is required",
		})
	}
	if obj.Metadata.Name != "" {
		if err := validateName(obj.Metadata.Name); err != nil {
			errs = append(errs, &ValidationError{Field: "metadata.name", Reason: err.Error()})
		}
	}

	// metadata.namespace (if present).
	if obj.Metadata.Namespace != "" {
		if err := validateNamespaceName(obj.Metadata.Namespace); err != nil {
			errs = append(errs, &ValidationError{Field: "metadata.namespace", Reason: err.Error()})
		}
	}

	// labels keys/values.
	for k, v := range obj.Metadata.Labels {
		if err := validateLabelKey(k); err != nil {
			errs = append(errs, &ValidationError{
				Field:  fmt.Sprintf("metadata.labels[%q]", k),
				Reason: err.Error(),
			})
		}
		if len(v) > 63 {
			errs = append(errs, &ValidationError{
				Field:  fmt.Sprintf("metadata.labels[%q]", k),
				Reason: "label value must be 63 characters or fewer",
			})
		}
	}

	return obj, errs
}

// ─── Name validation ──────────────────────────────────────────────────────────

func validateName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("must not be empty")
	}
	if len(name) > 253 {
		return fmt.Errorf("must be 253 characters or fewer, got %d", len(name))
	}
	if !dns1123Subdomain.MatchString(name) {
		return fmt.Errorf("must consist of lowercase alphanumeric characters, '-', or '.', and must start and end with an alphanumeric character (got %q)", name)
	}
	return nil
}

func validateNamespaceName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("must not be empty")
	}
	if len(name) > 63 {
		return fmt.Errorf("must be 63 characters or fewer")
	}
	if !dns1123Label.MatchString(name) {
		return fmt.Errorf("must consist of lowercase alphanumeric characters or '-', and must start and end with an alphanumeric character (got %q)", name)
	}
	return nil
}

func validateLabelKey(key string) error {
	if key == "" {
		return fmt.Errorf("label key must not be empty")
	}
	// Keys may have an optional prefix (e.g., "app.kubernetes.io/name").
	parts := strings.SplitN(key, "/", 2)
	if len(parts) == 2 {
		prefix, name := parts[0], parts[1]
		if len(prefix) > 253 || !dns1123Subdomain.MatchString(prefix) {
			return fmt.Errorf("label prefix %q is not a valid DNS subdomain", prefix)
		}
		if len(name) > 63 {
			return fmt.Errorf("label name segment must be 63 characters or fewer")
		}
		return nil
	}
	if len(key) > 63 {
		return fmt.Errorf("label key must be 63 characters or fewer")
	}
	return nil
}

// ─── Kind-specific validation ─────────────────────────────────────────────────

func validateKindSpecific(kind string, rawObj []byte) ValidationErrors {
	switch kind {
	case "Pod":
		return validatePod(rawObj)
	case "Namespace":
		return validateNamespace(rawObj)
	case "Service":
		return validateService(rawObj)
	case "ConfigMap":
		return validateConfigMap(rawObj)
	case "Deployment", "StatefulSet", "DaemonSet", "ReplicaSet":
		return validateWorkload(kind, rawObj)
	case "Job":
		return validateJob(rawObj)
	case "TarakSecurityPolicy":
		return validateTarakSecurityPolicy(rawObj)
	case "TarakApplication":
		return validateTarakApplication(rawObj)
	default:
		return nil
	}
}

func validatePod(rawObj []byte) ValidationErrors {
	var pod struct {
		Spec struct {
			Containers     []struct{ Name, Image string } `json:"containers"`
			InitContainers []struct{ Name, Image string } `json:"initContainers"`
		} `json:"spec"`
	}
	var errs ValidationErrors
	if err := json.Unmarshal(rawObj, &pod); err != nil {
		return errs
	}
	if len(pod.Spec.Containers) == 0 {
		errs = append(errs, &ValidationError{
			Field:  "spec.containers",
			Reason: "at least one container is required",
		})
	}
	for i, c := range pod.Spec.Containers {
		if c.Name == "" {
			errs = append(errs, &ValidationError{
				Field:  fmt.Sprintf("spec.containers[%d].name", i),
				Reason: "container name is required",
			})
		}
		if c.Image == "" {
			errs = append(errs, &ValidationError{
				Field:  fmt.Sprintf("spec.containers[%d].image", i),
				Reason: "container image is required",
			})
		}
	}
	return errs
}

func validateNamespace(rawObj []byte) ValidationErrors {
	var ns struct {
		Metadata struct{ Name string } `json:"metadata"`
	}
	var errs ValidationErrors
	if err := json.Unmarshal(rawObj, &ns); err != nil {
		return errs
	}
	if err := validateNamespaceName(ns.Metadata.Name); err != nil {
		errs = append(errs, &ValidationError{Field: "metadata.name", Reason: err.Error()})
	}
	return errs
}

func validateService(rawObj []byte) ValidationErrors {
	var svc struct {
		Spec struct {
			Type  string `json:"type"`
			Ports []struct {
				Port     int    `json:"port"`
				Protocol string `json:"protocol"`
			} `json:"ports"`
		} `json:"spec"`
	}
	var errs ValidationErrors
	if err := json.Unmarshal(rawObj, &svc); err != nil {
		return errs
	}
	validTypes := map[string]bool{"": true, "ClusterIP": true, "NodePort": true, "LoadBalancer": true, "ExternalName": true}
	if !validTypes[svc.Spec.Type] {
		errs = append(errs, &ValidationError{
			Field:  "spec.type",
			Reason: fmt.Sprintf("unknown service type %q", svc.Spec.Type),
		})
	}
	for i, p := range svc.Spec.Ports {
		if p.Port <= 0 || p.Port > 65535 {
			errs = append(errs, &ValidationError{
				Field:  fmt.Sprintf("spec.ports[%d].port", i),
				Reason: "port must be between 1 and 65535",
			})
		}
	}
	return errs
}

func validateConfigMap(rawObj []byte) ValidationErrors {
	var cm struct {
		Data       map[string]interface{} `json:"data"`
		BinaryData map[string]interface{} `json:"binaryData"`
	}
	var errs ValidationErrors
	if err := json.Unmarshal(rawObj, &cm); err != nil {
		return errs
	}
	// Total key count limit (Kubernetes limit).
	total := len(cm.Data) + len(cm.BinaryData)
	if total > 1000000 {
		errs = append(errs, &ValidationError{
			Field:  "data",
			Reason: "too many keys in configmap",
		})
	}
	return errs
}

func validateWorkload(kind string, rawObj []byte) ValidationErrors {
	var w struct {
		Spec struct {
			Replicas *int32 `json:"replicas"`
			Selector *struct {
				MatchLabels map[string]string `json:"matchLabels"`
			} `json:"selector"`
			Template *struct {
				Spec *struct {
					Containers []struct{ Name, Image string } `json:"containers"`
				} `json:"spec"`
			} `json:"template"`
		} `json:"spec"`
	}
	var errs ValidationErrors
	if err := json.Unmarshal(rawObj, &w); err != nil {
		return errs
	}
	if w.Spec.Selector == nil {
		errs = append(errs, &ValidationError{
			Field:  "spec.selector",
			Reason: "selector is required for " + kind,
		})
	}
	if w.Spec.Template == nil {
		errs = append(errs, &ValidationError{
			Field:  "spec.template",
			Reason: "pod template is required for " + kind,
		})
	} else if w.Spec.Template.Spec != nil && len(w.Spec.Template.Spec.Containers) == 0 {
		errs = append(errs, &ValidationError{
			Field:  "spec.template.spec.containers",
			Reason: "at least one container is required",
		})
	}
	if w.Spec.Replicas != nil && *w.Spec.Replicas < 0 {
		errs = append(errs, &ValidationError{
			Field:  "spec.replicas",
			Reason: "replicas must be non-negative",
		})
	}
	return errs
}

func validateJob(rawObj []byte) ValidationErrors {
	var job struct {
		Spec struct {
			Template *struct {
				Spec *struct {
					Containers []struct{ Name, Image string } `json:"containers"`
				} `json:"spec"`
			} `json:"template"`
			Completions  *int32 `json:"completions"`
			Parallelism  *int32 `json:"parallelism"`
		} `json:"spec"`
	}
	var errs ValidationErrors
	if err := json.Unmarshal(rawObj, &job); err != nil {
		return errs
	}
	if job.Spec.Template == nil {
		errs = append(errs, &ValidationError{
			Field:  "spec.template",
			Reason: "pod template is required for Job",
		})
	}
	if job.Spec.Completions != nil && *job.Spec.Completions < 0 {
		errs = append(errs, &ValidationError{Field: "spec.completions", Reason: "must be non-negative"})
	}
	if job.Spec.Parallelism != nil && *job.Spec.Parallelism < 0 {
		errs = append(errs, &ValidationError{Field: "spec.parallelism", Reason: "must be non-negative"})
	}
	return errs
}

func validateTarakSecurityPolicy(rawObj []byte) ValidationErrors {
	var tsp struct {
		Spec struct {
			RunAsUser *int64 `json:"runAsUser"`
		} `json:"spec"`
	}
	var errs ValidationErrors
	if err := json.Unmarshal(rawObj, &tsp); err != nil {
		return errs
	}
	if tsp.Spec.RunAsUser != nil && *tsp.Spec.RunAsUser < 0 {
		errs = append(errs, &ValidationError{
			Field:  "spec.runAsUser",
			Reason: "runAsUser must be a non-negative integer",
		})
	}
	return errs
}

func validateTarakApplication(rawObj []byte) ValidationErrors {
	var app struct {
		Spec struct {
			Image    string `json:"image"`
			Port     int    `json:"port"`
			Replicas int32  `json:"replicas"`
		} `json:"spec"`
	}
	var errs ValidationErrors
	if err := json.Unmarshal(rawObj, &app); err != nil {
		return errs
	}
	if app.Spec.Image == "" {
		errs = append(errs, &ValidationError{
			Field:  "spec.image",
			Reason: "image is required for TarakApplication",
		})
	}
	if app.Spec.Port <= 0 || app.Spec.Port > 65535 {
		errs = append(errs, &ValidationError{
			Field:  "spec.port",
			Reason: "port must be between 1 and 65535",
		})
	}
	if app.Spec.Replicas < 0 {
		errs = append(errs, &ValidationError{
			Field:  "spec.replicas",
			Reason: "replicas must be non-negative",
		})
	}
	return errs
}
