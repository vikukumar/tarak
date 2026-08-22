package ingress

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// Route maps a host and path pattern to a target backend URL.
type Route struct {
	Host        string
	Path        string
	BackendURL  *url.URL
	ServiceName string
	ServicePort int
	Namespace   string
}

// Router provides high-speed dynamic HTTP reverse proxying for Ingress rules.
type Router struct {
	mu      sync.RWMutex
	routes  []Route
	proxies map[string]*httputil.ReverseProxy
	log     *zap.Logger
}

// NewRouter creates a new Ingress HTTP router.
func NewRouter(log *zap.Logger) *Router {
	if log == nil {
		log = zap.NewNop()
	}
	return &Router{
		proxies: make(map[string]*httputil.ReverseProxy),
		log:     log.Named("ingress-router"),
	}
}

// UpdateRoutes atomically replaces the active routing table.
func (r *Router) UpdateRoutes(routes []Route) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.routes = routes
	r.proxies = make(map[string]*httputil.ReverseProxy)

	for _, rt := range routes {
		target := rt.BackendURL
		targetStr := target.String()
		if _, exists := r.proxies[targetStr]; !exists {
			proxy := httputil.NewSingleHostReverseProxy(target)
			proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
				r.log.Warn("ingress proxy error",
					zap.String("host", req.Host),
					zap.String("path", req.URL.Path),
					zap.String("target", targetStr),
					zap.Error(err),
				)
				http.Error(w, fmt.Sprintf("502 Bad Gateway (Tarak Ingress): %v", err), http.StatusBadGateway)
			}
			r.proxies[targetStr] = proxy
		}
	}

	r.log.Debug("ingress routes updated", zap.Int("totalRoutes", len(routes)))
}

// ServeHTTP routes incoming HTTP requests to the matching backend proxy.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	host := req.Host
	if colon := strings.IndexByte(host, ':'); colon != -1 {
		host = host[:colon]
	}
	path := req.URL.Path

	var matchedProxy *httputil.ReverseProxy

	for _, rt := range r.routes {
		// Host matching (exact or wildcard)
		hostMatch := rt.Host == "" || rt.Host == "*" || strings.EqualFold(rt.Host, host)
		if !hostMatch && strings.HasPrefix(rt.Host, "*.") {
			suffix := rt.Host[1:]
			hostMatch = strings.HasSuffix(strings.ToLower(host), strings.ToLower(suffix))
		}

		if hostMatch {
			// Path matching (prefix match)
			if rt.Path == "" || rt.Path == "/" || strings.HasPrefix(path, rt.Path) {
				matchedProxy = r.proxies[rt.BackendURL.String()]
				break
			}
		}
	}

	if matchedProxy != nil {
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Tarak-Ingress", "true")
		matchedProxy.ServeHTTP(w, req)
		return
	}

	// Default fallback response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"No Ingress backend matched for host %q and path %q","status":404}`, host, path)))
}

// Routes returns a snapshot of active routes.
func (r *Router) Routes() []Route {
	r.mu.RLock()
	defer r.mu.RUnlock()
	copied := make([]Route, len(r.routes))
	copy(copied, r.routes)
	return copied
}
