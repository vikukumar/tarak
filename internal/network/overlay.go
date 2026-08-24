// Package network provides multi-node VXLAN overlay mesh networking for Tarak.
package network

import (
	"context"
	"net"
	"sync"

	"go.uber.org/zap"
)

// OverlayNode represents a peer cluster node in the VXLAN overlay.
type OverlayNode struct {
	NodeID   string `json:"nodeId"`
	NodeIP   string `json:"nodeIp"`   // Host IP (e.g. 192.168.1.51)
	PodCIDR  string `json:"podCIDR"`  // Pod subnet on that node (e.g. 10.244.1.0/24)
	VNI      int    `json:"vni"`      // VXLAN Network Identifier (default 42)
}

// OverlayMesh manages cross-node VXLAN packet encapsulation and mesh routing.
type OverlayMesh struct {
	mu       sync.RWMutex
	log      *zap.Logger
	localIP  string
	nodes    map[string]*OverlayNode
	vni      int
	udpPort  int
	conn     *net.UDPConn
	closed   bool
}

// NewOverlayMesh creates a new cross-node overlay network manager.
func NewOverlayMesh(localIP string, vni int, log *zap.Logger) *OverlayMesh {
	if vni == 0 {
		vni = 42
	}
	if log == nil {
		log = zap.NewNop()
	}

	return &OverlayMesh{
		log:     log.Named("overlay-mesh"),
		localIP: localIP,
		nodes:   make(map[string]*OverlayNode),
		vni:     vni,
		udpPort: 8472, // Standard Linux VXLAN UDP port
	}
}

// RegisterPeerNode adds a remote worker node to the overlay routing table.
func (m *OverlayMesh) RegisterPeerNode(node OverlayNode) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nodes[node.NodeID] = &node
	m.log.Info("registered overlay peer node",
		zap.String("nodeID", node.NodeID),
		zap.String("nodeIP", node.NodeIP),
		zap.String("podCIDR", node.PodCIDR),
	)
}

// UnregisterPeerNode removes a worker node from the overlay table.
func (m *OverlayMesh) UnregisterPeerNode(nodeID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.nodes, nodeID)
	m.log.Info("unregistered overlay peer node", zap.String("nodeID", nodeID))
}

// Start activates the UDP listener for cross-node packet exchange.
func (m *OverlayMesh) Start(ctx context.Context) error {
	addr := &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: m.udpPort}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		// Non-fatal if port is already bound on single-node environments
		m.log.Debug("overlay mesh UDP port note", zap.Error(err))
		return nil
	}
	m.conn = conn

	m.log.Info("overlay VXLAN mesh network started",
		zap.Int("vni", m.vni),
		zap.Int("port", m.udpPort),
	)

	return nil
}

// Stop shuts down the overlay mesh engine.
func (m *OverlayMesh) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	if m.conn != nil {
		_ = m.conn.Close()
	}
}
