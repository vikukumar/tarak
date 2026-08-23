package loadbalancer

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// Endpoint represents a backend pod or service destination.
type Endpoint struct {
	Address string
	Port    int
	Healthy bool
}

// Forwarder manages L4 TCP proxy listeners for LoadBalancer services.
type Forwarder struct {
	mu        sync.RWMutex
	log       *zap.Logger
	listeners map[string]net.Listener // key: listenAddr (e.g. 0.0.0.0:8080)
	endpoints map[string][]Endpoint   // key: listenAddr -> endpoints
	rrCounter map[string]*uint64      // round-robin counters
}

// NewForwarder creates a new L4/L7 forwarder.
func NewForwarder(log *zap.Logger) *Forwarder {
	return &Forwarder{
		log:       log.Named("lb-forwarder"),
		listeners: make(map[string]net.Listener),
		endpoints: make(map[string][]Endpoint),
		rrCounter: make(map[string]*uint64),
	}
}

// UpdateServiceRoutes updates the active backend endpoints for a listening port.
func (f *Forwarder) UpdateServiceRoutes(ctx context.Context, listenAddr string, endpoints []Endpoint) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.endpoints[listenAddr] = endpoints
	if _, ok := f.rrCounter[listenAddr]; !ok {
		var zero uint64
		f.rrCounter[listenAddr] = &zero
	}

	// Check if listener is already active
	if _, ok := f.listeners[listenAddr]; ok {
		f.log.Debug("updated endpoints for listener", zap.String("listenAddr", listenAddr), zap.Int("endpoints", len(endpoints)))
		return nil
	}

	// Start new TCP listener
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", listenAddr, err)
	}

	f.listeners[listenAddr] = ln
	f.log.Info("started bare-metal loadbalancer listener", zap.String("listenAddr", listenAddr), zap.Int("endpoints", len(endpoints)))

	go f.acceptLoop(ctx, listenAddr, ln)
	return nil
}

func (f *Forwarder) acceptLoop(ctx context.Context, listenAddr string, ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				f.log.Debug("accept error", zap.Error(err))
				return
			}
		}

		go f.handleConnection(listenAddr, conn)
	}
}

func (f *Forwarder) handleConnection(listenAddr string, clientConn net.Conn) {
	defer clientConn.Close()

	// Select next healthy endpoint via Round-Robin
	endpoint := f.selectEndpoint(listenAddr)
	if endpoint == nil {
		f.log.Warn("no healthy backend endpoint available for connection", zap.String("listenAddr", listenAddr))
		return
	}

	targetAddr := fmt.Sprintf("%s:%d", endpoint.Address, endpoint.Port)
	backendConn, err := net.DialTimeout("tcp", targetAddr, 3*time.Second)
	if err != nil {
		f.log.Error("failed to connect to backend endpoint", zap.String("target", targetAddr), zap.Error(err))
		return
	}
	defer backendConn.Close()

	// Bidirectional pipe
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = io.Copy(backendConn, clientConn)
	}()

	go func() {
		defer wg.Done()
		_, _ = io.Copy(clientConn, backendConn)
	}()

	wg.Wait()
}

func (f *Forwarder) selectEndpoint(listenAddr string) *Endpoint {
	f.mu.RLock()
	defer f.mu.RUnlock()

	eps := f.endpoints[listenAddr]
	if len(eps) == 0 {
		return nil
	}

	ctr := f.rrCounter[listenAddr]
	idx := atomic.AddUint64(ctr, 1) % uint64(len(eps))
	return &eps[idx]
}

// CloseAll closes all active proxy listeners.
func (f *Forwarder) CloseAll() {
	f.mu.Lock()
	defer f.mu.Unlock()

	for addr, ln := range f.listeners {
		_ = ln.Close()
		delete(f.listeners, addr)
	}
}
