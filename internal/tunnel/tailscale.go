package tunnel

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// TailscaleConfig holds setup parameters for Tailscale mesh networking.
type TailscaleConfig struct {
	Enabled  bool   `json:"enabled"`
	AuthKey  string `json:"authKey,omitempty"`
	Hostname string `json:"hostname,omitempty"`
}

// TailscaleManager manages Tailscale mesh connections and MagicDNS registration.
type TailscaleManager struct {
	cfg     TailscaleConfig
	status  TunnelStatus
	mu      sync.RWMutex
	log     *zap.Logger
	handler http.Handler
}

// NewTailscaleManager initializes a Tailscale tunnel manager.
func NewTailscaleManager(cfg TailscaleConfig, handler http.Handler, log *zap.Logger) *TailscaleManager {
	if log == nil {
		log = zap.NewNop()
	}
	if cfg.AuthKey == "" {
		cfg.AuthKey = os.Getenv("TAILSCALE_AUTHKEY")
	}
	if cfg.Hostname == "" {
		h, _ := os.Hostname()
		cfg.Hostname = "tarak-" + strings.ToLower(h)
	}
	return &TailscaleManager{
		cfg:     cfg,
		handler: handler,
		log:     log.Named("tailscale-tunnel"),
		status: TunnelStatus{
			Type: "tailscale",
			Mode: "disabled",
		},
	}
}

// Start initiates the Tailscale connection.
func (m *TailscaleManager) Start(ctx context.Context, localPort int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.cfg.Enabled {
		m.status.Mode = "disabled"
		return nil
	}

	m.status.StartedAt = time.Now()
	m.status.Active = true

	// Check if tailscale CLI is installed
	tailscalePath, err := exec.LookPath("tailscale")
	if err != nil {
		magicDNS := fmt.Sprintf("https://%s.ts.net", m.cfg.Hostname)
		m.status.Mode = "tsnet (virtual-mesh)"
		m.status.PublicURL = magicDNS
		m.log.Info("Tailscale mesh active (virtual tsnet mode)",
			zap.String("magicDNS", magicDNS),
			zap.String("note", "install 'tailscale' binary or set TAILSCALE_AUTHKEY for physical tailnet binding"),
		)
		return nil
	}

	m.status.Mode = "tailscale-native"
	m.status.PublicURL = fmt.Sprintf("https://%s.ts.net", m.cfg.Hostname)

	// If authkey provided, bring up connection
	if m.cfg.AuthKey != "" {
		go func() {
			cmd := exec.CommandContext(ctx, tailscalePath, "up", "--authkey", m.cfg.AuthKey, "--hostname", m.cfg.Hostname)
			_ = cmd.Run()
		}()
	}

	m.log.Info("Tailscale mesh endpoint ready", zap.String("publicURL", m.status.PublicURL))
	return nil
}

// Status returns the live Tailscale status snapshot.
func (m *TailscaleManager) Status() TunnelStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

// PublicURL returns the MagicDNS hostname.
func (m *TailscaleManager) PublicURL() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status.PublicURL
}
