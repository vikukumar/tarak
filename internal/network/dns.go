// Package network implements embedded micro-CoreDNS service discovery for Tarak.
package network

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DNSRecord represents an A or CNAME DNS mapping.
type DNSRecord struct {
	Name string
	IP   net.IP
	TTL  uint32
}

// MicroCoreDNS is an embedded, lightweight RFC-1035 UDP/TCP DNS server for Kubernetes service discovery.
type MicroCoreDNS struct {
	mu        sync.RWMutex
	log       *zap.Logger
	listenIP  string
	port      int
	domain    string // default "cluster.local"
	records   map[string]net.IP
	upstream  string // e.g. "8.8.8.8:53"
	udpConn   *net.UDPConn
	closed    bool
}

// NewMicroCoreDNS creates a new micro-CoreDNS server.
func NewMicroCoreDNS(listenIP string, port int, domain string, log *zap.Logger) *MicroCoreDNS {
	if listenIP == "" {
		listenIP = "0.0.0.0"
	}
	if port == 0 {
		port = 5353
	}
	if domain == "" {
		domain = "cluster.local"
	}
	if log == nil {
		log = zap.NewNop()
	}

	return &MicroCoreDNS{
		log:      log.Named("micro-coredns"),
		listenIP: listenIP,
		port:     port,
		domain:   domain,
		records:  make(map[string]net.IP),
		upstream: "8.8.8.8:53",
	}
}

// RegisterService adds a service name mapping (e.g. "demo-service.default.svc.cluster.local" -> 10.96.0.149).
func (d *MicroCoreDNS) RegisterService(namespace, serviceName, clusterIP string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	ip := net.ParseIP(clusterIP)
	if ip == nil {
		return
	}

	// 1. Full FQDN: <svc>.<ns>.svc.cluster.local
	fqdn := strings.ToLower(fmt.Sprintf("%s.%s.svc.%s", serviceName, namespace, d.domain))
	d.records[fqdn] = ip

	// 2. Short name within namespace: <svc>.<ns>
	short := strings.ToLower(fmt.Sprintf("%s.%s", serviceName, namespace))
	d.records[short] = ip

	// 3. Simple service name: <svc>
	d.records[strings.ToLower(serviceName)] = ip

	d.log.Debug("registered service DNS record", zap.String("name", fqdn), zap.String("ip", clusterIP))
}

// UnregisterService removes a service DNS mapping.
func (d *MicroCoreDNS) UnregisterService(namespace, serviceName string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	fqdn := strings.ToLower(fmt.Sprintf("%s.%s.svc.%s", serviceName, namespace, d.domain))
	short := strings.ToLower(fmt.Sprintf("%s.%s", serviceName, namespace))
	delete(d.records, fqdn)
	delete(d.records, short)
	delete(d.records, strings.ToLower(serviceName))
}

// Start launches the UDP DNS listener.
func (d *MicroCoreDNS) Start(ctx context.Context) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", d.listenIP, d.port))
	if err != nil {
		return fmt.Errorf("resolve DNS addr: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		// Fallback to ephemeral port if 53/5353 is busy
		conn, err = net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(d.listenIP), Port: 0})
		if err != nil {
			return fmt.Errorf("listen DNS UDP: %w", err)
		}
	}
	d.udpConn = conn

	d.log.Info("micro-CoreDNS server started",
		zap.String("addr", conn.LocalAddr().String()),
		zap.String("domain", d.domain),
	)

	go d.serveUDP(ctx)
	return nil
}

func (d *MicroCoreDNS) serveUDP(ctx context.Context) {
	buf := make([]byte, 512)
	for {
		n, remoteAddr, err := d.udpConn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				if d.closed {
					return
				}
				continue
			}
		}

		go d.handleQuery(buf[:n], remoteAddr)
	}
}

func (d *MicroCoreDNS) handleQuery(req []byte, remote *net.UDPAddr) {
	if len(req) < 12 {
		return
	}

	// Extract requested query domain name (simplified RFC 1035 wire parser)
	queryName := d.parseQueryName(req[12:])
	if queryName == "" {
		return
	}

	d.mu.RLock()
	ip, found := d.records[strings.ToLower(queryName)]
	d.mu.RUnlock()

	if !found {
		// Try upstream forwarder
		d.forwardUpstream(req, remote)
		return
	}

	// Build A-record response
	resp := make([]byte, len(req)+16)
	copy(resp, req)
	// Flags: QR=1 (response), AA=1, RA=1
	resp[2] = 0x81
	resp[3] = 0x80
	// Answer count: 1
	resp[6] = 0x00
	resp[7] = 0x01

	// Answer section: Name offset pointer
	ansOffset := len(req)
	resp[ansOffset] = 0xc0
	resp[ansOffset+1] = 0x0c
	// Type: A (1)
	resp[ansOffset+2] = 0x00
	resp[ansOffset+3] = 0x01
	// Class: IN (1)
	resp[ansOffset+4] = 0x00
	resp[ansOffset+5] = 0x01
	// TTL: 30s
	resp[ansOffset+6] = 0x00
	resp[ansOffset+7] = 0x00
	resp[ansOffset+8] = 0x00
	resp[ansOffset+9] = 0x1e
	// Data length: 4
	resp[ansOffset+10] = 0x00
	resp[ansOffset+11] = 0x04
	// IPv4 bytes
	ipv4 := ip.To4()
	if ipv4 != nil {
		copy(resp[ansOffset+12:], ipv4)
	}

	_, _ = d.udpConn.WriteToUDP(resp, remote)
}

func (d *MicroCoreDNS) parseQueryName(buf []byte) string {
	var parts []string
	idx := 0
	for idx < len(buf) {
		length := int(buf[idx])
		if length == 0 {
			break
		}
		idx++
		if idx+length > len(buf) {
			return ""
		}
		parts = append(parts, string(buf[idx:idx+length]))
		idx += length
	}
	return strings.Join(parts, ".")
}

func (d *MicroCoreDNS) forwardUpstream(req []byte, remote *net.UDPAddr) {
	upConn, err := net.DialTimeout("udp", d.upstream, 1*time.Second)
	if err != nil {
		return
	}
	defer upConn.Close()

	_, _ = upConn.Write(req)
	buf := make([]byte, 512)
	_ = upConn.SetReadDeadline(time.Now().Add(1 * time.Second))
	n, err := upConn.Read(buf)
	if err == nil && n > 0 {
		_, _ = d.udpConn.WriteToUDP(buf[:n], remote)
	}
}

// Stop terminates the micro-CoreDNS server.
func (d *MicroCoreDNS) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.closed = true
	if d.udpConn != nil {
		_ = d.udpConn.Close()
	}
}
