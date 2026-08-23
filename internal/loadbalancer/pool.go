package loadbalancer

import (
	"net"
	"sync"

	"go.uber.org/zap"
)

// IPPool manages an address pool for bare-metal LoadBalancer Services.
type IPPool struct {
	mu           sync.RWMutex
	log          *zap.Logger
	publicIP     string
	lanIP        string
	allocated    map[string]string // key: serviceKey (ns/name), val: assignedIP
	reverseAlloc map[string]string // key: IP, val: serviceKey
	cidrPool     []string          // list of allocatable pool IPs
}

// NewIPPool creates an IPPool with public and LAN addresses.
func NewIPPool(publicIP, lanIP string, log *zap.Logger) *IPPool {
	p := &IPPool{
		log:          log.Named("ip-pool"),
		publicIP:     publicIP,
		lanIP:        lanIP,
		allocated:    make(map[string]string),
		reverseAlloc: make(map[string]string),
	}

	// Auto-populate a default VIP pool in the LAN subnet range (e.g. .200-.250)
	p.generateDefaultSubnetPool(lanIP)
	return p
}

func (p *IPPool) generateDefaultSubnetPool(lanIP string) {
	ip := net.ParseIP(lanIP).To4()
	if ip == nil {
		p.cidrPool = []string{lanIP, "127.0.0.1"}
		return
	}

	// Generate 30 VIP slots in the subnet
	for i := 200; i <= 230; i++ {
		vip := net.IPv4(ip[0], ip[1], ip[2], byte(i)).String()
		p.cidrPool = append(p.cidrPool, vip)
	}
}

// Allocate assigns a VIP for a given service. If public is requested, returns the public WAN IP.
func (p *IPPool) Allocate(serviceKey string, preferPublic bool) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if already allocated
	if existing, ok := p.allocated[serviceKey]; ok {
		return existing, nil
	}

	if preferPublic && p.publicIP != "" && p.publicIP != "127.0.0.1" {
		p.allocated[serviceKey] = p.publicIP
		p.reverseAlloc[p.publicIP] = serviceKey
		p.log.Info("allocated public WAN VIP", zap.String("service", serviceKey), zap.String("ip", p.publicIP))
		return p.publicIP, nil
	}

	// Find first free IP from pool
	for _, ip := range p.cidrPool {
		if _, taken := p.reverseAlloc[ip]; !taken {
			p.allocated[serviceKey] = ip
			p.reverseAlloc[ip] = serviceKey
			p.log.Info("allocated pool VIP", zap.String("service", serviceKey), zap.String("ip", ip))
			return ip, nil
		}
	}

	// Fallback to LAN IP
	fallback := p.lanIP
	if fallback == "" {
		fallback = "127.0.0.1"
	}
	p.allocated[serviceKey] = fallback
	return fallback, nil
}

// Release frees the VIP assigned to a service.
func (p *IPPool) Release(serviceKey string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if ip, ok := p.allocated[serviceKey]; ok {
		delete(p.allocated, serviceKey)
		delete(p.reverseAlloc, ip)
		p.log.Info("released VIP lease", zap.String("service", serviceKey), zap.String("ip", ip))
	}
}

// GetStatus returns the current VIP allocation summary.
func (p *IPPool) GetStatus() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"publicIP":       p.publicIP,
		"lanIP":          p.lanIP,
		"totalPoolSlots": len(p.cidrPool),
		"allocatedCount": len(p.allocated),
		"activeBindings": p.allocated,
	}
}
