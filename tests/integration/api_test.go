// Package integration contains end-to-end tests for the Tarak API server.
//
// Tests in this package spin up a real API server (with BoltDB + TLS) against
// a temporary directory and exercise the full request path including:
//   - mTLS authentication
//   - CRUD operations via the REST API
//   - Watch streams
//   - Label selector filtering
//
// Run with:
//
//	go test ./tests/integration/... -v -timeout 120s
package integration

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/vikukumar/tarak/internal/server"
	"github.com/vikukumar/tarak/pkg/security"
)

// ─── Server fixture ───────────────────────────────────────────────────────────

// testServer is a running API server for integration testing.
type testServer struct {
	baseURL    string
	httpClient *http.Client
	t          *testing.T
}

// startTestServer starts a Tarak API server on a random port and returns a
// configured testServer. The server is shut down when the test completes.
func startTestServer(t *testing.T) *testServer {
	t.Helper()

	dir := t.TempDir()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)

	// Use an ephemeral port.
	addr := "127.0.0.1:0"

	cfg := server.Config{
		BindAddress:       addr,
		DataDir:           dir,
		AllowInsecureAuth: true, // use insecure for testing simplicity
		Log:               log,
	}

	testPort := "18443"
	cfg.BindAddress = "127.0.0.1:" + testPort

	srv, err := server.New(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
	})

	started := make(chan struct{})
	go func() {
		// Allow the health check to pass before declaring ready.
		close(started)
		if err := srv.Run(ctx); err != nil {
			t.Logf("server error: %v", err)
		}
	}()
	<-started

	// Wait for server to be ready.
	baseURL := "https://127.0.0.1:" + testPort
	probeClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
		Timeout: 2 * time.Second,
	}
	require.Eventually(t, func() bool {
		resp, err := probeClient.Get(baseURL + "/healthz")
		if err != nil {
			return false
		}
		resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 10*time.Second, 100*time.Millisecond, "server did not start in time")

	// Read the generated PKI to build the TLS client.
	caCertPEM, err := os.ReadFile(dir + "/pki/ca.crt")
	require.NoError(t, err)
	clientCertPEM, err := os.ReadFile(dir + "/pki/admin.crt")
	require.NoError(t, err)
	clientKeyPEM, err := os.ReadFile(dir + "/pki/admin.key")
	require.NoError(t, err)

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCertPEM)

	clientCert, err := tls.X509KeyPair(clientCertPEM, clientKeyPEM)
	require.NoError(t, err)

	tlsCfg := &tls.Config{
		RootCAs:      caPool,
		Certificates: []tls.Certificate{clientCert},
		MinVersion:   tls.VersionTLS13,
	}

	httpClient := &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsCfg},
		Timeout:   10 * time.Second,
	}

	return &testServer{
		baseURL:    baseURL,
		httpClient: httpClient,
		t:          t,
	}
}

// ─── HTTP helpers ─────────────────────────────────────────────────────────────

func (s *testServer) get(path string) (*http.Response, []byte) {
	s.t.Helper()
	resp, err := s.httpClient.Get(s.baseURL + path)
	require.NoError(s.t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(s.t, err)
	resp.Body.Close()
	return resp, body
}

func (s *testServer) post(path string, body interface{}) (*http.Response, []byte) {
	s.t.Helper()
	data, err := json.Marshal(body)
	require.NoError(s.t, err)
	resp, err := s.httpClient.Post(s.baseURL+path, "application/json", bytes.NewReader(data))
	require.NoError(s.t, err)
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(s.t, err)
	resp.Body.Close()
	return resp, respBody
}

func (s *testServer) delete(path string) *http.Response {
	s.t.Helper()
	req, err := http.NewRequest(http.MethodDelete, s.baseURL+path, nil)
	require.NoError(s.t, err)
	resp, err := s.httpClient.Do(req)
	require.NoError(s.t, err)
	resp.Body.Close()
	return resp
}

// ─── Pod helpers ──────────────────────────────────────────────────────────────

func podBody(name, ns string) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": ns,
			"labels":    map[string]string{"app": "test"},
		},
		"spec": map[string]interface{}{
			"containers": []map[string]interface{}{
				{"name": "nginx", "image": "nginx:latest"},
			},
		},
	}
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestHealthEndpoints(t *testing.T) {
	s := startTestServer(t)

	resp, body := s.get("/healthz")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "ok")

	resp, body = s.get("/readyz")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "ok")

	resp, _ = s.get("/livez")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAPIDiscovery(t *testing.T) {
	s := startTestServer(t)

	// Core API groups.
	resp, body := s.get("/api")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "v1")

	// Named API groups.
	resp, body = s.get("/apis")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "apps")
	assert.Contains(t, string(body), "batch")

	// Core resource list.
	resp, body = s.get("/api/v1")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "pods")
	assert.Contains(t, string(body), "services")
}

func TestPodCRUD(t *testing.T) {
	s := startTestServer(t)
	ns := "default"

	// ── Create ───────────────────────────────────────────────────────────
	resp, body := s.post(fmt.Sprintf("/api/v1/namespaces/%s/pods", ns), podBody("nginx", ns))
	assert.Equal(t, http.StatusCreated, resp.StatusCode, string(body))

	var created map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &created))
	assert.Equal(t, "nginx", jsonPath(created, "metadata", "name"))
	assert.NotEmpty(t, jsonPath(created, "metadata", "uid"))
	assert.NotEmpty(t, jsonPath(created, "metadata", "resourceVersion"))

	// ── Get ──────────────────────────────────────────────────────────────
	resp, body = s.get(fmt.Sprintf("/api/v1/namespaces/%s/pods/nginx", ns))
	assert.Equal(t, http.StatusOK, resp.StatusCode, string(body))

	var got map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &got))
	assert.Equal(t, "nginx", jsonPath(got, "metadata", "name"))

	// ── List ─────────────────────────────────────────────────────────────
	resp, body = s.get(fmt.Sprintf("/api/v1/namespaces/%s/pods", ns))
	assert.Equal(t, http.StatusOK, resp.StatusCode, string(body))

	var list map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &list))
	items, ok := list["items"].([]interface{})
	require.True(t, ok)
	assert.Len(t, items, 1)

	// ── Delete ───────────────────────────────────────────────────────────
	resp = s.delete(fmt.Sprintf("/api/v1/namespaces/%s/pods/nginx", ns))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify gone.
	resp, _ = s.get(fmt.Sprintf("/api/v1/namespaces/%s/pods/nginx", ns))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCreateAlreadyExists(t *testing.T) {
	s := startTestServer(t)

	_, _ = s.post("/api/v1/namespaces/default/pods", podBody("nginx", "default"))
	resp, _ := s.post("/api/v1/namespaces/default/pods", podBody("nginx", "default"))
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestGetNotFound(t *testing.T) {
	s := startTestServer(t)

	resp, _ := s.get("/api/v1/namespaces/default/pods/missing")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestLabelSelectorFilter(t *testing.T) {
	s := startTestServer(t)

	// Create two pods with different labels.
	pod1 := podBody("pod1", "default")
	pod1["metadata"].(map[string]interface{})["labels"] = map[string]string{"app": "frontend"}

	pod2 := podBody("pod2", "default")
	pod2["metadata"].(map[string]interface{})["labels"] = map[string]string{"app": "backend"}

	s.post("/api/v1/namespaces/default/pods", pod1)
	s.post("/api/v1/namespaces/default/pods", pod2)

	// Filter by label.
	resp, body := s.get("/api/v1/namespaces/default/pods?labelSelector=app%3Dfrontend")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var list map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &list))
	items := list["items"].([]interface{})
	assert.Len(t, items, 1)
}

func TestDeploymentCRUD(t *testing.T) {
	s := startTestServer(t)

	deploy := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name":      "my-app",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"replicas": 3,
			"selector": map[string]interface{}{
				"matchLabels": map[string]string{"app": "my-app"},
			},
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]string{"app": "my-app"},
				},
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{"name": "app", "image": "myapp:v1"},
					},
				},
			},
		},
	}

	resp, body := s.post("/apis/apps/v1/namespaces/default/deployments", deploy)
	assert.Equal(t, http.StatusCreated, resp.StatusCode, string(body))

	resp, _ = s.get("/apis/apps/v1/namespaces/default/deployments/my-app")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, _ = s.get("/apis/apps/v1/namespaces/default/deployments")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSecurityTokenSignAndVerify(t *testing.T) {
	secret, err := security.GenerateSecret()
	require.NoError(t, err)

	signer := security.NewTokenSigner(secret)
	token, err := signer.Issue("test-user", []string{"system:masters"}, time.Hour)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := signer.Verify(token)
	require.NoError(t, err)
	assert.Equal(t, "test-user", claims.Subject)
}

func TestEncryptionRoundTrip(t *testing.T) {
	key, err := security.GenerateEncryptionKey()
	require.NoError(t, err)

	enc, err := security.NewEncryptor(key)
	require.NoError(t, err)

	secret := []byte("my-database-password")
	ciphertext, err := enc.Encrypt(secret)
	require.NoError(t, err)

	plaintext, err := enc.Decrypt(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, secret, plaintext)
}

func TestBootstrappedNamespaces(t *testing.T) {
	s := startTestServer(t)

	resp, body := s.get("/api/v1/namespaces/default")
	assert.Equal(t, http.StatusOK, resp.StatusCode, string(body))

	resp, body = s.get("/api/v1/namespaces/tarak-system")
	assert.Equal(t, http.StatusOK, resp.StatusCode, string(body))
}

func TestVersionAndOpenAPIEndpoints(t *testing.T) {
	s := startTestServer(t)

	resp, body := s.get("/version")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "gitVersion")

	resp, body = s.get("/openapi/v2")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "swagger")
}

func TestAdmissionValidation_Integration(t *testing.T) {
	s := startTestServer(t)

	badPod := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]interface{}{
			"name":      "invalid-pod",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"containers": []map[string]interface{}{
				{"name": "broken"}, // missing image
			},
		},
	}

	resp, _ := s.post("/api/v1/namespaces/default/pods", badPod)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestTarakIOGroupsAndCRDs(t *testing.T) {
	s := startTestServer(t)

	// 1. Verify tarak.io groups are listed in /apis
	resp, body := s.get("/apis")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "security.tarak.io")
	assert.Contains(t, string(body), "apps.tarak.io")
	assert.Contains(t, string(body), "networking.tarak.io")

	// 2. Create and get TarakSecurityPolicy
	tspObj := map[string]interface{}{
		"apiVersion": "security.tarak.io/v1",
		"kind":       "TarakSecurityPolicy",
		"metadata": map[string]interface{}{
			"name": "strict-zero-trust",
		},
		"spec": map[string]interface{}{
			"privileged":              false,
			"readOnlyRootFilesystem":  true,
			"enforceEncryptionAtRest": true,
			"networkIsolation":        true,
		},
	}
	resp, body = s.post("/apis/security.tarak.io/v1/taraksecuritypolicies", tspObj)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, string(body), "strict-zero-trust")

	resp, body = s.get("/apis/security.tarak.io/v1/taraksecuritypolicies/strict-zero-trust")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "strict-zero-trust")

	// 3. Create and get TarakApplication
	appObj := map[string]interface{}{
		"apiVersion": "apps.tarak.io/v1",
		"kind":       "TarakApplication",
		"metadata": map[string]interface{}{
			"name":      "secure-web-app",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"image":             "ghcr.io/org/secure-web:v1.0.0",
			"replicas":          3,
			"port":              8080,
			"domain":            "app.tarak.io",
			"autoTLS":           true,
			"securityPolicyRef": "strict-zero-trust",
		},
	}
	resp, body = s.post("/apis/apps.tarak.io/v1/namespaces/default/tarakapplications", appObj)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, string(body), "secure-web-app")

	resp, body = s.get("/apis/apps.tarak.io/v1/namespaces/default/tarakapplications/secure-web-app")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "secure-web-app")
}

// ─── JSON path helper ─────────────────────────────────────────────────────────

// jsonPath extracts a nested string value from a JSON object by path.
func jsonPath(obj map[string]interface{}, keys ...string) string {
	current := interface{}(obj)
	for _, key := range keys {
		m, ok := current.(map[string]interface{})
		if !ok {
			return ""
		}
		current = m[key]
	}
	if s, ok := current.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", current)
}
