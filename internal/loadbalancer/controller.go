package loadbalancer

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// Controller integrates the detector, pool, and forwarder into a unified bare-metal LB manager.
type Controller struct {
	log       *zap.Logger
	detector  *IPDetector
	pool      *IPPool
	forwarder *Forwarder
	publicIP  string
	lanIP     string
}

// NewController initializes a new bare-metal load balancer controller.
func NewController(log *zap.Logger) *Controller {
	cLog := log.Named("loadbalancer")
	detector := NewIPDetector(cLog)
	lanIP := detector.DetectLocalLANIP()

	// Initial pool with local LAN IP
	pool := NewIPPool("", lanIP, cLog)
	forwarder := NewForwarder(cLog)

	c := &Controller{
		log:       cLog,
		detector:  detector,
		pool:      pool,
		forwarder: forwarder,
		lanIP:     lanIP,
	}

	return c
}

// Start launches background public IP discovery and reconciler loops.
func (c *Controller) Start(ctx context.Context) {
	// Async public IP detection so startup is instantaneous
	go func() {
		pubIP := c.detector.DetectPublicIP(ctx)
		c.publicIP = pubIP
		c.pool = NewIPPool(pubIP, c.lanIP, c.log)
		c.log.Info("bare-metal loadbalancer initialized",
			zap.String("publicIP", pubIP),
			zap.String("lanIP", c.lanIP),
		)
	}()
}

// HandleStatus returns the live loadbalancer status and IP assignments for the API/UI.
func (c *Controller) HandleStatus(w http.ResponseWriter, r *http.Request) {
	status := c.pool.GetStatus()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf(`{
		"publicIP": %q,
		"lanIP": %q,
		"status": "Healthy",
		"allocatedVIPs": %d,
		"supportedIngressClasses": ["tarak", "nginx", "traefik", "kong", "kuma", "custom"]
	}`, c.publicIP, c.lanIP, status["allocatedCount"])))
}

// AllocateVIP assigns a VIP for a given service key.
func (c *Controller) AllocateVIP(serviceKey string, preferPublic bool) (string, error) {
	return c.pool.Allocate(serviceKey, preferPublic)
}

// ReleaseVIP releases an assigned VIP.
func (c *Controller) ReleaseVIP(serviceKey string) {
	c.pool.Release(serviceKey)
}

// SyncServiceForwarding updates the live TCP proxy listener for a service or nodePort.
func (c *Controller) SyncServiceForwarding(ctx context.Context, listenAddr string, endpoints []Endpoint) error {
	return c.forwarder.UpdateServiceRoutes(ctx, listenAddr, endpoints)
}

// PublicIP returns the active detected public WAN IP.
func (c *Controller) PublicIP() string {
	if c.publicIP != "" {
		return c.publicIP
	}
	return c.lanIP
}
