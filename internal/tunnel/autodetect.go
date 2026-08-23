package tunnel

import (
	"context"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"go.uber.org/zap"
)

// HostTunnelDetection holds auto-detected tunnel capabilities on the host system.
type HostTunnelDetection struct {
	CloudflareDetected bool   `json:"cloudflareDetected"`
	CloudflareBinPath  string `json:"cloudflareBinPath,omitempty"`
	CloudflareToken    string `json:"cloudflareToken,omitempty"`
	CloudflareConfig   string `json:"cloudflareConfig,omitempty"`

	TailscaleDetected  bool   `json:"tailscaleDetected"`
	TailscaleBinPath   string `json:"tailscaleBinPath,omitempty"`
	TailscaleIP        string `json:"tailscaleIP,omitempty"`
	TailscaleMagicDNS  string `json:"tailscaleMagicDNS,omitempty"`
	TailscaleAuthKey   string `json:"tailscaleAuthKey,omitempty"`
	TailscaleSocket    string `json:"tailscaleSocket,omitempty"`
}

// DetectHostTunnels inspects the host OS for installed Cloudflare and Tailscale agents,
// network interfaces, system services, and configuration files.
func DetectHostTunnels(log *zap.Logger) HostTunnelDetection {
	if log == nil {
		log = zap.NewNop()
	}

	det := HostTunnelDetection{}

	// ─── 1. Cloudflare Detection ──────────────────────────────────────────────
	// Check PATH
	if p, err := exec.LookPath("cloudflared"); err == nil {
		det.CloudflareDetected = true
		det.CloudflareBinPath = p
	} else if runtime.GOOS == "windows" {
		// Common Windows install directories
		winPaths := []string{
			`C:\Program Files\cloudflared\cloudflared.exe`,
			`C:\Program Files (x86)\cloudflared\cloudflared.exe`,
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "cloudflared", "cloudflared.exe"),
		}
		for _, wp := range winPaths {
			if _, err := os.Stat(wp); err == nil {
				det.CloudflareDetected = true
				det.CloudflareBinPath = wp
				break
			}
		}
	}

	// Check environment token
	if tok := os.Getenv("CLOUDFLARE_TUNNEL_TOKEN"); tok != "" {
		det.CloudflareToken = tok
		det.CloudflareDetected = true
	}

	// Check common config files
	cfConfigs := []string{
		"/etc/cloudflared/config.yml",
		"/etc/cloudflared/config.yaml",
		filepath.Join(os.Getenv("HOME"), ".cloudflared", "config.yml"),
		filepath.Join(os.Getenv("USERPROFILE"), ".cloudflared", "config.yml"),
	}
	for _, cfc := range cfConfigs {
		if _, err := os.Stat(cfc); err == nil {
			det.CloudflareConfig = cfc
			det.CloudflareDetected = true
			break
		}
	}

	// ─── 2. Tailscale Detection ───────────────────────────────────────────────
	// Check PATH
	if p, err := exec.LookPath("tailscale"); err == nil {
		det.TailscaleDetected = true
		det.TailscaleBinPath = p
	} else if runtime.GOOS == "windows" {
		winPaths := []string{
			`C:\Program Files\Tailscale\tailscale.exe`,
			`C:\Program Files (x86)\Tailscale\tailscale.exe`,
		}
		for _, wp := range winPaths {
			if _, err := os.Stat(wp); err == nil {
				det.TailscaleDetected = true
				det.TailscaleBinPath = wp
				break
			}
		}
	}

	// Check Auth Key
	if key := os.Getenv("TAILSCALE_AUTHKEY"); key != "" {
		det.TailscaleAuthKey = key
		det.TailscaleDetected = true
	}

	// Check Tailscale daemon sockets & service files
	tsSockets := []string{
		"/var/run/tailscale/tailscaled.sock",
		"/run/tailscale/tailscaled.sock",
	}
	for _, sock := range tsSockets {
		if _, err := os.Stat(sock); err == nil {
			det.TailscaleSocket = sock
			det.TailscaleDetected = true
			break
		}
	}

	// Check Network Interfaces for Tailscale 100.64.0.0/10 CGNAT IP
	if ifaces, err := net.Interfaces(); err == nil {
		for _, iface := range ifaces {
			nameLower := strings.ToLower(iface.Name)
			if strings.Contains(nameLower, "tailscale") || strings.Contains(nameLower, "utun") || strings.Contains(nameLower, "wg") {
				if addrs, err := iface.Addrs(); err == nil {
					for _, addr := range addrs {
						if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
							if ip4 := ipnet.IP.To4(); ip4 != nil {
								// Tailscale CGNAT range is 100.64.0.0/10 (100.64.0.0 - 100.127.255.255)
								if ip4[0] == 100 && (ip4[1]&0xC0) == 64 {
									det.TailscaleDetected = true
									det.TailscaleIP = ip4.String()
									break
								}
							}
						}
					}
				}
			}
		}
	}

	// Query tailscale status if CLI is available
	if det.TailscaleBinPath != "" && det.TailscaleIP == "" {
		out, err := exec.Command(det.TailscaleBinPath, "ip", "-4").Output()
		if err == nil && len(out) > 0 {
			det.TailscaleIP = strings.TrimSpace(string(out))
			det.TailscaleDetected = true
		}
	}

	if det.CloudflareDetected {
		log.Info("🔍 auto-detected Cloudflare Tunnel on host",
			zap.String("bin", det.CloudflareBinPath),
			zap.Bool("hasToken", det.CloudflareToken != ""),
		)
	}
	if det.TailscaleDetected {
		log.Info("🔍 auto-detected Tailscale WireGuard Mesh on host",
			zap.String("bin", det.TailscaleBinPath),
			zap.String("tailscaleIP", det.TailscaleIP),
		)
	}

	return det
}

// AutoRegisterNodeTunnels syncs auto-detected tunnel endpoints when a node registers.
func AutoRegisterNodeTunnels(ctx context.Context, det HostTunnelDetection, nodeName string, log *zap.Logger) map[string]string {
	endpoints := make(map[string]string)
	if det.CloudflareDetected {
		if det.CloudflareToken != "" {
			endpoints["cloudflare"] = "named-tunnel"
		} else {
			endpoints["cloudflare"] = "quick-tunnel"
		}
	}
	if det.TailscaleDetected {
		if det.TailscaleIP != "" {
			endpoints["tailscale"] = det.TailscaleIP
		} else {
			endpoints["tailscale"] = "auto-mesh"
		}
	}
	return endpoints
}
