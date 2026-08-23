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

	if len(endpoints) == 0 {
		if ln, ok := f.listeners[listenAddr]; ok {
			_ = ln.Close()
			delete(f.listeners, listenAddr)
			delete(f.endpoints, listenAddr)
			delete(f.rrCounter, listenAddr)
			f.log.Info("closed loadbalancer listener with 0 endpoints", zap.String("listenAddr", listenAddr))
		}
		return nil
	}

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

// RemoveServiceRoute closes the listener and clears routes for a specific address.
func (f *Forwarder) RemoveServiceRoute(listenAddr string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if ln, ok := f.listeners[listenAddr]; ok {
		_ = ln.Close()
		delete(f.listeners, listenAddr)
		delete(f.endpoints, listenAddr)
		delete(f.rrCounter, listenAddr)
		f.log.Info("removed loadbalancer proxy listener", zap.String("listenAddr", listenAddr))
	}
}

// SyncAllRoutes updates active routes and closes any listeners no longer in desired set.
func (f *Forwarder) SyncAllRoutes(ctx context.Context, desiredRoutes map[string][]Endpoint) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// 1. Close listeners no longer in desired set or having 0 endpoints
	for addr, ln := range f.listeners {
		eps, ok := desiredRoutes[addr]
		if !ok || len(eps) == 0 {
			f.log.Info("closing obsolete loadbalancer proxy listener", zap.String("listenAddr", addr))
			_ = ln.Close()
			delete(f.listeners, addr)
			delete(f.endpoints, addr)
			delete(f.rrCounter, addr)
		}
	}

	// 2. Open or update desired routes with active endpoints
	for addr, eps := range desiredRoutes {
		if len(eps) == 0 {
			continue
		}
		f.endpoints[addr] = eps
		if _, ok := f.rrCounter[addr]; !ok {
			var zero uint64
			f.rrCounter[addr] = &zero
		}

		if _, ok := f.listeners[addr]; !ok {
			ln, err := net.Listen("tcp", addr)
			if err != nil {
				f.log.Warn("failed to start listener for route", zap.String("listenAddr", addr), zap.Error(err))
				continue
			}
			f.listeners[addr] = ln
			f.log.Info("started bare-metal loadbalancer listener", zap.String("listenAddr", addr), zap.Int("endpoints", len(eps)))
			go f.acceptLoop(ctx, addr, ln)
		}
	}
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
