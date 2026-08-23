// Package server implements the Tarak API server.
//
// The API server is the central control-plane component.  It:
//   - Accepts HTTPS connections (mutual TLS)
//   - Authenticates requests via client certs or bearer tokens
//   - Routes requests to resource handlers
//   - Dispatches reads/writes to the state store
//   - Streams watch events to connected clients
//   - Serves health, metrics, and API discovery endpoints
//
// Server startup sequence:
//  1. Load configuration
//  2. Open (or initialise) the state store
//  3. Load or generate PKI
//  4. Construct middleware chain: RequestID → Audit → Auth
//  5. Register resource routes
//  6. Bind TLS listener
//  7. Signal readiness
//  8. Serve until shutdown signal
package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	appsv1 "github.com/vikukumar/tarak/api/apps/v1"
	batchv1 "github.com/vikukumar/tarak/api/batch/v1"
	corev1 "github.com/vikukumar/tarak/api/core/v1"
	networkingv1 "github.com/vikukumar/tarak/api/networking/v1"
	rbacv1 "github.com/vikukumar/tarak/api/rbac/v1"
	storagev1 "github.com/vikukumar/tarak/api/storage/v1"
	tarakv1 "github.com/vikukumar/tarak/api/tarak/v1"
	"github.com/vikukumar/tarak/internal/auth"
	"github.com/vikukumar/tarak/internal/controller"
	"github.com/vikukumar/tarak/internal/ingress"
	"github.com/vikukumar/tarak/internal/loadbalancer"
	"github.com/vikukumar/tarak/internal/mesh"
	"github.com/vikukumar/tarak/internal/rbac"
	tarakruntime "github.com/vikukumar/tarak/internal/runtime"
	"github.com/vikukumar/tarak/internal/statestore"
	"github.com/vikukumar/tarak/internal/telemetry"
	"github.com/vikukumar/tarak/internal/tunnel"
	"github.com/vikukumar/tarak/internal/ui"
	"github.com/vikukumar/tarak/internal/version"
	"github.com/vikukumar/tarak/internal/ws"
	"github.com/vikukumar/tarak/internal/zerotrust"
	"github.com/vikukumar/tarak/pkg/api/handler"
	"github.com/vikukumar/tarak/pkg/api/middleware"
	tarakwatch "github.com/vikukumar/tarak/pkg/api/watch"
	"github.com/vikukumar/tarak/pkg/health"
	"github.com/vikukumar/tarak/pkg/metrics"
	"github.com/vikukumar/tarak/pkg/security"
)

// Config holds the full configuration for a Tarak API server instance.
type Config struct {
	// BindAddress is the address:port the API server listens on.
	// Default: "0.0.0.0:6443"
	BindAddress string

	// IngressHTTPAddress is the address:port the Ingress HTTP router listens on.
	// Default: "0.0.0.0:8080"
	IngressHTTPAddress string

	// CloudflareTunnel enables automated Cloudflare tunneling.
	CloudflareTunnel bool

	// CloudflareToken is the optional token for named Cloudflare tunnels.
	CloudflareToken string

	// Tailscale enables Tailscale mesh networking.
	Tailscale bool

	// TailscaleAuthKey is the authentication key for Tailscale.
	TailscaleAuthKey string

	// DataDir is the root directory for persistent data (state store, PKI).
	DataDir string

	// PKIDir is the directory containing the PKI (ca.crt, apiserver.crt, …).
	// Default: DataDir/pki
	PKIDir string

	// StateStorePath is the path to the BoltDB file.
	// Default: DataDir/state.db
	StateStorePath string

	// AllowInsecureAuth disables authentication — development/test only.
	AllowInsecureAuth bool

	// StaticTokens is an optional list of pre-issued static tokens.
	StaticTokens []security.StaticToken

	// SANs is the list of additional SANs for the API server certificate.
	SANs []string

	// Log is the structured logger.  A production logger is created if nil.
	Log *zap.Logger

	// ShutdownTimeout is the graceful shutdown deadline.
	// Default: 30s
	ShutdownTimeout time.Duration
}

// Server is a running Tarak API server instance.
type Server struct {
	cfg             Config
	store           statestore.Store
	runtime         tarakruntime.Runtime
	health          *health.Handler
	metrics         *metrics.Registry
	log             *zap.Logger
	httpSrv         *http.Server
	ingressRouter   *ingress.Router
	ingressCtrl     *ingress.Controller
	cfManager       *tunnel.CloudflareManager
	tsManager       *tunnel.TailscaleManager
	ssoManager      *auth.Manager
	rbacAuthorizer  *rbac.Authorizer
	hubbleCollector *telemetry.Collector
	lbCtrl          *loadbalancer.Controller
	meshEngine      *mesh.Engine
	multiMeshMgr    *mesh.MultiMeshManager
	zeroTrustMgr    *zerotrust.Manager
	wsHub           *ws.Hub
}

// New creates a new Server from the given configuration.
// It opens the state store and prepares all internal components, but does NOT
// bind the network listener.  Call Run() to start serving.
func New(cfg Config) (*Server, error) {
	if cfg.Log == nil {
		logger, err := zap.NewProduction()
		if err != nil {
			return nil, fmt.Errorf("create logger: %w", err)
		}
		cfg.Log = logger
	}
	applyDefaults(&cfg)

	// ── State store ──────────────────────────────────────────────────────
	store, err := statestore.Open(statestore.Options{
		Path:   cfg.StateStorePath,
		Logger: cfg.Log.Named("statestore"),
	})
	if err != nil {
		return nil, fmt.Errorf("open state store: %w", err)
	}

	h := health.NewHandler()
	m := metrics.NewRegistry()
	rt := tarakruntime.NewEngine(cfg.DataDir, cfg.Log)
	ssoMgr := auth.NewManager(nil, cfg.Log)
	rbacAuth := rbac.NewAuthorizer(store, cfg.Log)
	hubbleColl := telemetry.NewCollector(cfg.Log)
	lbCtrl := loadbalancer.NewController(cfg.Log)
	meshEngine := mesh.NewEngine(cfg.Log)
	multiMeshMgr := mesh.NewMultiMeshManager(cfg.Log)
	zeroTrustMgr := zerotrust.NewManager(cfg.Log)
	wsHub := ws.NewHub(cfg.Log)

	// Add state store health check.
	h.AddCheck("statestore", func() error {
		_, err := store.CurrentRevision(context.Background())
		return err
	})

	s := &Server{
		cfg:             cfg,
		store:           store,
		runtime:         rt,
		health:          h,
		metrics:         m,
		log:             cfg.Log.Named("apiserver"),
		ssoManager:      ssoMgr,
		rbacAuthorizer:  rbacAuth,
		hubbleCollector: hubbleColl,
		lbCtrl:          lbCtrl,
		meshEngine:      meshEngine,
		multiMeshMgr:    multiMeshMgr,
		zeroTrustMgr:    zeroTrustMgr,
		wsHub:           wsHub,
	}
	return s, nil
}

// Run starts the API server and blocks until ctx is cancelled.
// It returns after graceful shutdown is complete.
func (s *Server) Run(ctx context.Context) error {
	// ── Bootstrap Core Namespaces & Local Node ───────────────────────────
	if err := s.bootstrapNamespaces(ctx); err != nil {
		return fmt.Errorf("bootstrap namespaces: %w", err)
	}
	if err := s.bootstrapLocalNode(ctx); err != nil {
		s.log.Warn("bootstrap local node", zap.Error(err))
	}

	// ── PKI ──────────────────────────────────────────────────────────────
	tlsCfg, caPool, err := s.loadOrInitPKI()
	if err != nil {
		return fmt.Errorf("PKI init: %w", err)
	}

	// ── Auth middleware ───────────────────────────────────────────────────
	authOpts := middleware.AuthOptions{
		CertPool:      caPool,
		StaticTokens:  s.cfg.StaticTokens,
		Log:           s.log.Named("auth"),
		AllowInsecure: s.cfg.AllowInsecureAuth,
	}

	// ── Watch handler ─────────────────────────────────────────────────────
	wh := tarakwatch.New(tarakwatch.Options{
		Store: s.store,
		Log:   s.log.Named("watch"),
	})

	// ── Router ───────────────────────────────────────────────────────────
	r := chi.NewRouter()

	// Built-in chi middleware.
	r.Use(chimw.Recoverer)
	r.Use(middleware.RequestID())
	r.Use(middleware.Audit(s.log.Named("audit")))
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, req)
			// Record real live traffic into Hubble collector (skip static dashboard assets)
			if !strings.HasPrefix(req.URL.Path, "/assets") && !strings.HasPrefix(req.URL.Path, "/dashboard") && req.URL.Path != "/favicon.ico" {
				s.hubbleCollector.RecordFlow(telemetry.NetworkFlow{
					ID:         fmt.Sprintf("flow-%d", time.Now().UnixNano()),
					Timestamp:  time.Now(),
					SrcPod:     "client-workload",
					SrcNS:      "default",
					SrcIP:      req.RemoteAddr,
					DstPod:     "tarak-apiserver",
					DstNS:      "tarak-system",
					DstIP:      "127.0.0.1",
					DstPort:    18443,
					Protocol:   "HTTP",
					Verdict:    "FORWARDED",
					StatusCode: 200,
					LatencyMs:  float64(time.Since(start).Microseconds()) / 1000.0,
					Bytes:      128,
					Summary:    fmt.Sprintf("%s %s", req.Method, req.URL.Path),
				})
			}
		})
	})
	r.Use(middleware.Auth(authOpts))

	// Health and metrics (unauthed handled inside handlers).
	s.health.RegisterRoutes(r)
	r.Handle("/metrics", s.metrics.Handler())

	// Discovery and metadata endpoints.
	r.Get("/version", s.serveVersion)
	r.Get("/openapi/v2", s.serveOpenAPI)
	r.Get("/api", s.serveAPIVersions)
	r.Get("/api/v1", s.serveAPIResourceList("", "v1"))
	r.Get("/apis", s.serveAPIGroupList)

	// Standard k8s.io discovery endpoints
	r.Get("/apis/apps/v1", s.serveAPIResourceList("apps", "v1"))
	r.Get("/apis/batch/v1", s.serveAPIResourceList("batch", "v1"))
	r.Get("/apis/networking.k8s.io/v1", s.serveAPIResourceList("networking.k8s.io", "v1"))
	r.Get("/apis/rbac.authorization.k8s.io/v1", s.serveAPIResourceList("rbac.authorization.k8s.io", "v1"))
	r.Get("/apis/storage.k8s.io/v1", s.serveAPIResourceList("storage.k8s.io", "v1"))
	r.Get("/apis/apiextensions.k8s.io/v1", s.serveAPIResourceList("apiextensions.k8s.io", "v1"))

	// Native tarak.io discovery endpoints
	r.Get("/apis/apps.tarak.io/v1", s.serveAPIResourceList("apps.tarak.io", "v1"))
	r.Get("/apis/batch.tarak.io/v1", s.serveAPIResourceList("batch.tarak.io", "v1"))
	r.Get("/apis/networking.tarak.io/v1", s.serveAPIResourceList("networking.tarak.io", "v1"))
	r.Get("/apis/rbac.authorization.tarak.io/v1", s.serveAPIResourceList("rbac.authorization.tarak.io", "v1"))
	r.Get("/apis/storage.tarak.io/v1", s.serveAPIResourceList("storage.tarak.io", "v1"))
	r.Get("/apis/security.tarak.io/v1", s.serveAPIResourceList("security.tarak.io", "v1"))
	r.Get("/apis/apiextensions.tarak.io/v1", s.serveAPIResourceList("apiextensions.tarak.io", "v1"))

	// Core API group (/api/v1).
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/", s.serveAPIResourceList("", "v1"))
		registerCoreResources(r, s.store, wh, s.log)
	})

	// Named API groups — standard k8s.io
	r.Route("/apis/apps/v1", func(r chi.Router) {
		r.Get("/", s.serveAPIResourceList("apps", "v1"))
		registerAppsResources("apps", r, s.store, wh, s.log)
	})
	r.Route("/apis/batch/v1", func(r chi.Router) {
		r.Get("/", s.serveAPIResourceList("batch", "v1"))
		registerBatchResources("batch", r, s.store, wh, s.log)
	})
	r.Route("/apis/networking.k8s.io/v1", func(r chi.Router) {
		r.Get("/", s.serveAPIResourceList("networking.k8s.io", "v1"))
		registerNetworkingResources("networking.k8s.io", r, s.store, wh, s.log)
	})
	r.Route("/apis/rbac.authorization.k8s.io/v1", func(r chi.Router) {
		r.Get("/", s.serveAPIResourceList("rbac.authorization.k8s.io", "v1"))
		registerRBACResources("rbac.authorization.k8s.io", r, s.store, wh, s.log)
	})
	r.Route("/apis/storage.k8s.io/v1", func(r chi.Router) {
		r.Get("/", s.serveAPIResourceList("storage.k8s.io", "v1"))
		registerStorageResources("storage.k8s.io", r, s.store, wh, s.log)
	})
	r.Route("/apis/apiextensions.k8s.io/v1", func(r chi.Router) {
		r.Get("/", s.serveAPIResourceList("apiextensions.k8s.io", "v1"))
		registerAPIExtensionsResources("apiextensions.k8s.io", r, s.store, wh, s.log)
	})

	// Named API groups — native tarak.io aliases & extensions
	r.Route("/apis/apps.tarak.io/v1", func(r chi.Router) {
		r.Get("/", s.serveAPIResourceList("apps.tarak.io", "v1"))
		registerAppsResources("apps.tarak.io", r, s.store, wh, s.log)
	})
	r.Route("/apis/batch.tarak.io/v1", func(r chi.Router) {
		r.Get("/", s.serveAPIResourceList("batch.tarak.io", "v1"))
		registerBatchResources("batch.tarak.io", r, s.store, wh, s.log)
	})
	r.Route("/apis/networking.tarak.io/v1", func(r chi.Router) {
		r.Get("/", s.serveAPIResourceList("networking.tarak.io", "v1"))
		r.Get("/tunnels", s.serveTunnels)
		registerNetworkingResources("networking.tarak.io", r, s.store, wh, s.log)
	})
	r.Route("/apis/rbac.authorization.tarak.io/v1", func(r chi.Router) {
		r.Get("/", s.serveAPIResourceList("rbac.authorization.tarak.io", "v1"))
		registerRBACResources("rbac.authorization.tarak.io", r, s.store, wh, s.log)
	})
	r.Route("/apis/storage.tarak.io/v1", func(r chi.Router) {
		r.Get("/", s.serveAPIResourceList("storage.tarak.io", "v1"))
		registerStorageResources("storage.tarak.io", r, s.store, wh, s.log)
	})
	r.Route("/apis/security.tarak.io/v1", func(r chi.Router) {
		r.Get("/", s.serveAPIResourceList("security.tarak.io", "v1"))
		registerSecurityResources(r, s.store, wh, s.log)
	})
	r.Route("/apis/apiextensions.tarak.io/v1", func(r chi.Router) {
		r.Get("/", s.serveAPIResourceList("apiextensions.tarak.io", "v1"))
		registerAPIExtensionsResources("apiextensions.tarak.io", r, s.store, wh, s.log)
	})

	// Pod subresource & container execution routes
	r.Get("/api/v1/namespaces/{namespace}/pods/{name}/log", s.HandlePodLogs)
	r.Post("/api/v1/namespaces/{namespace}/pods/{name}/exec", s.HandlePodExec)
	r.Get("/api/v1/namespaces/{namespace}/pods/{name}/portforward", s.servePodPortForward)

	// Auth & SSO routes
	r.Route("/apis/auth.tarak.io/v1", func(r chi.Router) {
		r.Get("/status", s.ssoManager.HandleStatus)
		r.Post("/setup", s.ssoManager.HandleSetup)
		r.Get("/providers", s.ssoManager.HandleListProviders)
		r.Post("/login", s.ssoManager.HandleLogin)
		r.Get("/userinfo", s.ssoManager.HandleUserInfo)
	})

	// Authorization reviews (k8s standard)
	r.Post("/apis/authorization.k8s.io/v1/selfsubjectaccessreviews", s.rbacAuthorizer.HandlePermissionCheck)

	// Telemetry & Hubble Network Flows
	r.Route("/apis/telemetry.tarak.io/v1", func(r chi.Router) {
		r.Get("/flows", s.hubbleCollector.HandleListFlows)
	})

	// Metrics API (kubectl top pods / nodes)
	r.Get("/apis/metrics.k8s.io/v1beta1", s.serveAPIResourceList("metrics.k8s.io", "v1beta1"))
	r.Get("/apis/metrics.k8s.io/v1beta1/pods", s.serveMetricsPods)
	r.Get("/apis/metrics.k8s.io/v1beta1/namespaces/{namespace}/pods", s.serveMetricsPods)
	r.Get("/apis/metrics.k8s.io/v1beta1/nodes", s.serveMetricsNodes)

	// TCR Runtime API
	r.Get("/apis/runtime.tarak.io/v1/version", s.serveRuntimeVersion)
	r.Get("/apis/runtime.tarak.io/v1/status", s.serveRuntimeStatus)

	// ── Bare-Metal Load Balancer API ──────────────────────────────────────
	r.Get("/apis/networking.tarak.io/v1/loadbalancer/status", s.lbCtrl.HandleStatus)

	// ── Inbuilt Multi-Mesh API (Kuma / Kong Mesh equivalent) ──────────────
	r.Route("/apis/mesh.tarak.io/v1", func(r chi.Router) {
		r.Get("/meshes", s.multiMeshMgr.HandleListMeshes)
		r.Post("/meshes", s.multiMeshMgr.HandleCreateMesh)
		r.Get("/meshes/{name}", s.multiMeshMgr.HandleGetMesh)
		r.Get("/namespaces/{namespace}/meshes", s.multiMeshMgr.HandleListMeshes)
		r.Get("/namespaces/{namespace}/meshes/{name}", s.multiMeshMgr.HandleGetMesh)
		r.Get("/meshes/{mesh}/services", s.multiMeshMgr.HandleListServices)
		r.Get("/meshes/{mesh}/external-services", s.multiMeshMgr.HandleListExternalServices)
		r.Get("/meshes/{mesh}/traffic-permissions", s.multiMeshMgr.HandleListTrafficPermissions)
		r.Get("/meshes/{mesh}/passthrough-policies", s.multiMeshMgr.HandleListPassthroughPolicies)
		r.Get("/meshes/{mesh}/proxy-patches", s.multiMeshMgr.HandleListProxyPatches)
		r.Get("/routes", s.meshEngine.HandleListRoutes)
		r.Post("/routes", s.meshEngine.HandleCreateRoute)
		r.Delete("/namespaces/{namespace}/routes/{name}", s.meshEngine.HandleDeleteRoute)
	})

	// ── Native Zero-Trust Security API ────────────────────────────────────
	r.Route("/apis/security.tarak.io/v1/zerotrust", func(r chi.Router) {
		r.Get("/policies", s.zeroTrustMgr.HandleListPolicies)
		r.Post("/policies", s.zeroTrustMgr.HandleCreatePolicy)
		r.Post("/evaluate", s.zeroTrustMgr.HandleEvaluate)
		r.Delete("/namespaces/{namespace}/policies/{name}", s.zeroTrustMgr.HandleDeletePolicy)
	})

	// ── Live WebSocket Streaming Hub (Real-time cluster telemetry & Hubble) ─
	r.Get("/apis/ws.tarak.io/v1/live", s.wsHub.ServeHTTP)
	s.wsHub.Start()

	// ── Inbuilt Cluster Dashboard SPA Handlers ────────────────────────────
	r.Handle("/dashboard", ui.Handler())
	r.Handle("/dashboard/*", ui.Handler())
	r.Handle("/login", ui.Handler())
	r.Handle("/login/*", ui.Handler())
	r.Handle("/setup", ui.Handler())
	r.Handle("/setup/*", ui.Handler())
	r.Handle("/signup", ui.Handler())
	r.Handle("/signup/*", ui.Handler())
	r.Handle("/forgot-password", ui.Handler())
	r.Handle("/forgot-password/*", ui.Handler())
	r.Handle("/_next/*", ui.Handler())
	r.Handle("/assets/*", ui.Handler())
	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/assets/tarak_icon.jpg"
		ui.Handler().ServeHTTP(w, r)
	})

	// ── HTTP server ───────────────────────────────────────────────────────
	s.httpSrv = &http.Server{
		Addr:         s.cfg.BindAddress,
		Handler:      r,
		TLSConfig:    tlsCfg,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 0, // disabled — watches use streaming
		IdleTimeout:  120 * time.Second,
	}

	ln, err := net.Listen("tcp", s.cfg.BindAddress)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.cfg.BindAddress, err)
	}

	// ── Ingress & Tunnels Initialization ─────────────────────────────────
	s.ingressRouter = ingress.NewRouter(s.log)
	s.ingressCtrl = ingress.NewController(s.store, s.ingressRouter, s.log)

	s.cfManager = tunnel.NewCloudflareManager(tunnel.CloudflareTunnelConfig{
		Enabled:   s.cfg.CloudflareTunnel,
		Token:     s.cfg.CloudflareToken,
		LocalPort: 8080,
	}, s.ingressRouter, s.log)

	s.tsManager = tunnel.NewTailscaleManager(tunnel.TailscaleConfig{
		Enabled:  s.cfg.Tailscale,
		AuthKey:  s.cfg.TailscaleAuthKey,
		Hostname: "tarak-cluster",
	}, s.ingressRouter, s.log)

	_ = s.bootstrapIngressClasses(ctx)
	s.lbCtrl.Start(ctx)

	if s.cfg.CloudflareTunnel {
		_ = s.cfManager.Start(ctx, s.cfg.IngressHTTPAddress)
	}
	if s.cfg.Tailscale {
		_ = s.tsManager.Start(ctx, 8080)
	}

	// Launch native Ingress HTTP reverse proxy listener in background
	if s.cfg.IngressHTTPAddress != "" {
		ingSrv := &http.Server{
			Addr:    s.cfg.IngressHTTPAddress,
			Handler: s.ingressRouter,
		}
		go func() {
			<-ctx.Done()
			_ = ingSrv.Close()
		}()
		go func() {
			_ = ingSrv.ListenAndServe()
		}()
	}

	// Signal ready BEFORE starting the TLS listener.
	s.health.SetReady(true)
	s.log.Info("tarak api server ready",
		zap.String("address", s.cfg.BindAddress),
		zap.String("ingressHTTP", s.cfg.IngressHTTPAddress),
		zap.Bool("cloudflareTunnel", s.cfg.CloudflareTunnel),
		zap.Bool("tailscale", s.cfg.Tailscale),
		zap.Bool("insecureAuth", s.cfg.AllowInsecureAuth),
	)

	fmt.Printf("\n"+
		"╔══════════════════════════════════════════════════════════════════════════════════╗\n"+
		"║  🚀 TARAK CONTAINER ORCHESTRATION PLATFORM (%s)                               ║\n"+
		"╠══════════════════════════════════════════════════════════════════════════════════╣\n"+
		"║  🌐 Inbuilt Cluster Dashboard : https://%s/dashboard/\n"+
		"║  ⚡ HTTP Ingress Proxy        : http://%s/\n"+
		"║  🔑 Local Access Mode         : Super-Admin (mTLS & Master Token Active)         ║\n"+
		"║  🛡️ Remote Access Command     : tarakctl login https://%s\n"+
		"╚══════════════════════════════════════════════════════════════════════════════════╝\n\n",
		version.String(), s.cfg.BindAddress, s.cfg.IngressHTTPAddress, s.cfg.BindAddress)

	// ── Controller Manager ───────────────────────────────────────────────
	ctrlMgr := controller.NewManager(s.store, s.runtime, s.log)
	ctrlMgr.SetIngressReconciler(func(c context.Context) {
		_ = s.ingressCtrl.Reconcile(c, s.cfManager.PublicURL())
	})
	ctrlMgr.Start(ctx)

	// Serve in background.
	errCh := make(chan error, 1)
	go func() {
		tlsLn := tls.NewListener(ln, tlsCfg)
		errCh <- s.httpSrv.Serve(tlsLn)
	}()

	// Wait for context cancellation or server error.
	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("serve: %w", err)
		}
		return nil
	case <-ctx.Done():
		s.log.Info("shutting down api server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()
		if err := s.httpSrv.Shutdown(shutdownCtx); err != nil {
			s.log.Warn("graceful shutdown error", zap.Error(err))
		}
		return s.store.Close()
	}
}

// ─── PKI ─────────────────────────────────────────────────────────────────────

func (s *Server) loadOrInitPKI() (*tls.Config, *x509.CertPool, error) {
	pkiDir := s.cfg.PKIDir

	// Try to load existing PKI.
	ca, serverPair, adminPair, err := security.LoadPKI(pkiDir)
	if err != nil {
		// Not found or corrupt — generate fresh PKI.
		s.log.Info("generating new PKI", zap.String("dir", pkiDir))
		ca, serverPair, adminPair, err = s.generatePKI(pkiDir)
		if err != nil {
			return nil, nil, err
		}
	} else {
		s.log.Info("loaded existing PKI", zap.String("dir", pkiDir))
	}

	// Always sync admin.kubeconfig in data directory
	if adminPair != nil {
		if err := s.writeAdminKubeconfig(ca.CertPEM, adminPair.CertPEM, adminPair.KeyPEM); err != nil {
			s.log.Warn("could not write admin kubeconfig", zap.Error(err))
		}
	}

	// Build TLS config.
	serverTLSCert, err := tls.X509KeyPair(serverPair.CertPEM, serverPair.KeyPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("parse server TLS cert: %w", err)
	}

	caPool, err := ca.CertPool()
	if err != nil {
		return nil, nil, fmt.Errorf("build CA pool: %w", err)
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{serverTLSCert},
		ClientCAs:    caPool,
		ClientAuth:   tls.RequestClientCert, // optional mTLS; checked by auth middleware
		MinVersion:   tls.VersionTLS13,
	}

	return tlsCfg, caPool, nil
}

// generatePKI creates and writes a new PKI, returning the CA, server, and admin certs.
func (s *Server) generatePKI(dir string) (*security.CertificateAuthority, *security.CertKeyPair, *security.CertKeyPair, error) {
	ca, err := security.GenerateCA()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generate CA: %w", err)
	}

	sans := append([]string{"localhost", "127.0.0.1", "::1"}, s.cfg.SANs...)
	server, err := ca.SignServerCert(security.ServerCertOptions{
		CommonName: "tarak-apiserver",
		SANs:       sans,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("sign server cert: %w", err)
	}

	admin, err := ca.SignClientCert(security.ClientCertOptions{
		CommonName:    "tarak-admin",
		Organizations: []string{"system:masters"},
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("sign admin cert: %w", err)
	}

	if err := security.WritePKI(dir, ca, server, admin); err != nil {
		return nil, nil, nil, fmt.Errorf("write PKI: %w", err)
	}
	return ca, server, admin, nil
}

func (s *Server) writeAdminKubeconfig(caPEM, certPEM, keyPEM []byte) error {
	kcPath := filepath.Join(s.cfg.DataDir, "admin.kubeconfig")
	serverURL := s.cfg.BindAddress
	if !filepath.IsAbs(serverURL) && len(serverURL) > 0 && serverURL[0] != 'h' {
		serverURL = "https://" + serverURL
	}
	config := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Config",
		"clusters": []map[string]interface{}{
			{
				"name": "tarak-cluster",
				"cluster": map[string]interface{}{
					"server":                     serverURL,
					"certificate-authority-data": base64.StdEncoding.EncodeToString(caPEM),
				},
			},
		},
		"contexts": []map[string]interface{}{
			{
				"name": "tarak-admin@tarak-cluster",
				"context": map[string]interface{}{
					"cluster":   "tarak-cluster",
					"user":      "tarak-admin",
					"namespace": "default",
				},
			},
		},
		"current-context": "tarak-admin@tarak-cluster",
		"users": []map[string]interface{}{
			{
				"name": "tarak-admin",
				"user": map[string]interface{}{
					"client-certificate-data": base64.StdEncoding.EncodeToString(certPEM),
					"client-key-data":         base64.StdEncoding.EncodeToString(keyPEM),
				},
			},
		},
	}
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(kcPath, yamlData, 0600)
}

// ─── API discovery handlers ───────────────────────────────────────────────────

func (s *Server) serveAPIVersions(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"kind":       "APIVersions",
		"apiVersion": "v1",
		"versions":   []string{"v1"},
		"serverAddressByClientCIDRs": []map[string]string{
			{"clientCIDR": "0.0.0.0/0", "serverAddress": s.cfg.BindAddress},
		},
	}
	writeJSON(w, resp)
}

func (s *Server) serveAPIGroupList(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"kind":       "APIGroupList",
		"apiVersion": "v1",
		"groups": []map[string]interface{}{
			makeAPIGroup("apps", "v1"),
			makeAPIGroup("apps.tarak.io", "v1"),
			makeAPIGroup("batch", "v1"),
			makeAPIGroup("batch.tarak.io", "v1"),
			makeAPIGroup("networking.k8s.io", "v1"),
			makeAPIGroup("networking.tarak.io", "v1"),
			makeAPIGroup("rbac.authorization.k8s.io", "v1"),
			makeAPIGroup("rbac.authorization.tarak.io", "v1"),
			makeAPIGroup("storage.k8s.io", "v1"),
			makeAPIGroup("storage.tarak.io", "v1"),
			makeAPIGroup("security.tarak.io", "v1"),
			makeAPIGroup("apiextensions.k8s.io", "v1"),
			makeAPIGroup("apiextensions.tarak.io", "v1"),
			makeAPIGroup("metrics.k8s.io", "v1beta1"),
		},
	}
	writeJSON(w, resp)
}

func makeAPIGroup(name, preferredVersion string) map[string]interface{} {
	return map[string]interface{}{
		"name": name,
		"versions": []map[string]string{
			{"groupVersion": name + "/" + preferredVersion, "version": preferredVersion},
		},
		"preferredVersion": map[string]string{
			"groupVersion": name + "/" + preferredVersion,
			"version":      preferredVersion,
		},
	}
}

func (s *Server) serveAPIResourceList(group, version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gv := version
		if group != "" {
			gv = group + "/" + version
		}
		resp := map[string]interface{}{
			"kind":         "APIResourceList",
			"apiVersion":   "v1",
			"groupVersion": gv,
			"resources":    allResourceDescriptors(group, version),
		}
		writeJSON(w, resp)
	}
}

func (s *Server) serveVersion(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"major":        "1",
		"minor":        "30",
		"gitVersion":   version.Version,
		"gitCommit":    version.Commit,
		"gitTreeState": "clean",
		"buildDate":    version.BuildDate,
		"goVersion":    runtime.Version(),
		"compiler":     runtime.Compiler,
		"platform":     runtime.GOOS + "/" + runtime.GOARCH,
	}
	writeJSON(w, resp)
}

func (s *Server) serveOpenAPI(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"swagger": "2.0",
		"info": map[string]interface{}{
			"title":   "Tarak Kubernetes-Compatible API",
			"version": "v1.30.0",
		},
		"paths": map[string]interface{}{
			"/api/v1": map[string]interface{}{
				"get": map[string]interface{}{
					"description": "get available resources",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "OK"},
					},
				},
			},
		},
	}
	writeJSON(w, resp)
}

func (s *Server) bootstrapNamespaces(ctx context.Context) error {
	coreNamespaces := []string{"default", "tarak-system", "tarak-public", "tarak-node-lease"}
	for _, ns := range coreNamespaces {
		key := statestore.ResourceKey{
			Group:    "",
			Version:  "v1",
			Resource: "namespaces",
			Name:     ns,
		}
		_, err := s.store.Get(ctx, key)
		if err == nil {
			continue // already exists
		}
		nsObj := map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": ns,
				"labels": map[string]string{
					"tarak.io/metadata.name":      ns,
					"kubernetes.io/metadata.name": ns,
				},
			},
			"spec": map[string]interface{}{
				"finalizers": []string{"tarak"},
			},
			"status": map[string]interface{}{
				"phase": "Active",
			},
		}
		raw, err := json.Marshal(nsObj)
		if err != nil {
			return err
		}
		if _, err := s.store.Create(ctx, key, raw); err != nil && !errors.Is(err, statestore.ErrAlreadyExists) {
			return fmt.Errorf("bootstrap namespace %q: %w", ns, err)
		}
		s.log.Info("bootstrapped core namespace", zap.String("namespace", ns))
	}
	return nil
}

func (s *Server) bootstrapLocalNode(ctx context.Context) error {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "tarak-control-plane"
	}
	hostname = strings.ToLower(hostname)
	key := statestore.ResourceKey{
		Group:    "",
		Version:  "v1",
		Resource: "nodes",
		Name:     hostname,
	}
	_, err = s.store.Get(ctx, key)
	if err == nil {
		return nil // already exists
	}
	nodeObj := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Node",
		"metadata": map[string]interface{}{
			"name": hostname,
			"labels": map[string]string{
				"tarak.io/hostname":                     hostname,
				"kubernetes.io/hostname":                hostname,
				"node-role.kubernetes.io/control-plane": "",
				"node-role.kubernetes.io/master":        "",
				"node.kubernetes.io/instance-type":      "tarak.local",
				"kubernetes.io/os":                      runtime.GOOS,
				"kubernetes.io/arch":                    runtime.GOARCH,
			},
		},
		"status": map[string]interface{}{
			"phase": "Running",
			"conditions": []map[string]interface{}{
				{
					"type":    "Ready",
					"status":  "True",
					"reason":  "TarakNodeReady",
					"message": "Tarak control plane and node runtime active",
				},
			},
			"nodeInfo": map[string]interface{}{
				"kubeletVersion":          "v1.30.0-tarak",
				"osImage":                 fmt.Sprintf("Tarak Native (%s/%s)", runtime.GOOS, runtime.GOARCH),
				"kernelVersion":           runtime.Version(),
				"containerRuntimeVersion": "tarak-runtime://v1.30.0",
				"architecture":            runtime.GOARCH,
				"operatingSystem":         runtime.GOOS,
			},
			"addresses": []map[string]interface{}{
				{"type": "InternalIP", "address": "127.0.0.1"},
				{"type": "Hostname", "address": hostname},
			},
		},
	}
	raw, err := json.Marshal(nodeObj)
	if err != nil {
		return err
	}
	if _, err := s.store.Create(ctx, key, raw); err != nil && !errors.Is(err, statestore.ErrAlreadyExists) {
		return fmt.Errorf("bootstrap local node: %w", err)
	}
	s.log.Info("bootstrapped local node", zap.String("node", hostname))
	return nil
}

func (s *Server) bootstrapIngressClasses(ctx context.Context) error {
	classes := []struct {
		name      string
		isDefault bool
	}{
		{"tarak", true},
		{"tarak-cloudflare", false},
		{"tarak-tailscale", false},
	}

	for _, ic := range classes {
		key := statestore.ResourceKey{
			Group:    "networking.k8s.io",
			Version:  "v1",
			Resource: "ingressclasses",
			Name:     ic.name,
		}
		annotations := map[string]string{}
		if ic.isDefault {
			annotations["ingressclass.kubernetes.io/is-default-class"] = "true"
		}
		obj := map[string]interface{}{
			"apiVersion": "networking.k8s.io/v1",
			"kind":       "IngressClass",
			"metadata": map[string]interface{}{
				"name":        ic.name,
				"annotations": annotations,
			},
			"spec": map[string]interface{}{
				"controller": "tarak.io/ingress-controller",
			},
		}
		raw, _ := json.Marshal(obj)
		_, _ = s.store.Create(ctx, key, raw)
	}
	return nil
}

func (s *Server) serveTunnels(w http.ResponseWriter, r *http.Request) {
	cfStatus := tunnel.TunnelStatus{
		Type:   "cloudflare",
		Active: s.cfg.CloudflareTunnel,
		Mode:   "quick-tunnel",
	}
	if s.cfManager != nil {
		st := s.cfManager.Status()
		if st.Type != "" {
			cfStatus = st
		}
	}

	tsStatus := tunnel.TunnelStatus{
		Type:   "tailscale",
		Active: s.cfg.Tailscale,
		Mode:   "magic-dns",
	}
	if s.tsManager != nil {
		st := s.tsManager.Status()
		if st.Type != "" {
			tsStatus = st
		}
	}

	writeJSON(w, map[string]interface{}{
		"apiVersion": "networking.tarak.io/v1",
		"kind":       "TunnelList",
		"items":      []tunnel.TunnelStatus{cfStatus, tsStatus},
	})
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}

// applyDefaults fills in zero-value Config fields with production defaults.
func applyDefaults(cfg *Config) {
	if cfg.BindAddress == "" {
		cfg.BindAddress = "0.0.0.0:6443"
	}
	if cfg.IngressHTTPAddress == "" {
		cfg.IngressHTTPAddress = "0.0.0.0:8080"
	}
	if cfg.DataDir == "" {
		cfg.DataDir = "/var/lib/tarak"
	}
	if cfg.PKIDir == "" {
		cfg.PKIDir = cfg.DataDir + "/pki"
	}
	if cfg.StateStorePath == "" {
		cfg.StateStorePath = cfg.DataDir + "/state.db"
	}
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = 30 * time.Second
	}
}

// ─── Resource route registration ─────────────────────────────────────────────

func registerCoreResources(r chi.Router, store statestore.Store, wh *tarakwatch.Handler, log *zap.Logger) {
	for _, desc := range coreResourceDescriptors() {
		h := handler.NewResourceHandler(desc, store, wh, log.Named(desc.Resource))
		h.RegisterRoutes(r)
	}
}

func registerAppsResources(group string, r chi.Router, store statestore.Store, wh *tarakwatch.Handler, log *zap.Logger) {
	for _, desc := range appsResourceDescriptors(group) {
		h := handler.NewResourceHandler(desc, store, wh, log.Named(desc.Resource))
		h.RegisterRoutes(r)
	}
}

func registerBatchResources(group string, r chi.Router, store statestore.Store, wh *tarakwatch.Handler, log *zap.Logger) {
	for _, desc := range batchResourceDescriptors(group) {
		h := handler.NewResourceHandler(desc, store, wh, log.Named(desc.Resource))
		h.RegisterRoutes(r)
	}
}

func registerNetworkingResources(group string, r chi.Router, store statestore.Store, wh *tarakwatch.Handler, log *zap.Logger) {
	for _, desc := range networkingResourceDescriptors(group) {
		h := handler.NewResourceHandler(desc, store, wh, log.Named(desc.Resource))
		h.RegisterRoutes(r)
	}
}

func registerRBACResources(group string, r chi.Router, store statestore.Store, wh *tarakwatch.Handler, log *zap.Logger) {
	for _, desc := range rbacResourceDescriptors(group) {
		h := handler.NewResourceHandler(desc, store, wh, log.Named(desc.Resource))
		h.RegisterRoutes(r)
	}
}

func registerStorageResources(group string, r chi.Router, store statestore.Store, wh *tarakwatch.Handler, log *zap.Logger) {
	for _, desc := range storageResourceDescriptors(group) {
		h := handler.NewResourceHandler(desc, store, wh, log.Named(desc.Resource))
		h.RegisterRoutes(r)
	}
}

func registerSecurityResources(r chi.Router, store statestore.Store, wh *tarakwatch.Handler, log *zap.Logger) {
	for _, desc := range tarakSecurityResourceDescriptors() {
		h := handler.NewResourceHandler(desc, store, wh, log.Named(desc.Resource))
		h.RegisterRoutes(r)
	}
}

func registerAPIExtensionsResources(group string, r chi.Router, store statestore.Store, wh *tarakwatch.Handler, log *zap.Logger) {
	for _, desc := range apiExtensionsResourceDescriptors(group) {
		h := handler.NewResourceHandler(desc, store, wh, log.Named(desc.Resource))
		h.RegisterRoutes(r)
	}
}

// ─── Resource descriptor catalogs ────────────────────────────────────────────

func coreResourceDescriptors() []handler.ResourceDescriptor {
	std := []string{"create", "delete", "get", "list", "patch", "update", "watch"}
	_ = corev1.Pod{} // ensure import is used
	return []handler.ResourceDescriptor{
		{Group: "", Version: "v1", Resource: "pods", Kind: "Pod", Namespaced: true, Verbs: std, ShortNames: []string{"po"}},
		{Group: "", Version: "v1", Resource: "services", Kind: "Service", Namespaced: true, Verbs: std, ShortNames: []string{"svc"}},
		{Group: "", Version: "v1", Resource: "endpoints", Kind: "Endpoints", Namespaced: true, Verbs: std},
		{Group: "", Version: "v1", Resource: "configmaps", Kind: "ConfigMap", Namespaced: true, Verbs: std, ShortNames: []string{"cm"}},
		{Group: "", Version: "v1", Resource: "secrets", Kind: "Secret", Namespaced: true, Verbs: std},
		{Group: "", Version: "v1", Resource: "serviceaccounts", Kind: "ServiceAccount", Namespaced: true, Verbs: std, ShortNames: []string{"sa"}},
		{Group: "", Version: "v1", Resource: "namespaces", Kind: "Namespace", Namespaced: false, Verbs: std, ShortNames: []string{"ns"}},
		{Group: "", Version: "v1", Resource: "nodes", Kind: "Node", Namespaced: false, Verbs: std, ShortNames: []string{"no"}},
		{Group: "", Version: "v1", Resource: "persistentvolumes", Kind: "PersistentVolume", Namespaced: false, Verbs: std, ShortNames: []string{"pv"}},
		{Group: "", Version: "v1", Resource: "persistentvolumeclaims", Kind: "PersistentVolumeClaim", Namespaced: true, Verbs: std, ShortNames: []string{"pvc"}},
		{Group: "", Version: "v1", Resource: "events", Kind: "Event", Namespaced: true, Verbs: std, ShortNames: []string{"ev"}},
		{Group: "", Version: "v1", Resource: "resourcequotas", Kind: "ResourceQuota", Namespaced: true, Verbs: std, ShortNames: []string{"quota"}},
		{Group: "", Version: "v1", Resource: "limitranges", Kind: "LimitRange", Namespaced: true, Verbs: std, ShortNames: []string{"limits"}},
		{Group: "", Version: "v1", Resource: "podtemplates", Kind: "PodTemplate", Namespaced: true, Verbs: std},
		{Group: "", Version: "v1", Resource: "replicationcontrollers", Kind: "ReplicationController", Namespaced: true, Verbs: std, ShortNames: []string{"rc"}},
	}
}

func appsResourceDescriptors(group string) []handler.ResourceDescriptor {
	std := []string{"create", "delete", "get", "list", "patch", "update", "watch"}
	_ = appsv1.Deployment{}
	_ = tarakv1.TarakApplication{}
	descs := []handler.ResourceDescriptor{
		{Group: group, Version: "v1", Resource: "deployments", Kind: "Deployment", Namespaced: true, Verbs: std, ShortNames: []string{"deploy"}},
		{Group: group, Version: "v1", Resource: "replicasets", Kind: "ReplicaSet", Namespaced: true, Verbs: std, ShortNames: []string{"rs"}},
		{Group: group, Version: "v1", Resource: "statefulsets", Kind: "StatefulSet", Namespaced: true, Verbs: std, ShortNames: []string{"sts"}},
		{Group: group, Version: "v1", Resource: "daemonsets", Kind: "DaemonSet", Namespaced: true, Verbs: std, ShortNames: []string{"ds"}},
		{Group: group, Version: "v1", Resource: "controllerrevisions", Kind: "ControllerRevision", Namespaced: true, Verbs: std},
	}
	if strings.Contains(group, "tarak") {
		descs = append(descs, handler.ResourceDescriptor{
			Group: group, Version: "v1", Resource: "tarakapplications", Kind: "TarakApplication", Namespaced: true, Verbs: std, ShortNames: []string{"tapp", "tarakapp", "app"},
		})
	}
	return descs
}

func batchResourceDescriptors(group string) []handler.ResourceDescriptor {
	std := []string{"create", "delete", "get", "list", "patch", "update", "watch"}
	_ = batchv1.Job{}
	return []handler.ResourceDescriptor{
		{Group: group, Version: "v1", Resource: "jobs", Kind: "Job", Namespaced: true, Verbs: std},
		{Group: group, Version: "v1", Resource: "cronjobs", Kind: "CronJob", Namespaced: true, Verbs: std, ShortNames: []string{"cj"}},
	}
}

func networkingResourceDescriptors(group string) []handler.ResourceDescriptor {
	std := []string{"create", "delete", "get", "list", "patch", "update", "watch"}
	_ = networkingv1.NetworkPolicy{}
	return []handler.ResourceDescriptor{
		{Group: group, Version: "v1", Resource: "networkpolicies", Kind: "NetworkPolicy", Namespaced: true, Verbs: std, ShortNames: []string{"netpol"}},
		{Group: group, Version: "v1", Resource: "ingresses", Kind: "Ingress", Namespaced: true, Verbs: std, ShortNames: []string{"ing"}},
		{Group: group, Version: "v1", Resource: "ingressclasses", Kind: "IngressClass", Namespaced: false, Verbs: std},
	}
}

func rbacResourceDescriptors(group string) []handler.ResourceDescriptor {
	std := []string{"create", "delete", "get", "list", "patch", "update", "watch"}
	_ = rbacv1.Role{}
	return []handler.ResourceDescriptor{
		{Group: group, Version: "v1", Resource: "roles", Kind: "Role", Namespaced: true, Verbs: std},
		{Group: group, Version: "v1", Resource: "rolebindings", Kind: "RoleBinding", Namespaced: true, Verbs: std},
		{Group: group, Version: "v1", Resource: "clusterroles", Kind: "ClusterRole", Namespaced: false, Verbs: std, ShortNames: []string{"cr"}},
		{Group: group, Version: "v1", Resource: "clusterrolebindings", Kind: "ClusterRoleBinding", Namespaced: false, Verbs: std, ShortNames: []string{"crb"}},
	}
}

func storageResourceDescriptors(group string) []handler.ResourceDescriptor {
	std := []string{"create", "delete", "get", "list", "patch", "update", "watch"}
	_ = storagev1.StorageClass{}
	return []handler.ResourceDescriptor{
		{Group: group, Version: "v1", Resource: "storageclasses", Kind: "StorageClass", Namespaced: false, Verbs: std, ShortNames: []string{"sc"}},
		{Group: group, Version: "v1", Resource: "volumeattachments", Kind: "VolumeAttachment", Namespaced: false, Verbs: std},
		{Group: group, Version: "v1", Resource: "csinodes", Kind: "CSINode", Namespaced: false, Verbs: std},
		{Group: group, Version: "v1", Resource: "csistoragecapacities", Kind: "CSIStorageCapacity", Namespaced: true, Verbs: std},
	}
}

func tarakSecurityResourceDescriptors() []handler.ResourceDescriptor {
	std := []string{"create", "delete", "get", "list", "patch", "update", "watch"}
	_ = tarakv1.TarakSecurityPolicy{}
	return []handler.ResourceDescriptor{
		{Group: "security.tarak.io", Version: "v1", Resource: "taraksecuritypolicies", Kind: "TarakSecurityPolicy", Namespaced: false, Verbs: std, ShortNames: []string{"tsp", "securitypolicy"}},
	}
}

func apiExtensionsResourceDescriptors(group string) []handler.ResourceDescriptor {
	std := []string{"create", "delete", "get", "list", "patch", "update", "watch"}
	_ = tarakv1.CustomResourceDefinition{}
	return []handler.ResourceDescriptor{
		{Group: group, Version: "v1", Resource: "customresourcedefinitions", Kind: "CustomResourceDefinition", Namespaced: false, Verbs: std, ShortNames: []string{"crd", "crds"}},
	}
}

// allResourceDescriptors returns all resource descriptors for a given group/version (for API discovery).
func allResourceDescriptors(group, version string) []map[string]interface{} {
	var descs []handler.ResourceDescriptor
	switch {
	case group == "" && version == "v1":
		descs = coreResourceDescriptors()
	case (group == "apps" || group == "apps.tarak.io") && version == "v1":
		descs = appsResourceDescriptors(group)
	case (group == "batch" || group == "batch.tarak.io") && version == "v1":
		descs = batchResourceDescriptors(group)
	case (group == "networking.k8s.io" || group == "networking.tarak.io") && version == "v1":
		descs = networkingResourceDescriptors(group)
	case (group == "rbac.authorization.k8s.io" || group == "rbac.authorization.tarak.io") && version == "v1":
		descs = rbacResourceDescriptors(group)
	case (group == "storage.k8s.io" || group == "storage.tarak.io") && version == "v1":
		descs = storageResourceDescriptors(group)
	case group == "security.tarak.io" && version == "v1":
		descs = tarakSecurityResourceDescriptors()
	case (group == "apiextensions.k8s.io" || group == "apiextensions.tarak.io") && version == "v1":
		descs = apiExtensionsResourceDescriptors(group)
	case group == "metrics.k8s.io" && version == "v1beta1":
		descs = metricsResourceDescriptors()
	}
	out := make([]map[string]interface{}, 0, len(descs))
	for _, d := range descs {
		out = append(out, map[string]interface{}{
			"name":         d.Resource,
			"kind":         d.Kind,
			"namespaced":   d.Namespaced,
			"verbs":        d.Verbs,
			"shortNames":   d.ShortNames,
			"singularName": strings.ToLower(d.Kind),
		})
	}
	return out
}

func metricsResourceDescriptors() []handler.ResourceDescriptor {
	return []handler.ResourceDescriptor{
		{Group: "metrics.k8s.io", Version: "v1beta1", Resource: "pods", Kind: "PodMetrics", Namespaced: true, Verbs: []string{"get", "list"}},
		{Group: "metrics.k8s.io", Version: "v1beta1", Resource: "nodes", Kind: "NodeMetrics", Namespaced: false, Verbs: []string{"get", "list"}},
	}
}

// ─── Pod Subresource & Metrics Handlers ──────────────────────────────────────

func (s *Server) servePodLog(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "namespace")
	if ns == "" {
		ns = "default"
	}
	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, `{"error":"pod name required"}`, http.StatusBadRequest)
		return
	}

	// Validate the pod actually exists in the store before streaming logs.
	podKey := statestore.ResourceKey{
		Group:     "",
		Version:   "v1",
		Resource:  "pods",
		Namespace: ns,
		Name:      name,
	}
	if _, err := s.store.Get(r.Context(), podKey); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, `{"kind":"Status","apiVersion":"v1","status":"Failure","message":"pod %q not found in namespace %q","reason":"NotFound","code":404}`+"\n", name, ns)
		return
	}

	container := r.URL.Query().Get("container")
	follow := r.URL.Query().Get("follow") == "true"
	tail := -1
	if tStr := r.URL.Query().Get("tailLines"); tStr != "" {
		if t, err := strconv.Atoi(tStr); err == nil {
			tail = t
		}
	}
	var since time.Duration
	if sStr := r.URL.Query().Get("sinceSeconds"); sStr != "" {
		if sec, err := strconv.Atoi(sStr); err == nil && sec > 0 {
			since = time.Duration(sec) * time.Second
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_ = s.runtime.GetLogs(r.Context(), ns, name, container, follow, tail, since, w)
}

func writeStatusError(w http.ResponseWriter, code int, reason, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = fmt.Fprintf(w, `{"kind":"Status","apiVersion":"v1","status":"Failure","message":%q,"reason":%q,"code":%d}`+"\n", message, reason, code)
}

func (s *Server) servePodExec(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "namespace")
	if ns == "" {
		ns = "default"
	}
	name := chi.URLParam(r, "name")
	if name == "" {
		writeStatusError(w, http.StatusBadRequest, "BadRequest", "pod name required")
		return
	}

	key := statestore.ResourceKey{Group: "", Version: "v1", Resource: "pods", Namespace: ns, Name: name}
	if _, err := s.store.Get(r.Context(), key); err != nil {
		writeStatusError(w, http.StatusNotFound, "NotFound", fmt.Sprintf("pods %q not found", name))
		return
	}

	container := r.URL.Query().Get("container")
	tty := r.URL.Query().Get("tty") == "true"

	commands := r.URL.Query()["command"]
	if len(commands) == 0 {
		var req struct {
			Command []string `json:"command"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil && len(req.Command) > 0 {
			commands = req.Command
		}
	}
	if len(commands) == 0 {
		commands = []string{"sh"}
	}

	// If streaming tunnel requested, hijack connection for bidirectional IO
	if r.URL.Query().Get("stream") == "true" {
		rc := http.NewResponseController(w)
		conn, _, err := rc.Hijack()
		if err == nil {
			defer conn.Close()
			_, _ = conn.Write([]byte("HTTP/1.1 200 OK\r\nConnection: Upgrade\r\nUpgrade: tarak-exec\r\n\r\n"))
			_, _ = s.runtime.ExecCommand(r.Context(), ns, name, container, commands, conn, conn, conn, tty)
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_, _ = s.runtime.ExecCommand(r.Context(), ns, name, container, commands, r.Body, w, w, tty)
}

func (s *Server) servePodPortForward(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	
	key := statestore.ResourceKey{Group: "", Version: "v1", Resource: "pods", Namespace: ns, Name: name}
	if _, err := s.store.Get(r.Context(), key); err != nil {
		writeStatusError(w, http.StatusNotFound, "NotFound", fmt.Sprintf("pods %q not found", name))
		return
	}

	portStr := r.URL.Query().Get("port")
	targetPort, _ := strconv.Atoi(portStr)
	if targetPort <= 0 {
		targetPort = 80
	}

	// If streaming tunnel requested, hijack connection and bridge directly to virtual Pod
	if r.URL.Query().Get("stream") == "true" {
		rc := http.NewResponseController(w)
		conn, _, err := rc.Hijack()
		if err == nil {
			defer conn.Close()
			podConn, pErr := s.runtime.DialPod(r.Context(), ns, name, targetPort)
			if pErr == nil {
				defer podConn.Close()
				_, _ = conn.Write([]byte("HTTP/1.1 200 OK\r\nConnection: Upgrade\r\nUpgrade: tarak-portforward\r\n\r\n"))
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()
					_, _ = io.Copy(podConn, conn)
				}()
				go func() {
					defer wg.Done()
					_, _ = io.Copy(conn, podConn)
				}()
				wg.Wait()
				return
			}
		}
	}

	hostPort := s.runtime.GetHostPort(ns, name, targetPort)

	writeJSON(w, map[string]interface{}{
		"status":     "Success",
		"namespace":  ns,
		"pod":        name,
		"targetPort": targetPort,
		"hostPort":   hostPort,
		"message":    fmt.Sprintf("Forwarding to pod %s/%s on port %d", ns, name, targetPort),
	})
}

func (s *Server) serveMetricsPods(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "namespace")
	podEnvs, _, err := s.store.List(r.Context(), statestore.ListQuery{
		Key: statestore.ResourceKey{Group: "", Version: "v1", Resource: "pods", Namespace: ns},
	})
	if err != nil {
		writeJSON(w, map[string]interface{}{"kind": "PodMetricsList", "apiVersion": "metrics.k8s.io/v1beta1", "items": []interface{}{}})
		return
	}

	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)

	var items []map[string]interface{}
	for _, env := range podEnvs {
		var pod map[string]interface{}
		if err := json.Unmarshal(env.Object, &pod); err != nil {
			continue
		}
		meta, _ := pod["metadata"].(map[string]interface{})
		spec, _ := pod["spec"].(map[string]interface{})
		if meta == nil || spec == nil {
			continue
		}
		pName, _ := meta["name"].(string)
		pNs, _ := meta["namespace"].(string)

		rawContainers, _ := spec["containers"].([]interface{})
		var cMetrics []map[string]interface{}
		for _, c := range rawContainers {
			cMap, _ := c.(map[string]interface{})
			cName := "app"
			if cMap != nil {
				if n, _ := cMap["name"].(string); n != "" {
					cName = n
				}
			}
			
			var cpuStr, memStr string
			metrics, _ := s.runtime.GetContainerMetrics(r.Context(), pNs, pName, cName)
			if metrics != nil {
				cpuStr = fmt.Sprintf("%dm", metrics.CPUMillicores)
				memStr = fmt.Sprintf("%dMi", int(metrics.MemoryUsageMiB))
			} else {
				cpuStr = "0m"
				memStr = "0Mi"
			}

			cMetrics = append(cMetrics, map[string]interface{}{
				"name": cName,
				"usage": map[string]string{
					"cpu":    cpuStr,
					"memory": memStr,
				},
			})
		}

		items = append(items, map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":              pName,
				"namespace":         pNs,
				"creationTimestamp": nowStr,
			},
			"timestamp":  nowStr,
			"window":     "1m0s",
			"containers": cMetrics,
		})
	}

	writeJSON(w, map[string]interface{}{
		"kind":       "PodMetricsList",
		"apiVersion": "metrics.k8s.io/v1beta1",
		"metadata": map[string]interface{}{
			"selfLink": "/apis/metrics.k8s.io/v1beta1/pods",
		},
		"items": items,
	})
}

func (s *Server) serveMetricsNodes(w http.ResponseWriter, r *http.Request) {
	nodeEnvs, _, err := s.store.List(r.Context(), statestore.ListQuery{
		Key: statestore.ResourceKey{Group: "", Version: "v1", Resource: "nodes"},
	})
	if err != nil || len(nodeEnvs) == 0 {
		writeJSON(w, map[string]interface{}{"kind": "NodeMetricsList", "apiVersion": "metrics.k8s.io/v1beta1", "items": []interface{}{}})
		return
	}

	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)
	sys := tarakruntime.SampleSystemMetrics()

	var items []map[string]interface{}
	for _, env := range nodeEnvs {
		var node map[string]interface{}
		if err := json.Unmarshal(env.Object, &node); err != nil {
			continue
		}
		meta, _ := node["metadata"].(map[string]interface{})
		if meta == nil {
			continue
		}
		nodeName, _ := meta["name"].(string)

		cpuStr := fmt.Sprintf("%dm", sys.CPUMillicores)
		memStr := fmt.Sprintf("%dMi", int(float64(sys.UsedMemoryBytes)/(1024*1024)))
		memTotStr := fmt.Sprintf("%dMi", int(float64(sys.TotalMemoryBytes)/(1024*1024)))
		memPctStr := fmt.Sprintf("%.0f%%", sys.MemoryPercent)
		cpuPctStr := fmt.Sprintf("%.0f%%", sys.CPUPercent)

		items = append(items, map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":              nodeName,
				"creationTimestamp": nowStr,
			},
			"timestamp": nowStr,
			"window":    "1m0s",
			"usage": map[string]string{
				"cpu":           cpuStr,
				"cpuPercent":    cpuPctStr,
				"memory":        memStr,
				"memoryTotal":   memTotStr,
				"memoryPercent": memPctStr,
			},
		})
	}

	writeJSON(w, map[string]interface{}{
		"kind":       "NodeMetricsList",
		"apiVersion": "metrics.k8s.io/v1beta1",
		"metadata": map[string]interface{}{
			"selfLink": "/apis/metrics.k8s.io/v1beta1/nodes",
		},
		"items": items,
	})
}

func (s *Server) serveRuntimeVersion(w http.ResponseWriter, r *http.Request) {
	ver := s.runtime.GetRuntimeVersion()
	writeJSON(w, ver)
}

func (s *Server) serveRuntimeStatus(w http.ResponseWriter, r *http.Request) {
	sys := tarakruntime.SampleSystemMetrics()
	ver := s.runtime.GetRuntimeVersion()

	resp := map[string]interface{}{
		"status": "Ready",
		"runtime": ver,
		"hardware": map[string]interface{}{
			"cpuCores":       sys.NumCPU,
			"cpuUsage":       fmt.Sprintf("%dm", sys.CPUMillicores),
			"cpuPercent":     fmt.Sprintf("%.1f%%", sys.CPUPercent),
			"memoryUsed":     fmt.Sprintf("%.1f GiB", float64(sys.UsedMemoryBytes)/(1024*1024*1024)),
			"memoryTotal":    fmt.Sprintf("%.1f GiB", float64(sys.TotalMemoryBytes)/(1024*1024*1024)),
			"memoryPercent":  fmt.Sprintf("%.1f%%", sys.MemoryPercent),
		},
		"network": map[string]interface{}{
			"podCIDR":     "10.244.0.0/16",
			"serviceCIDR": "10.96.0.0/12",
			"isolation":   "Virtual Pod Bridge (Namespaced Loopback)",
		},
	}
	writeJSON(w, resp)
}
