package tunnel

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CloudflareTunnelConfig holds setup parameters for Cloudflare tunneling.
type CloudflareTunnelConfig struct {
	Enabled   bool   `json:"enabled"`
	Token     string `json:"token,omitempty"`
	LocalPort int    `json:"localPort"`
}

// TunnelStatus represents the live operational status of a tunnel.
type TunnelStatus struct {
	Type        string    `json:"type"`
	Active      bool      `json:"active"`
	PublicURL   string    `json:"publicURL,omitempty"`
	StartedAt   time.Time `json:"startedAt,omitempty"`
	Mode        string    `json:"mode"`
	LastError   string    `json:"lastError,omitempty"`
}

// CloudflareManager manages the lifecycle of Cloudflare Tunnels (Quick and Named).
type CloudflareManager struct {
	cfg     CloudflareTunnelConfig
	status  TunnelStatus
	cmd     *exec.Cmd
	mu      sync.RWMutex
	log     *zap.Logger
	handler http.Handler
}

// NewCloudflareManager initializes a Cloudflare tunnel manager.
func NewCloudflareManager(cfg CloudflareTunnelConfig, handler http.Handler, log *zap.Logger) *CloudflareManager {
	if log == nil {
		log = zap.NewNop()
	}
	if cfg.Token == "" {
		cfg.Token = os.Getenv("CLOUDFLARE_TUNNEL_TOKEN")
	}
	return &CloudflareManager{
		cfg:     cfg,
		handler: handler,
		log:     log.Named("cloudflare-tunnel"),
		status: TunnelStatus{
			Type: "cloudflare",
			Mode: "disabled",
		},
	}
}

// Start launches the Cloudflare tunnel in the background.
func (m *CloudflareManager) Start(ctx context.Context, localServerAddr string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.cfg.Enabled {
		m.status.Mode = "disabled"
		return nil
	}

	m.status.StartedAt = time.Now()
	m.status.Active = true

	// Check if cloudflared binary is available on host
	cloudflaredPath, err := exec.LookPath("cloudflared")
	if err != nil {
		m.status.Mode = "binary-missing"
		m.status.Active = false
		m.log.Warn("cloudflared binary not found on host path; install 'cloudflared' to enable live Cloudflare ingress tunneling")
		return fmt.Errorf("cloudflared binary not found in PATH")
	}

	var args []string
	if m.cfg.Token != "" {
		m.status.Mode = "named-tunnel"
		args = []string{"tunnel", "run", "--token", m.cfg.Token}
		m.log.Info("launching Cloudflare Named Tunnel", zap.String("mode", m.status.Mode))
	} else {
		m.status.Mode = "quick-tunnel"
		args = []string{"tunnel", "--url", "http://" + localServerAddr}
		m.log.Info("launching Cloudflare Quick Tunnel (auto-URL mode)", zap.String("target", localServerAddr))
	}

	cmd := exec.CommandContext(ctx, cloudflaredPath, args...)
	m.cmd = cmd

	// Capture output to parse quick tunnel URL
	stdoutPipe, err := cmd.StderrPipe()
	if err == nil {
		go func() {
			buf := make([]byte, 2048)
			urlRegex := regexp.MustCompile(`https://[a-zA-Z0-9-]+\.trycloudflare\.com`)
			for {
				n, err := stdoutPipe.Read(buf)
				if n > 0 {
					text := string(buf[:n])
					if matches := urlRegex.FindString(text); matches != "" {
						m.mu.Lock()
						m.status.PublicURL = matches
						m.log.Info("⚡ Cloudflare Quick Tunnel ready!", zap.String("publicURL", matches))
						m.mu.Unlock()
					}
				}
				if err != nil {
					break
				}
			}
		}()
	}

	if err := cmd.Start(); err != nil {
		m.status.Active = false
		m.status.LastError = err.Error()
		m.log.Error("failed to start cloudflared tunnel", zap.Error(err))
		return err
	}

	return nil
}

// Status returns current tunnel status snapshot.
func (m *CloudflareManager) Status() TunnelStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

// PublicURL returns the active public hostname.
func (m *CloudflareManager) PublicURL() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status.PublicURL
}
