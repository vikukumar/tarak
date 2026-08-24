// Package handler implements the core Kubernetes-compatible REST handlers
// for the Tarak API server.
//
// A ResourceHandler provides the full CRUD + watch surface for a single
// resource type (e.g., pods, deployments) and dispatches to the state store.
//
// Request routing:
//
//	POST   /api/v1/namespaces/{ns}/{resource}              → Create
//	GET    /api/v1/namespaces/{ns}/{resource}              → List  (+ ?watch=true → Watch)
//	GET    /api/v1/namespaces/{ns}/{resource}/{name}       → Get
//	PUT    /api/v1/namespaces/{ns}/{resource}/{name}       → Update
//	PATCH  /api/v1/namespaces/{ns}/{resource}/{name}       → Patch
//	DELETE /api/v1/namespaces/{ns}/{resource}/{name}       → Delete
//	PUT    /api/v1/namespaces/{ns}/{resource}/{name}/status → StatusUpdate
//
// Cluster-scoped resources omit the /namespaces/{ns} segment.
//
// All responses are JSON with the same wire format as Kubernetes so that
// kubectl, Helm, and other Kubernetes tooling works without modification.
package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/vikukumar/tarak/internal/statestore"
	"github.com/vikukumar/tarak/pkg/api/admission"
	"github.com/vikukumar/tarak/pkg/api/watch"
	lsels "github.com/vikukumar/tarak/api/meta"
)

// ResourceDescriptor describes a single resource type registered with the API server.
type ResourceDescriptor struct {
	// Group is the API group ("" for core, "apps" for apps/v1).
	Group string
	// Version is the API version ("v1", "v1beta1").
	Version string
	// Resource is the plural resource name ("pods", "deployments").
	Resource string
	// Kind is the singular Kind ("Pod", "Deployment").
	Kind string
	// Namespaced is true for namespaced resources, false for cluster-scoped.
	Namespaced bool
	// ShortNames are alternative names (e.g., "po" for pods).
	ShortNames []string
	// Verbs is the list of supported verbs.
	Verbs []string
}

// ResourceHandler handles REST requests for a single resource type.
type ResourceHandler struct {
	desc      ResourceDescriptor
	store     statestore.Store
	watcher   *watch.Handler
	validator *admission.Validator
	log       *zap.Logger
}

// NewResourceHandler creates a ResourceHandler.
func NewResourceHandler(desc ResourceDescriptor, store statestore.Store, wh *watch.Handler, log *zap.Logger) *ResourceHandler {
	if log == nil {
		log = zap.NewNop()
	}
	return &ResourceHandler{
		desc:      desc,
		store:     store,
		watcher:   wh,
		validator: admission.New(),
		log:       log,
	}
}

// ─── Route registration ───────────────────────────────────────────────────────

// RegisterRoutes registers all routes for this resource onto the given chi router.
func (h *ResourceHandler) RegisterRoutes(r chi.Router) {
	if h.desc.Namespaced {
		r.Post(fmt.Sprintf("/namespaces/{namespace}/%s", h.desc.Resource), h.Create)
		r.Get(fmt.Sprintf("/namespaces/{namespace}/%s", h.desc.Resource), h.ListOrWatch)
		r.Get(fmt.Sprintf("/namespaces/{namespace}/%s/{name}", h.desc.Resource), h.Get)
		r.Put(fmt.Sprintf("/namespaces/{namespace}/%s/{name}", h.desc.Resource), h.Update)
		r.Patch(fmt.Sprintf("/namespaces/{namespace}/%s/{name}", h.desc.Resource), h.Patch)
		r.Delete(fmt.Sprintf("/namespaces/{namespace}/%s/{name}", h.desc.Resource), h.Delete)
		r.Put(fmt.Sprintf("/namespaces/{namespace}/%s/{name}/status", h.desc.Resource), h.StatusUpdate)
		r.Get(fmt.Sprintf("/namespaces/{namespace}/%s/{name}/status", h.desc.Resource), h.GetStatus)
		// Also register namespace-less routes for watching all namespaces.
		r.Get(fmt.Sprintf("/%s", h.desc.Resource), h.ListOrWatch)
	} else {
		r.Post(fmt.Sprintf("/%s", h.desc.Resource), h.Create)
		r.Get(fmt.Sprintf("/%s", h.desc.Resource), h.ListOrWatch)
		r.Get(fmt.Sprintf("/%s/{name}", h.desc.Resource), h.Get)
		r.Put(fmt.Sprintf("/%s/{name}", h.desc.Resource), h.Update)
		r.Patch(fmt.Sprintf("/%s/{name}", h.desc.Resource), h.Patch)
		r.Delete(fmt.Sprintf("/%s/{name}", h.desc.Resource), h.Delete)
		r.Put(fmt.Sprintf("/%s/{name}/status", h.desc.Resource), h.StatusUpdate)
		r.Get(fmt.Sprintf("/%s/{name}/status", h.desc.Resource), h.GetStatus)
	}
}

// ─── CRUD handlers ────────────────────────────────────────────────────────────

// Create handles POST .../resource — creates a new resource.
func (h *ResourceHandler) Create(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r, 4<<20) // 4 MiB max
	if err != nil {
		writeError(w, http.StatusBadRequest, "BadRequest", err.Error())
		return
	}

	// Admission validation.
	if err := h.validator.ValidateCreate(h.desc.Kind, body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "Invalid", err.Error())
		return
	}

	key := h.keyFromRequest(r, "")
	if key.Name == "" {
		var meta struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"metadata"`
		}
		if err := json.Unmarshal(body, &meta); err == nil {
			key.Name = meta.Metadata.Name
			if key.Namespace == "" && h.desc.Namespaced && meta.Metadata.Namespace != "" {
				key.Namespace = meta.Metadata.Namespace
			}
		}
	}
	if key.Name == "" {
		writeError(w, http.StatusBadRequest, "BadRequest", "metadata.name is required")
		return
	}

	if h.desc.Namespaced {
		if key.Namespace == "" {
			key.Namespace = "default"
		}
		var objMap map[string]interface{}
		if err := json.Unmarshal(body, &objMap); err == nil {
			meta, _ := objMap["metadata"].(map[string]interface{})
			if meta == nil {
				meta = make(map[string]interface{})
				objMap["metadata"] = meta
			}
			if currNS, _ := meta["namespace"].(string); currNS == "" {
				meta["namespace"] = key.Namespace
				if mutated, mErr := json.Marshal(objMap); mErr == nil {
					body = mutated
				}
			}
		}
	}

	// Service ClusterIP defaulting
	if h.desc.Resource == "services" || h.desc.Kind == "Service" {
		var svcMap map[string]interface{}
		if err := json.Unmarshal(body, &svcMap); err == nil {
			spec, _ := svcMap["spec"].(map[string]interface{})
			if spec != nil {
				sType, _ := spec["type"].(string)
				cIP, _ := spec["clusterIP"].(string)
				if cIP == "" && sType != "ExternalName" {
					hVal := 0
					for _, ch := range key.Name {
						hVal = (hVal*31 + int(ch)) % 240
					}
					spec["clusterIP"] = fmt.Sprintf("10.96.0.%d", hVal+10)
					svcMap["spec"] = spec
					if mutated, mErr := json.Marshal(svcMap); mErr == nil {
						body = mutated
					}
				}
			}
		}
	}

	env, err := h.store.Create(r.Context(), key, body)
	if err != nil {
		h.writeStoreError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(json.RawMessage(env.Object))
}

// ListOrWatch handles GET .../resource — lists resources or starts a watch.
func (h *ResourceHandler) ListOrWatch(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("watch") == "true" {
		h.watch(w, r)
		return
	}

	query := statestore.ListQuery{
		Key:   h.keyFromRequest(r, ""),
		Limit: parseLimitParam(r),
	}

	// Parse label selector.
	if ls := r.URL.Query().Get("labelSelector"); ls != "" {
		sel, err := lsels.ParseLabelSelector(ls)
		if err != nil {
			writeError(w, http.StatusBadRequest, "BadRequest", "invalid labelSelector: "+err.Error())
			return
		}
		query.LabelSelector = sel
	}

	// Parse field selector.
	if fs := r.URL.Query().Get("fieldSelector"); fs != "" {
		sel, err := lsels.ParseFieldSelector(fs)
		if err != nil {
			writeError(w, http.StatusBadRequest, "BadRequest", "invalid fieldSelector: "+err.Error())
			return
		}
		query.FieldSelector = sel
	}

	// Parse continue token.
	query.Continue = r.URL.Query().Get("continue")

	envs, rev, err := h.store.List(r.Context(), query)
	if err != nil {
		h.writeStoreError(w, err)
		return
	}

	// Build a list response in Kubernetes list format.
	items := make([]json.RawMessage, len(envs))
	for i, env := range envs {
		items[i] = env.Object
	}

	listObj := struct {
		APIVersion string            `json:"apiVersion"`
		Kind       string            `json:"kind"`
		Metadata   listMeta          `json:"metadata"`
		Items      []json.RawMessage `json:"items"`
	}{
		APIVersion: apiVersion(h.desc.Group, h.desc.Version),
		Kind:       h.desc.Kind + "List",
		Metadata:   listMeta{ResourceVersion: strconv.FormatInt(rev, 10)},
		Items:      items,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(listObj)
}

// Get handles GET .../resource/{name} — retrieves a single resource.
func (h *ResourceHandler) Get(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "BadRequest", "name is required")
		return
	}

	key := h.keyFromRequest(r, name)
	env, err := h.store.Get(r.Context(), key)
	if err != nil {
		h.writeStoreError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(json.RawMessage(env.Object))
}

// GetStatus handles GET .../resource/{name}/status.
func (h *ResourceHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	h.Get(w, r)
}

// Update handles PUT .../resource/{name} — replaces a resource.
func (h *ResourceHandler) Update(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	body, err := readBody(r, 4<<20)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BadRequest", err.Error())
		return
	}

	key := h.keyFromRequest(r, name)

	// Fetch existing object to validate updates against immutable fields.
	existing, err := h.store.Get(r.Context(), key)
	if err != nil {
		h.writeStoreError(w, err)
		return
	}

	// Admission validation.
	if err := h.validator.ValidateUpdate(h.desc.Kind, body, existing.Object); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "Invalid", err.Error())
		return
	}

	rv := parseResourceVersion(r)
	env, err := h.store.Update(r.Context(), key, body, rv)
	if err != nil {
		h.writeStoreError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(json.RawMessage(env.Object))
}

// StatusUpdate handles PUT .../resource/{name}/status — updates only the status subresource.
func (h *ResourceHandler) StatusUpdate(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	body, err := readBody(r, 4<<20)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BadRequest", err.Error())
		return
	}

	rv := parseResourceVersion(r)
	key := h.keyFromRequest(r, name)
	env, err := h.store.StatusUpdate(r.Context(), key, body, rv)
	if err != nil {
		h.writeStoreError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(json.RawMessage(env.Object))
}

// Patch handles PATCH .../resource/{name} — applies a strategic-merge or JSON-merge patch.
func (h *ResourceHandler) Patch(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	patchType := r.Header.Get("Content-Type")

	body, err := readBody(r, 4<<20)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BadRequest", err.Error())
		return
	}

	key := h.keyFromRequest(r, name)

	// Get the existing object.
	existing, err := h.store.Get(r.Context(), key)
	if err != nil {
		h.writeStoreError(w, err)
		return
	}

	// Apply the patch.
	patched, err := applyPatch(existing.Object, body, patchType)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "Invalid", "patch failed: "+err.Error())
		return
	}

	// Store the patched object.
	env, err := h.store.Update(r.Context(), key, patched, existing.ResourceVersion)
	if err != nil {
		h.writeStoreError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(json.RawMessage(env.Object))
}

// Delete handles DELETE .../resource/{name} — removes a resource.
func (h *ResourceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "BadRequest", "name is required")
		return
	}

	key := h.keyFromRequest(r, name)
	rv := parseResourceVersion(r)

	_, err := h.store.Delete(r.Context(), key, rv)
	if err != nil {
		h.writeStoreError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeStatusOK(w, "deleted")
}

// ─── Watch ────────────────────────────────────────────────────────────────────

func (h *ResourceHandler) watch(w http.ResponseWriter, r *http.Request) {
	key := h.keyFromRequest(r, chi.URLParam(r, "name"))

	var sinceRev int64
	if rvStr := r.URL.Query().Get("resourceVersion"); rvStr != "" {
		sinceRev, _ = strconv.ParseInt(rvStr, 10, 64)
	}

	query := statestore.WatchQuery{
		Key:           key,
		SinceRevision: sinceRev,
		SendBookmarks: r.URL.Query().Get("allowWatchBookmarks") == "true",
	}

	h.watcher.ServeWatch(w, r, query)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// keyFromRequest builds a ResourceKey from the request path and name.
func (h *ResourceHandler) keyFromRequest(r *http.Request, name string) statestore.ResourceKey {
	ns := chi.URLParam(r, "namespace")
	if n := chi.URLParam(r, "name"); n != "" && name == "" {
		name = n
	}
	return statestore.ResourceKey{
		Group:     h.desc.Group,
		Version:   h.desc.Version,
		Resource:  h.desc.Resource,
		Namespace: ns,
		Name:      name,
	}
}

func parseLimitParam(r *http.Request) int64 {
	if l := r.URL.Query().Get("limit"); l != "" {
		n, err := strconv.ParseInt(l, 10, 64)
		if err == nil && n > 0 {
			return n
		}
	}
	return 0
}

func parseResourceVersion(r *http.Request) int64 {
	// Try body first for DELETE with options, then URL param.
	if rv := r.URL.Query().Get("resourceVersion"); rv != "" {
		n, _ := strconv.ParseInt(rv, 10, 64)
		return n
	}
	return 0
}

func readBody(r *http.Request, maxBytes int64) ([]byte, error) {
	r.Body = http.MaxBytesReader(nil, r.Body, maxBytes)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("read request body: %w", err)
	}
	return data, nil
}

func apiVersion(group, version string) string {
	if group == "" {
		return version
	}
	return group + "/" + version
}

type listMeta struct {
	ResourceVersion string `json:"resourceVersion"`
	Continue        string `json:"continue,omitempty"`
}

// writeError writes a Kubernetes-style Status JSON error.
func writeError(w http.ResponseWriter, code int, reason, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	status := struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
		Status     string `json:"status"`
		Message    string `json:"message"`
		Reason     string `json:"reason"`
		Code       int    `json:"code"`
	}{
		APIVersion: "v1",
		Kind:       "Status",
		Status:     "Failure",
		Message:    message,
		Reason:     reason,
		Code:       code,
	}
	_ = json.NewEncoder(w).Encode(status)
}

// writeStatusOK writes a minimal Status OK JSON response.
func writeStatusOK(w http.ResponseWriter, reason string) {
	status := struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
		Status     string `json:"status"`
		Reason     string `json:"reason"`
		Code       int    `json:"code"`
	}{
		APIVersion: "v1",
		Kind:       "Status",
		Status:     "Success",
		Reason:     reason,
		Code:       http.StatusOK,
	}
	_ = json.NewEncoder(w).Encode(status)
}

// writeStoreError converts a statestore error to an HTTP error response.
func (h *ResourceHandler) writeStoreError(w http.ResponseWriter, err error) {
	switch {
	case isNotFound(err):
		writeError(w, http.StatusNotFound, "NotFound", err.Error())
	case isAlreadyExists(err):
		writeError(w, http.StatusConflict, "AlreadyExists", err.Error())
	case isConflict(err):
		writeError(w, http.StatusConflict, "Conflict", err.Error())
	case isKeyInvalid(err):
		writeError(w, http.StatusBadRequest, "BadRequest", err.Error())
	default:
		h.log.Error("store error", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "InternalError", err.Error())
	}
}

func isNotFound(err error) bool {
	_, ok := err.(*statestore.NotFoundError)
	return ok
}

func isAlreadyExists(err error) bool {
	_, ok := err.(*statestore.AlreadyExistsError)
	return ok
}

func isConflict(err error) bool {
	_, ok := err.(*statestore.ConflictError)
	return ok
}

func isKeyInvalid(err error) bool {
	return strings.Contains(err.Error(), "invalid resource key")
}

// ─── Patch ────────────────────────────────────────────────────────────────────

// applyPatch applies a JSON-merge patch (RFC 7396) or strategic-merge patch to the existing object.
// Phase 1 implements JSON-merge patch only. Strategic merge patch will be Phase 5.
func applyPatch(existing, patch []byte, patchType string) ([]byte, error) {
	switch {
	case strings.Contains(patchType, "application/merge-patch+json"),
		strings.Contains(patchType, "application/strategic-merge-patch+json"),
		strings.Contains(patchType, "application/json"):
		return jsonMergePatch(existing, patch)
	default:
		// Default to merge patch for unknown types.
		return jsonMergePatch(existing, patch)
	}
}

// jsonMergePatch applies a JSON-merge patch (RFC 7396).
// Values in patch override corresponding values in target.
// Null values in patch delete the key from target.
func jsonMergePatch(target, patch []byte) ([]byte, error) {
	var targetMap map[string]interface{}
	if err := json.Unmarshal(target, &targetMap); err != nil {
		return nil, fmt.Errorf("parse target: %w", err)
	}
	var patchMap map[string]interface{}
	if err := json.Unmarshal(patch, &patchMap); err != nil {
		return nil, fmt.Errorf("parse patch: %w", err)
	}
	result := mergeMaps(targetMap, patchMap)
	return json.Marshal(result)
}

// mergeMaps recursively merges src into dst (RFC 7396 semantics).
func mergeMaps(dst, src map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(dst))
	for k, v := range dst {
		result[k] = v
	}
	for k, v := range src {
		if v == nil {
			delete(result, k)
			continue
		}
		if srcMap, ok := v.(map[string]interface{}); ok {
			if dstMap, ok := result[k].(map[string]interface{}); ok {
				result[k] = mergeMaps(dstMap, srcMap)
				continue
			}
		}
		result[k] = v
	}
	return result
}
