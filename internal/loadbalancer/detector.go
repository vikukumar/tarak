package loadbalancer

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// IPDetector detects WAN/Public IP as well as local LAN IPv4 addresses.
type IPDetector struct {
	log *zap.Logger
}

// NewIPDetector creates a new IP detector.
func NewIPDetector(log *zap.Logger) *IPDetector {
	return &IPDetector{
		log: log.Named("ip-detector"),
	}
}

// DetectPublicIP queries fast external STUN/HTTP endpoints to resolve the cluster's public WAN IP.
func (d *IPDetector) DetectPublicIP(ctx context.Context) string {
	endpoints := []string{
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
		"https://checkip.amazonaws.com",
	}

	client := &http.Client{
		Timeout: 2500 * time.Millisecond,
	}

	for _, ep := range endpoints {
		req, err := http.NewRequestWithContext(ctx, "GET", ep, nil)
		if err != nil {
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			continue
		}

		ipStr := strings.TrimSpace(string(body))
		if parsed := net.ParseIP(ipStr); parsed != nil && !parsed.IsLoopback() && !parsed.IsPrivate() {
			d.log.Info("auto-detected public WAN IP", zap.String("publicIP", ipStr), zap.String("source", ep))
			return ipStr
		}
	}

	// Fallback to primary local LAN IP
	lanIP := d.DetectLocalLANIP()
	d.log.Info("using local LAN IP as fallback", zap.String("lanIP", lanIP))
	return lanIP
}

// DetectLocalLANIP returns the first non-loopback private IPv4 address.
func (d *IPDetector) DetectLocalLANIP() string {
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}
	return "127.0.0.1"
}
