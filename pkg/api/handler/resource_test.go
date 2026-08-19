package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/vikukumar/tarak/internal/statestore"
	"github.com/vikukumar/tarak/pkg/api/handler"
	"github.com/vikukumar/tarak/pkg/api/watch"
)

func setupTestRouter(t *testing.T) (chi.Router, statestore.Store) {
	dir := t.TempDir()
	store, err := statestore.Open(statestore.Options{
		Path:   filepath.Join(dir, "test.db"),
		Logger: zap.NewNop(),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = store.Close()
	})

	wh := watch.New(watch.Options{
		Store: store,
		Log:   zap.NewNop(),
	})

	desc := handler.ResourceDescriptor{
		Group:      "",
		Version:    "v1",
		Resource:   "pods",
		Kind:       "Pod",
		Namespaced: true,
		Verbs:      []string{"create", "delete", "get", "list", "patch", "update", "watch"},
	}

	h := handler.NewResourceHandler(desc, store, wh, zap.NewNop())

	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		h.RegisterRoutes(r)
	})

	return r, store
}

func testPodPayload(name, ns string, labels map[string]string) []byte {
	pod := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": ns,
			"labels":    labels,
		},
		"spec": map[string]interface{}{
			"containers": []map[string]interface{}{
				{
					"name":  "nginx",
					"image": "nginx:1.25",
				},
			},
		},
	}
	raw, _ := json.Marshal(pod)
	return raw
}

func TestResourceHandler_Create_Success(t *testing.T) {
	r, _ := setupTestRouter(t)

	payload := testPodPayload("nginx-demo", "default", map[string]string{"app": "web"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces/default/pods", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	meta := resp["metadata"].(map[string]interface{})
	assert.Equal(t, "nginx-demo", meta["name"])
	assert.Equal(t, "default", meta["namespace"])
	assert.NotEmpty(t, meta["uid"])
	assert.NotEmpty(t, meta["resourceVersion"])
}

func TestResourceHandler_Create_InvalidAdmission(t *testing.T) {
	r, _ := setupTestRouter(t)

	// Missing container image -> should fail admission validation with 422 Unprocessable Entity
	badPayload := []byte(`{
		"apiVersion": "v1",
		"kind": "Pod",
		"metadata": {"name": "invalid-pod", "namespace": "default"},
		"spec": {"containers": [{"name": "no-image"}]}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces/default/pods", bytes.NewReader(badPayload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestResourceHandler_Get_And_NotFound(t *testing.T) {
	r, _ := setupTestRouter(t)

	// Create a pod first
	payload := testPodPayload("nginx-get", "default", nil)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces/default/pods", bytes.NewReader(payload))
	createRec := httptest.NewRecorder()
	r.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	// GET existing
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/default/pods/nginx-get", nil)
	getRec := httptest.NewRecorder()
	r.ServeHTTP(getRec, getReq)
	assert.Equal(t, http.StatusOK, getRec.Code)

	// GET non-existent
	notFoundReq := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/default/pods/does-not-exist", nil)
	notFoundRec := httptest.NewRecorder()
	r.ServeHTTP(notFoundRec, notFoundReq)
	assert.Equal(t, http.StatusNotFound, notFoundRec.Code)
}

func TestResourceHandler_List_WithSelector(t *testing.T) {
	r, _ := setupTestRouter(t)

	// Create pod 1: env=prod
	p1 := testPodPayload("pod-prod", "default", map[string]string{"env": "prod"})
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces/default/pods", bytes.NewReader(p1))
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)
	require.Equal(t, http.StatusCreated, rec1.Code)

	// Create pod 2: env=dev
	p2 := testPodPayload("pod-dev", "default", map[string]string{"env": "dev"})
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces/default/pods", bytes.NewReader(p2))
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	require.Equal(t, http.StatusCreated, rec2.Code)

	// List with labelSelector=env=prod
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/default/pods?labelSelector=env%3Dprod", nil)
	listRec := httptest.NewRecorder()
	r.ServeHTTP(listRec, listReq)
	assert.Equal(t, http.StatusOK, listRec.Code)

	var listResp struct {
		Kind  string                   `json:"kind"`
		Items []map[string]interface{} `json:"items"`
	}
	err := json.Unmarshal(listRec.Body.Bytes(), &listResp)
	require.NoError(t, err)
	assert.Equal(t, "PodList", listResp.Kind)
	require.Len(t, listResp.Items, 1)
	assert.Equal(t, "pod-prod", listResp.Items[0]["metadata"].(map[string]interface{})["name"])
}

func TestResourceHandler_Patch(t *testing.T) {
	r, _ := setupTestRouter(t)

	// Create pod
	payload := testPodPayload("nginx-patch", "default", map[string]string{"version": "1.0"})
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces/default/pods", bytes.NewReader(payload))
	createRec := httptest.NewRecorder()
	r.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	// Patch labels
	patchData := []byte(`{"metadata": {"labels": {"version": "2.0", "tier": "backend"}}}`)
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/namespaces/default/pods/nginx-patch", bytes.NewReader(patchData))
	patchReq.Header.Set("Content-Type", "application/merge-patch+json")
	patchRec := httptest.NewRecorder()
	r.ServeHTTP(patchRec, patchReq)
	assert.Equal(t, http.StatusOK, patchRec.Code)

	// Verify updated labels
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/default/pods/nginx-patch", nil)
	getRec := httptest.NewRecorder()
	r.ServeHTTP(getRec, getReq)
	assert.Equal(t, http.StatusOK, getRec.Code)

	var got map[string]interface{}
	_ = json.Unmarshal(getRec.Body.Bytes(), &got)
	labels := got["metadata"].(map[string]interface{})["labels"].(map[string]interface{})
	assert.Equal(t, "2.0", labels["version"])
	assert.Equal(t, "backend", labels["tier"])
}

func TestResourceHandler_Delete(t *testing.T) {
	r, store := setupTestRouter(t)

	// Create pod
	payload := testPodPayload("nginx-del", "default", nil)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces/default/pods", bytes.NewReader(payload))
	createRec := httptest.NewRecorder()
	r.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	// Delete
	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/namespaces/default/pods/nginx-del", nil)
	delRec := httptest.NewRecorder()
	r.ServeHTTP(delRec, delReq)
	assert.Equal(t, http.StatusOK, delRec.Code)

	// Verify store state
	key := statestore.ResourceKey{Version: "v1", Resource: "pods", Namespace: "default", Name: "nginx-del"}
	_, err := store.Get(context.Background(), key)
	assert.Error(t, err)
}
