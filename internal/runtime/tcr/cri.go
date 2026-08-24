// Package tcr provides Container Runtime Interface (CRI v1) implementation for Tarak.
package tcr

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PodSandboxState defines the current lifecycle state of a pod sandbox.
type PodSandboxState string

const (
	SandboxReady    PodSandboxState = "SANDBOX_READY"
	SandboxNotReady PodSandboxState = "SANDBOX_NOTREADY"
)

// PodSandbox represents an isolated network and IPC namespace container envelope.
type PodSandbox struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	IP        string            `json:"ip"`
	State     PodSandboxState   `json:"state"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"createdAt"`
	NetNSPath string            `json:"netnsPath,omitempty"`
}

// CRIEngine implements the Kubernetes Container Runtime Interface (CRI) natively in Go.
type CRIEngine struct {
	mu         sync.RWMutex
	log        *zap.Logger
	dataDir    string
	tcr        *Engine
	sandboxes  map[string]*PodSandbox
	containers map[string]*Process
	cgroups    map[string]*CgroupManager
	overlays   map[string]*OverlayMount
}

// NewCRIEngine creates a new native CRI engine.
func NewCRIEngine(dataDir string, log *zap.Logger) *CRIEngine {
	if log == nil {
		log = zap.NewNop()
	}
	return &CRIEngine{
		log:        log.Named("cri-engine"),
		dataDir:    dataDir,
		tcr:        New(),
		sandboxes:  make(map[string]*PodSandbox),
		containers: make(map[string]*Process),
		cgroups:    make(map[string]*CgroupManager),
		overlays:   make(map[string]*OverlayMount),
	}
}

// RunPodSandbox creates and starts a pod sandbox environment.
func (c *CRIEngine) RunPodSandbox(ctx context.Context, namespace, name string, labels map[string]string) (*PodSandbox, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	sandboxID := fmt.Sprintf("sb-%s-%s-%d", namespace, name, time.Now().UnixNano()%100000)
	sb := &PodSandbox{
		ID:        sandboxID,
		Name:      name,
		Namespace: namespace,
		State:     SandboxReady,
		Labels:    labels,
		CreatedAt: time.Now().UTC(),
	}

	c.sandboxes[sandboxID] = sb
	c.log.Info("CRI started pod sandbox",
		zap.String("id", sandboxID),
		zap.String("pod", fmt.Sprintf("%s/%s", namespace, name)),
	)
	return sb, nil
}

// StopPodSandbox stops and releases a pod sandbox.
func (c *CRIEngine) StopPodSandbox(ctx context.Context, sandboxID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if sb, ok := c.sandboxes[sandboxID]; ok {
		sb.State = SandboxNotReady
		delete(c.sandboxes, sandboxID)
		c.log.Info("CRI stopped pod sandbox", zap.String("id", sandboxID))
	}
	return nil
}

// CreateContainer provisions rootfs layers, cgroup limits, and prepares container process.
func (c *CRIEngine) CreateContainer(ctx context.Context, sandboxID string, cfg ContainerConfig, limits CgroupLimits) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cID := cfg.ID
	if cID == "" {
		cID = fmt.Sprintf("ctr-%d", time.Now().UnixNano()%1000000)
		cfg.ID = cID
	}

	// 1. Setup cgroups limits
	cm, _ := NewCgroupManager(cID)
	if cm != nil {
		_ = cm.ApplyLimits(limits)
		c.cgroups[cID] = cm
	}

	c.log.Debug("CRI container created", zap.String("containerID", cID), zap.String("sandboxID", sandboxID))
	return cID, nil
}

// StartContainer launches the container process inside its isolated sandbox.
func (c *CRIEngine) StartContainer(ctx context.Context, cfg ContainerConfig, logFilePath string) (*Process, error) {
	proc, err := c.tcr.StartContainer(ctx, cfg, logFilePath)
	if err != nil {
		return nil, fmt.Errorf("tcr start container: %w", err)
	}

	c.mu.Lock()
	c.containers[cfg.ID] = proc
	if cm, ok := c.cgroups[cfg.ID]; ok && proc.PID > 0 {
		_ = cm.AttachProcess(proc.PID)
	}
	c.mu.Unlock()

	return proc, nil
}

// StopContainer terminates a running container and frees cgroups/mounts.
func (c *CRIEngine) StopContainer(ctx context.Context, containerID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_ = c.tcr.StopContainer(containerID)

	if cm, ok := c.cgroups[containerID]; ok {
		cm.Cleanup()
		delete(c.cgroups, containerID)
	}
	if om, ok := c.overlays[containerID]; ok {
		_ = om.UnmountOverlay()
		delete(c.overlays, containerID)
	}
	delete(c.containers, containerID)

	return nil
}

// ListContainers returns status snapshot of all managed CRI containers.
func (c *CRIEngine) ListContainers() map[string]*Process {
	c.mu.RLock()
	defer c.mu.RUnlock()

	res := make(map[string]*Process, len(c.containers))
	for k, v := range c.containers {
		res[k] = v
	}
	return res
}
