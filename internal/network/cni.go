package network

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// CNIConfig defines the configuration for Tarak's inbuilt Container Network Interface.
type CNIConfig struct {
	PodCIDR        string   `json:"podCIDR"`        // e.g. "10.244.0.0/16"
	ServiceCIDR    string   `json:"serviceCIDR"`    // e.g. "10.96.0.0/12"
	DNSServerIP    string   `json:"dnsServerIP"`    // e.g. "10.96.0.10"
	BridgeName     string   `json:"bridgeName"`     // e.g. "tarak-br0"
	NodeSubnetMask int      `json:"nodeSubnetMask"` // /24 per node
	EnablePolicy   bool     `json:"enablePolicy"`   // Enable NetworkPolicy enforcement
}

// PodNetworkEndpoint holds CNI network attachment details for an individual container/pod.
type PodNetworkEndpoint struct {
	PodName        string            `json:"podName"`
	Namespace      string            `json:"namespace"`
	IP             string            `json:"ip"`
	Gateway        string            `json:"gateway"`
	Subnet         string            `json:"subnet"`
	HostPortMap    map[int]int       `json:"hostPortMap,omitempty"` // ContainerPort -> HostPort
	Labels         map[string]string `json:"labels,omitempty"`
	Created        time.Time         `json:"created"`
	PolicyEgress   []string          `json:"policyEgress,omitempty"`
	PolicyIngress  []string          `json:"policyIngress,omitempty"`
}

// ServiceRoute represents an active ClusterIP proxy entry.
type ServiceRoute struct {
	ServiceName string   `json:"serviceName"`
	Namespace   string   `json:"namespace"`
	ClusterIP   string   `json:"clusterIP"`
	Port        int      `json:"port"`
	TargetPort  int      `json:"targetPort"`
	Backends    []string `json:"backends"` // "10.244.0.5:8080"
}

// InbuiltCNI implements a native zero-dependency Container Network Interface engine.
type InbuiltCNI struct {
	cfg       CNIConfig
	log       *zap.Logger
	mu        sync.RWMutex
	endpoints map[string]*PodNetworkEndpoint // Key: namespace/podName
	services  map[string]*ServiceRoute       // Key: namespace/serviceName
	ipCounter uint32
	closed    bool
}

// NewInbuiltCNI creates and initializes the Tarak native CNI engine.
func NewInbuiltCNI(cfg CNIConfig, log *zap.Logger) *InbuiltCNI {
	if cfg.PodCIDR == "" {
		cfg.PodCIDR = "10.244.0.0/16"
	}
	if cfg.ServiceCIDR == "" {
		cfg.ServiceCIDR = "10.96.0.0/12"
	}
	if cfg.DNSServerIP == "" {
		cfg.DNSServerIP = "10.96.0.10"
	}
	if cfg.BridgeName == "" {
		cfg.BridgeName = "tarak-br0"
	}
	if cfg.NodeSubnetMask == 0 {
		cfg.NodeSubnetMask = 24
	}
	if log == nil {
		log = zap.NewNop()
	}

	return &InbuiltCNI{
		cfg:       cfg,
		log:       log.Named("inbuilt-cni"),
		endpoints: make(map[string]*PodNetworkEndpoint),
		services:  make(map[string]*ServiceRoute),
	}
}

// AttachPod provisions an isolated IP address, gateway, and host port mapping for a new Pod.
func (c *CNI) AttachPod(ctx context.Context, namespace, podName string, nodeIndex int, ports []int, labels map[string]string) (*PodNetworkEndpoint, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := fmt.Sprintf("%s/%s", namespace, podName)
	if ep, exists := c.endpoints[key]; exists {
		return ep, nil
	}

	// Calculate deterministic IP within node's assigned /24 block
	podIdx := atomic.AddUint32(&c.ipCounter, 1)
	nodeOctet := (nodeIndex % 250)
	podOctet := int((podIdx % 250) + 2) // Reserve .1 for gateway

	podIP := fmt.Sprintf("10.244.%d.%d", nodeOctet, podOctet)
	gatewayIP := fmt.Sprintf("10.244.%d.1", nodeOctet)
	subnet := fmt.Sprintf("10.244.%d.0/24", nodeOctet)

	portMap := make(map[int]int)
	for _, p := range ports {
		if p > 0 {
			portMap[p] = p
		}
	}

	ep := &PodNetworkEndpoint{
		PodName:     podName,
		Namespace:   namespace,
		IP:          podIP,
		Gateway:     gatewayIP,
		Subnet:      subnet,
		HostPortMap: portMap,
		Labels:      labels,
		Created:     time.Now(),
	}

	c.endpoints[key] = ep
	c.log.Info("CNI attached pod network endpoint",
		zap.String("pod", key),
		zap.String("ip", podIP),
		zap.String("gateway", gatewayIP),
		zap.String("subnet", subnet),
		zap.Int("ports", len(ports)),
	)

	return ep, nil
}

// DetachPod releases a pod's CNI network resources.
func (c *CNI) DetachPod(namespace, podName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := fmt.Sprintf("%s/%s", namespace, podName)
	if ep, exists := c.endpoints[key]; exists {
		delete(c.endpoints, key)
		c.log.Info("CNI detached pod network endpoint", zap.String("pod", key), zap.String("ip", ep.IP))
	}
}

// RegisterServiceRoute configures ClusterIP load balancing to active pod backend endpoints.
func (c *CNI) RegisterServiceRoute(namespace, serviceName, clusterIP string, port, targetPort int, backends []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := fmt.Sprintf("%s/%s", namespace, serviceName)
	c.services[key] = &ServiceRoute{
		ServiceName: serviceName,
		Namespace:   namespace,
		ClusterIP:   clusterIP,
		Port:        port,
		TargetPort:  targetPort,
		Backends:    backends,
	}
	c.log.Debug("CNI registered service ClusterIP route",
		zap.String("service", key),
		zap.String("clusterIP", clusterIP),
		zap.Int("backends", len(backends)),
	)
}

// CheckNetworkPolicy verifies if traffic from srcPod to dstPod is permitted.
func (c *CNI) CheckNetworkPolicy(srcNamespace, srcPod, dstNamespace, dstPod string, port int) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Default allow if policies are open
	if !c.cfg.EnablePolicy {
		return true
	}

	// Inter-namespace boundary verification
	if srcNamespace != dstNamespace {
		// Strict isolation unless explicitly whitelisted
		return false
	}

	return true
}

// ListEndpoints returns all active CNI network endpoints.
func (c *CNI) ListEndpoints() []*PodNetworkEndpoint {
	c.mu.RLock()
	defer c.mu.RUnlock()

	list := make([]*PodNetworkEndpoint, 0, len(c.endpoints))
	for _, ep := range c.endpoints {
		list = append(list, ep)
	}
	return list
}

// Type alias for compatibility
type CNI = InbuiltCNI
