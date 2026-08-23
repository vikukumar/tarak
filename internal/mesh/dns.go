package mesh

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

// DNSResolver provides virtual .mesh DNS resolution for mesh workloads.
type DNSResolver struct {
	mu      sync.RWMutex
	records map[string]string // key: hostname (e.g. "order-service.default.mesh"), val: VIP
	nextVIP uint32
}

// NewDNSResolver creates a new .mesh DNS resolver.
func NewDNSResolver() *DNSResolver {
	return &DNSResolver{
		records: make(map[string]string),
		nextVIP: 1,
	}
}

// RegisterService generates and assigns virtual .mesh hostnames and a VIP for a service.
func (d *DNSResolver) RegisterService(meshName, namespace, serviceName string) (string, []string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Virtual VIP allocation from 240.240.x.x range (CGNAT / Virtual Mesh Range)
	vip := fmt.Sprintf("240.240.%d.%d", (d.nextVIP>>8)&0xFF, d.nextVIP&0xFF)
	d.nextVIP++

	hostnames := []string{
		fmt.Sprintf("%s.%s.mesh", serviceName, namespace),
		fmt.Sprintf("%s.%s.%s.mesh", serviceName, namespace, meshName),
		fmt.Sprintf("%s.mesh", serviceName),
	}

	for _, h := range hostnames {
		d.records[strings.ToLower(h)] = vip
	}

	return vip, hostnames
}

// Resolve looks up a .mesh hostname to its virtual VIP.
func (d *DNSResolver) Resolve(hostname string) (string, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	vip, ok := d.records[strings.ToLower(strings.TrimSuffix(hostname, "."))]
	return vip, ok
}

// LookupIP performs standard net.IP lookup for .mesh domains.
func (d *DNSResolver) LookupIP(hostname string) net.IP {
	if vip, ok := d.Resolve(hostname); ok {
		return net.ParseIP(vip)
	}
	return nil
}
