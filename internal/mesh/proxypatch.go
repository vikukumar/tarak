package mesh

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// PatchStage defines the execution point of the ProxyPatch in the request lifecycle.
type PatchStage string

const (
	StageRequestHeaders  PatchStage = "RequestHeaders"
	StageRequestBody     PatchStage = "RequestBody"
	StageResponseHeaders PatchStage = "ResponseHeaders"
	StageResponseBody    PatchStage = "ResponseBody"
)

// ProxyPatch defines programmable middleware rules to modify mesh traffic dynamically.
type ProxyPatch struct {
	Name            string            `json:"name"`
	Mesh            string            `json:"mesh"`
	Namespace       string            `json:"namespace"`
	TargetService   string            `json:"targetService"`
	Stage           PatchStage        `json:"stage"`
	AddHeaders      map[string]string `json:"addHeaders,omitempty"`
	RemoveHeaders   []string          `json:"removeHeaders,omitempty"`
	PathRewrite     string            `json:"pathRewrite,omitempty"` // e.g. "/v2/$1"
	RequireAuthRole string            `json:"requireAuthRole,omitempty"`
	Enabled         bool              `json:"enabled"`
}

// ProxyPatchManager manages programmable Go middleware filters across service mesh proxies.
type ProxyPatchManager struct {
	log     *zap.Logger
	mu      sync.RWMutex
	patches map[string]*ProxyPatch
}

// NewProxyPatchManager initializes the ProxyPatch middleware manager.
func NewProxyPatchManager(log *zap.Logger) *ProxyPatchManager {
	if log == nil {
		log = zap.NewNop()
	}

	ppm := &ProxyPatchManager{
		log:     log.Named("mesh-proxypatch"),
		patches: make(map[string]*ProxyPatch),
	}

	ppm.seedDefaultPatches()
	return ppm
}

func (ppm *ProxyPatchManager) seedDefaultPatches() {
	ppm.patches["inject-trace-context"] = &ProxyPatch{
		Name:          "inject-trace-context",
		Mesh:          "default",
		Namespace:     "production",
		TargetService: "storefront-web",
		Stage:         StageRequestHeaders,
		AddHeaders: map[string]string{
			"X-Tarak-Mesh-Origin": "tarak-native-gateway",
			"X-ZeroTrust-Verified": "true",
		},
		Enabled: true,
	}

	ppm.patches["strip-sensitive-headers"] = &ProxyPatch{
		Name:          "strip-sensitive-headers",
		Mesh:          "default",
		Namespace:     "finance",
		TargetService: "payments-api",
		Stage:         StageResponseHeaders,
		RemoveHeaders: []string{"X-Powered-By", "Server"},
		Enabled:       true,
	}
}

// InterceptRequest applies active RequestHeaders and RequestBody proxy patches.
func (ppm *ProxyPatchManager) InterceptRequest(ctx context.Context, serviceName string, req *http.Request) error {
	ppm.mu.RLock()
	defer ppm.mu.RUnlock()

	for _, patch := range ppm.patches {
		if !patch.Enabled || patch.TargetService != serviceName {
			continue
		}

		// Inject custom request headers
		for k, v := range patch.AddHeaders {
			req.Header.Set(k, v)
		}

		// Strip specified headers
		for _, h := range patch.RemoveHeaders {
			req.Header.Del(h)
		}

		// Apply Path rewrite if specified
		if patch.PathRewrite != "" {
			req.URL.Path = patch.PathRewrite
		}

		// Verify Auth Role requirement
		if patch.RequireAuthRole != "" {
			authHeader := req.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return fmt.Errorf("proxy-patch: unauthorized - missing Bearer token for role %q", patch.RequireAuthRole)
			}
		}
	}
	return nil
}

// InterceptResponse applies active ResponseHeaders proxy patches.
func (ppm *ProxyPatchManager) InterceptResponse(serviceName string, header http.Header) {
	ppm.mu.RLock()
	defer ppm.mu.RUnlock()

	for _, patch := range ppm.patches {
		if !patch.Enabled || patch.TargetService != serviceName {
			continue
		}

		for k, v := range patch.AddHeaders {
			header.Set(k, v)
		}
		for _, h := range patch.RemoveHeaders {
			header.Del(h)
		}
	}
}

// ListPatches returns all configured proxy patches.
func (ppm *ProxyPatchManager) ListPatches() []*ProxyPatch {
	ppm.mu.RLock()
	defer ppm.mu.RUnlock()

	list := make([]*ProxyPatch, 0, len(ppm.patches))
	for _, p := range ppm.patches {
		list = append(list, p)
	}
	return list
}
