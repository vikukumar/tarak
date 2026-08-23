package network

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	DefaultPodCIDR     = "10.244.0.0/16"
	DefaultServiceCIDR = "10.96.0.0/12"
	DefaultBridgeName  = "tarak-br0"
)

// BridgeConfig contains options for the Tarak network bridge driver.
type BridgeConfig struct {
	PodCIDR     string
	ServiceCIDR string
	BridgeName  string
	EnablemTLS  bool
}

// HostNetworkInfo contains captured network IPs and interface details of the host machine.
type HostNetworkInfo struct {
	PublicIP       string   `json:"publicIP"`       // Captured public WAN IP
	PrimaryLANIP   string   `json:"primaryLANIP"`   // Primary host private LAN IPv4 (e.g. 192.168.1.50)
	AllHostIPs     []string `json:"allHostIPs"`     // All non-loopback host IPv4s
	PodCIDR        string   `json:"podCIDR"`        // 10.244.0.0/16
	ServiceCIDR    string   `json:"serviceCIDR"`    // 10.96.0.0/12
	BridgeActive   bool     `json:"bridgeActive"`   // True when bridge routing is active
	HostAccessMode string   `json:"hostAccessMode"` // "DirectHostBridge"
}

// Driver manages host-to-cluster and cluster-to-host network routing and bridge connections.
type Driver struct {
	cfg       BridgeConfig
	log       *zap.Logger
	mu        sync.RWMutex
	info      HostNetworkInfo
	routes    map[string]string // PodIP -> HostPort target or Endpoint
	listeners map[string]net.Listener
	closed    bool
}

// NewDriver initializes the Tarak native network bridge driver.
func NewDriver(cfg BridgeConfig, log *zap.Logger) *Driver {
	if cfg.PodCIDR == "" {
		cfg.PodCIDR = DefaultPodCIDR
	}
	if cfg.ServiceCIDR == "" {
		cfg.ServiceCIDR = DefaultServiceCIDR
	}
	if cfg.BridgeName == "" {
		cfg.BridgeName = DefaultBridgeName
	}
	if log == nil {
		log = zap.NewNop()
	}

	d := &Driver{
		cfg:       cfg,
		log:       log.Named("network-driver"),
		routes:    make(map[string]string),
		listeners: make(map[string]net.Listener),
	}

	d.refreshHostNetwork(context.Background())
	return d
}

// Start activates the host bridge connection, enabling direct host-to-pod routing.
func (d *Driver) Start(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.log.Info("initializing native host network bridge driver",
		zap.String("bridge", d.cfg.BridgeName),
		zap.String("podCIDR", d.cfg.PodCIDR),
		zap.String("serviceCIDR", d.cfg.ServiceCIDR),
		zap.String("lanIP", d.info.PrimaryLANIP),
		zap.String("publicIP", d.info.PublicIP),
	)

	// Platform-specific bridge and direct route initialization
	if err := d.setupHostBridgePlatform(); err != nil {
		d.log.Warn("platform host bridge setup note (fallback active)", zap.Error(err))
	}

	d.info.BridgeActive = true
	d.info.HostAccessMode = "DirectHostBridge"
	return nil
}

// Stop gracefully tears down bridge routing listeners.
func (d *Driver) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.closed = true
	for ip, ln := range d.listeners {
		_ = ln.Close()
		delete(d.listeners, ip)
	}
	d.teardownHostBridgePlatform()
	d.info.BridgeActive = false
	d.log.Info("tarak network bridge driver stopped")
	return nil
}

// GetHostNetworkInfo returns the current captured host network topology.
func (d *Driver) GetHostNetworkInfo() HostNetworkInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.info
}

// RegisterPodRoute registers a Pod IP with its physical host/container port for transparent host routing.
func (d *Driver) RegisterPodRoute(podIP, targetAddr string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.routes[podIP] = targetAddr
	d.log.Debug("registered pod host bridge route", zap.String("podIP", podIP), zap.String("target", targetAddr))
}

// UnregisterPodRoute removes a Pod IP route.
func (d *Driver) UnregisterPodRoute(podIP string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.routes, podIP)
}

// AllocatePodIP allocates a deterministic Pod IP from the Pod CIDR for a given index.
func (d *Driver) AllocatePodIP(nodeIdx, podIdx int) string {
	thirdOctet := (nodeIdx % 250)
	fourthOctet := ((podIdx % 250) + 2)
	return fmt.Sprintf("10.244.%d.%d", thirdOctet, fourthOctet)
}

// refreshHostNetwork queries the host's actual network interfaces and public IP.
func (d *Driver) refreshHostNetwork(ctx context.Context) {
	// 1. Detect LAN IPs
	var hostIPs []string
	primaryLAN := "127.0.0.1"

	if ifaces, err := net.Interfaces(); err == nil {
		for _, iface := range ifaces {
			if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
				continue
			}
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if ip == nil || ip.IsLoopback() || ip.To4() == nil {
					continue
				}
				ipStr := ip.String()
				hostIPs = append(hostIPs, ipStr)
				if primaryLAN == "127.0.0.1" {
					primaryLAN = ipStr
				}
			}
		}
	}

	// 2. Detect Public WAN IP
	publicIP := d.detectPublicWANIP(ctx)
	if publicIP == "" {
		publicIP = primaryLAN
	}

	d.info = HostNetworkInfo{
		PublicIP:       publicIP,
		PrimaryLANIP:   primaryLAN,
		AllHostIPs:     hostIPs,
		PodCIDR:        d.cfg.PodCIDR,
		ServiceCIDR:    d.cfg.ServiceCIDR,
		BridgeActive:   false,
		HostAccessMode: "DirectHostBridge",
	}
}

// detectPublicWANIP queries external STUN/HTTP endpoints to resolve the host's public WAN IP.
func (d *Driver) detectPublicWANIP(ctx context.Context) string {
	endpoints := []string{
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
		"https://checkip.amazonaws.com",
	}

	client := &http.Client{
		Timeout: 2000 * time.Millisecond,
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
			return ipStr
		}
	}
	return ""
}
