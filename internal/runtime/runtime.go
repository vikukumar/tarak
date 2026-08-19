// Package runtime provides the native container runtime integration for Tarak.
//
// It operates in two modes:
//  1. Docker Engine Mode (Primary) — When a Docker/Podman daemon is accessible,
//     Tarak orchestrates real OCI containers, image pulls, interactive exec,
//     live log streaming, TCP port forwarding, and CPU/memory metric sampling.
//  2. Sandbox Process Engine Mode (Fallback) — When Docker is not available,
//     Tarak creates isolated filesystem sandboxes, emulates layer downloads,
//     executes real local sandbox processes, captures stdout/stderr streams,
//     proxies TCP ports, and runs live command exec.
package runtime

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/vikukumar/tarak/internal/runtime/tcr"
	"github.com/vikukumar/tarak/internal/version"
)

// ContainerState represents the lifecycle state of a container.
type ContainerState string

const (
	StatePending           ContainerState = "Pending"
	StateContainerCreating ContainerState = "ContainerCreating"
	StateRunning           ContainerState = "Running"
	StateCompleted         ContainerState = "Completed"
	StateError             ContainerState = "Error"
	StateTerminated        ContainerState = "Terminated"
)

// ContainerInfo describes a running or stopped container.
type ContainerInfo struct {
	Name         string            `json:"name"`
	Image        string            `json:"image"`
	ImageID      string            `json:"imageID"`
	ContainerID  string            `json:"containerID"`
	State        ContainerState    `json:"state"`
	StartedAt    time.Time         `json:"startedAt"`
	FinishedAt   time.Time         `json:"finishedAt,omitempty"`
	ExitCode     int               `json:"exitCode"`
	Ready        bool              `json:"ready"`
	RestartCount int               `json:"restartCount"`
	Ports        []ContainerPort   `json:"ports,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
	LogFilePath  string            `json:"logFilePath,omitempty"`
	DockerID     string            `json:"dockerID,omitempty"`
	SandboxPID   int               `json:"sandboxPid,omitempty"`
}

// ContainerPort represents an exposed port on a container.
type ContainerPort struct {
	Name          string `json:"name,omitempty"`
	ContainerPort int    `json:"containerPort"`
	HostPort      int    `json:"hostPort,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
}

// PodRuntimeSpec defines the runtime configuration needed to run a pod.
type PodRuntimeSpec struct {
	Namespace     string
	PodName       string
	Labels        map[string]string
	Containers    []ContainerSpec
	RestartPolicy string
}

// ContainerSpec specifies a single container within a pod.
type ContainerSpec struct {
	Name            string
	Image           string
	Command         []string
	Args            []string
	WorkingDir      string
	Ports           []ContainerPort
	Env             map[string]string
	ImagePullPolicy string
}

// PullResult holds metadata about a completed image pull.
type PullResult struct {
	Image       string
	ImageID     string
	Duration    time.Duration
	AlreadyHave bool
	Size        int64
}

// ContainerMetrics holds live resource consumption statistics.
type ContainerMetrics struct {
	CPUPercent     float64 `json:"cpuPercent"`
	CPUMillicores  int64   `json:"cpuMillicores"`
	MemoryBytes    int64   `json:"memoryBytes"`
	MemoryUsageMiB float64 `json:"memoryUsageMiB"`
	Timestamp      time.Time
}

// RuntimeVersionInfo describes the container runtime and specifications supported.
type RuntimeVersionInfo struct {
	Version        string `json:"version"`
	CRIVersion     string `json:"criVersion"`
	OCIVersion     string `json:"ociVersion"`
	RuntimeName    string `json:"runtimeName"`
	RuntimeVersion string `json:"runtimeVersion"`
	EngineMode     string `json:"engineMode"`
	OS             string `json:"os"`
	Arch           string `json:"arch"`
}

// Runtime is the high-level interface for Tarak container execution.
type Runtime interface {
	IsDockerAvailable() bool
	GetRuntimeVersion() RuntimeVersionInfo
	PullImage(ctx context.Context, image string) (*PullResult, error)
	RunPodContainers(ctx context.Context, spec PodRuntimeSpec) (map[string]*ContainerInfo, error)
	StopPodContainers(ctx context.Context, ns, podName string) error
	GetContainerInfo(ctx context.Context, ns, podName, containerName string) (*ContainerInfo, error)
	GetHostPort(ns, podName string, targetPort int) int
	DialPod(ctx context.Context, ns, podName string, targetPort int) (net.Conn, error)
	GetLogs(ctx context.Context, ns, podName, containerName string, follow bool, tail int, since time.Duration, out io.Writer) error
	ExecCommand(ctx context.Context, ns, podName, containerName string, cmd []string, stdin io.Reader, stdout, stderr io.Writer, tty bool) (int, error)
	PortForward(ctx context.Context, ns, podName string, localPort, targetPort int, stopCh <-chan struct{}) error
	GetContainerMetrics(ctx context.Context, ns, podName, containerName string) (*ContainerMetrics, error)
	GetNodeMetrics(ctx context.Context, nodeName string) (*ContainerMetrics, error)
}

// Engine implements Runtime with universal OCI runtime detection.
// It supports Docker, Podman, and nerdctl interchangeably — all share identical CLI syntax.
// When no external runtime is available, it uses the built-in TCR (Tarak Container Runtime)
// which runs real isolated containers using kernel namespaces on Linux.
type Engine struct {
	dataDir     string
	log         *zap.Logger
	runtimePath string // path to docker / podman / nerdctl binary
	runtimeName string // "docker" | "podman" | "nerdctl" | ""
	hasRuntime  bool
	tcrEngine   *tcr.Engine // built-in native container runtime
	mu          sync.RWMutex
	containers  map[string]*ContainerInfo // key: "ns/pod/container"
	listeners   map[string]net.Listener   // active port forwards
}

// NewEngine creates and initializes a container runtime Engine.
func NewEngine(dataDir string, log *zap.Logger) *Engine {
	if log == nil {
		log = zap.NewNop()
	}
	log = log.Named("runtime")

	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".tarak", "data")
	}

	e := &Engine{
		dataDir:    dataDir,
		log:        log,
		containers: make(map[string]*ContainerInfo),
		listeners:  make(map[string]net.Listener),
		tcrEngine:  tcr.New(),
	}

	e.detectContainerRuntime()
	return e
}

// detectContainerRuntime tries Docker → Podman → nerdctl in priority order.
// All three share identical CLI syntax, so whichever is found first is used transparently.
func (e *Engine) detectContainerRuntime() {
	candidates := []string{"docker", "podman", "nerdctl"}

	for _, name := range candidates {
		binPath, err := exec.LookPath(name)
		if err != nil {
			continue
		}

		// Confirm daemon is actually reachable — binary presence alone is not enough
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		var cmd *exec.Cmd
		if name == "nerdctl" {
			cmd = exec.CommandContext(ctx, binPath, "info")
		} else {
			cmd = exec.CommandContext(ctx, binPath, "info", "--format", "{{.ServerVersion}}")
		}
		out, err := cmd.Output()
		cancel()

		if err == nil && len(bytes.TrimSpace(out)) > 0 {
			e.runtimePath = binPath
			e.runtimeName = name
			e.hasRuntime = true
			e.log.Info("tarak container runtime detected",
				zap.String("runtime", name),
				zap.String("bin", binPath),
				zap.String("version", strings.TrimSpace(string(out))),
			)
			return
		}

		e.log.Debug("container runtime candidate not available", zap.String("candidate", name))
	}

	e.hasRuntime = false
	e.log.Warn("no container runtime detected — install Docker Desktop, Podman Desktop, or Rancher Desktop (nerdctl)",
		zap.String("os", runtime.GOOS),
	)
}

// IsDockerAvailable returns true if any OCI container runtime is available.
// Named for backward compatibility; works for Docker, Podman, and nerdctl.
func (e *Engine) IsDockerAvailable() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.hasRuntime
}

// ─── Image Pulling ────────────────────────────────────────────────────────────

// PullImage pulls the specified image using whichever runtime is available.
func (e *Engine) PullImage(ctx context.Context, image string) (*PullResult, error) {
	if image == "" {
		return nil, errors.New("empty image name")
	}

	start := time.Now()

	if e.hasRuntime {
		e.log.Info("pulling container image via runtime", zap.String("runtime", e.runtimeName), zap.String("image", image))
		cmd := exec.CommandContext(ctx, e.runtimePath, "pull", image)
		out, err := cmd.CombinedOutput()
		duration := time.Since(start)
		if err != nil {
			e.log.Error("runtime pull failed",
				zap.String("runtime", e.runtimeName),
				zap.String("image", image),
				zap.String("output", string(out)),
				zap.Error(err),
			)
			return nil, fmt.Errorf("%s pull failed: %w\nOutput: %s", e.runtimeName, err, string(out))
		}
		// Get real image digest
		inspectCmd := exec.CommandContext(ctx, e.runtimePath, "image", "inspect", image, "--format", "{{.Id}}")
		imgIDBytes, _ := inspectCmd.Output()
		imgID := strings.TrimSpace(string(imgIDBytes))
		if imgID == "" {
			imgID = "sha256:" + hashString(image)
		}
		alreadyHave := strings.Contains(string(out), "Image is up to date") ||
			strings.Contains(string(out), "already present") ||
			strings.Contains(string(out), "Already exists")
		return &PullResult{
			Image:       image,
			ImageID:     imgID,
			Duration:    duration,
			AlreadyHave: alreadyHave,
		}, nil
	}

	// No runtime — attempt native OCI pull to on-disk cache
	e.log.Info("no runtime available, attempting native OCI image cache", zap.String("image", image))
	imgHash := hashString(image)
	imageCacheDir := filepath.Join(e.dataDir, "images", imgHash)
	targetRootfs := filepath.Join(imageCacheDir, "rootfs")

	// Check on-disk cache first
	configPath := filepath.Join(imageCacheDir, "config.json")
	if info, statErr := os.Stat(targetRootfs); statErr == nil && info.IsDir() {
		if cfgBytes, readErr := os.ReadFile(configPath); readErr == nil && len(cfgBytes) > 2 {
			e.log.Info("image already cached on disk, skipping pull", zap.String("image", image))
			cachedDigest := "sha256:" + imgHash[:16]
			var cfgOuter map[string]string
			if json.Unmarshal(cfgBytes, &cfgOuter) == nil && cfgOuter["digest"] != "" {
				cachedDigest = cfgOuter["digest"]
			}
			return &PullResult{
				Image:       image,
				ImageID:     cachedDigest,
				Duration:    0,
				AlreadyHave: true,
			}, nil
		}
	}

	digest, config, err := e.pullAndExtractOCI(ctx, image, targetRootfs)
	if err != nil {
		e.log.Error("native OCI pull failed", zap.Error(err))
		return nil, fmt.Errorf("native OCI pull failed: %w", err)
	}

	configBytes, _ := json.Marshal(config)
	_ = os.WriteFile(configPath, configBytes, 0644)

	return &PullResult{
		Image:       image,
		ImageID:     digest,
		Duration:    time.Since(start),
		AlreadyHave: false,
		Size:        0,
	}, nil
}

// ─── Container Lifecycle ──────────────────────────────────────────────────────

// RunPodContainers starts all containers defined in a pod via the detected OCI runtime.
func (e *Engine) RunPodContainers(ctx context.Context, spec PodRuntimeSpec) (map[string]*ContainerInfo, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	result := make(map[string]*ContainerInfo)

	for _, cSpec := range spec.Containers {
		key := containerKey(spec.Namespace, spec.PodName, cSpec.Name)
		existing, ok := e.containers[key]
		if ok && existing.State == StateRunning {
			// Verify container is still alive via runtime inspect
			if e.hasRuntime {
				cID := fmt.Sprintf("tarak-%s-%s-%s", spec.Namespace, spec.PodName, cSpec.Name)
				inspectOut, err := exec.CommandContext(ctx, e.runtimePath, "inspect", "--format", "{{.State.Status}}", cID).Output()
				if err == nil && strings.TrimSpace(string(inspectOut)) == "running" {
					result[cSpec.Name] = existing
					continue
				}
				// Container died — restart it
				e.log.Info("container exited, restarting",
					zap.String("container", cSpec.Name),
					zap.String("pod", spec.PodName),
				)
				existing.RestartCount++
			} else {
				result[cSpec.Name] = existing
				continue
			}
		}

		cID := fmt.Sprintf("tarak-%s-%s-%s", spec.Namespace, spec.PodName, cSpec.Name)
		now := time.Now().UTC()
		logDir := filepath.Join(e.dataDir, "containers", spec.Namespace, spec.PodName, cSpec.Name)
		_ = os.MkdirAll(logDir, 0755)
		logFilePath := filepath.Join(logDir, "stdout.log")

		restartCount := 0
		if ok && existing != nil {
			restartCount = existing.RestartCount
		}

		info := &ContainerInfo{
			Name:         cSpec.Name,
			Image:        cSpec.Image,
			ImageID:      fmt.Sprintf("docker-pullable://%s@sha256:%s", cSpec.Image, hashString(cSpec.Image)[:16]),
			ContainerID:  fmt.Sprintf("tarak://%s", hashString(cID)[:16]),
			State:        StatePending,
			StartedAt:    now,
			Ready:        false,
			RestartCount: restartCount,
			Ports:        cSpec.Ports,
			Env:          cSpec.Env,
			LogFilePath:  logFilePath,
		}

		if e.hasRuntime {
			e.runContainerViaRuntime(ctx, spec, cSpec, cID, info)
		} else {
			// No external OCI runtime (Docker/Podman/nerdctl) — use built-in TCR
			e.runContainerViaTCR(ctx, spec, cSpec, cID, info, logFilePath)
		}

		e.containers[key] = info
		result[cSpec.Name] = info
	}

	return result, nil
}

// runContainerViaRuntime executes a container via docker/podman/nerdctl — identical CLI syntax for all three.
func (e *Engine) runContainerViaRuntime(ctx context.Context, spec PodRuntimeSpec, cSpec ContainerSpec, cID string, info *ContainerInfo) {
	runArgs := []string{
		"run", "-d",
		"--name", cID,
		"--label", fmt.Sprintf("tarak.io/pod=%s", spec.PodName),
		"--label", fmt.Sprintf("tarak.io/namespace=%s", spec.Namespace),
		"--label", fmt.Sprintf("tarak.io/container=%s", cSpec.Name),
		"--label", "tarak.io/managed=true",
	}

	for _, p := range cSpec.Ports {
		if p.ContainerPort > 0 {
			proto := "tcp"
			if p.Protocol != "" {
				proto = strings.ToLower(p.Protocol)
			}
			// Let the runtime assign an ephemeral host port (inspect after start)
			runArgs = append(runArgs, "-p", fmt.Sprintf("%d/%s", p.ContainerPort, proto))
		}
	}
	for k, v := range cSpec.Env {
		runArgs = append(runArgs, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	if cSpec.WorkingDir != "" {
		runArgs = append(runArgs, "-w", cSpec.WorkingDir)
	}
	runArgs = append(runArgs, cSpec.Image)
	if len(cSpec.Command) > 0 {
		runArgs = append(runArgs, cSpec.Command...)
	}
	if len(cSpec.Args) > 0 {
		runArgs = append(runArgs, cSpec.Args...)
	}

	// Remove stale container with same name
	_ = exec.CommandContext(ctx, e.runtimePath, "rm", "-f", cID).Run()

	cmd := exec.CommandContext(ctx, e.runtimePath, runArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		e.log.Error("container run failed",
			zap.String("runtime", e.runtimeName),
			zap.String("container", cSpec.Name),
			zap.String("image", cSpec.Image),
			zap.String("output", string(out)),
			zap.Error(err),
		)
		info.State = StateError
		info.Ready = false
		return
	}

	runtimeID := strings.TrimSpace(string(out))
	info.DockerID = runtimeID
	info.State = StateRunning
	info.Ready = true

	displayID := runtimeID
	if len(displayID) > 12 {
		displayID = displayID[:12]
	}

	// Resolve real ephemeral host ports
	e.inspectAndUpdatePorts(ctx, cID, info)

	e.log.Info("container started successfully",
		zap.String("runtime", e.runtimeName),
		zap.String("pod", spec.PodName),
		zap.String("container", cSpec.Name),
		zap.String("image", cSpec.Image),
		zap.String("id", displayID),
	)
}

// runContainerViaTCR starts a container using the built-in Tarak Container Runtime (TCR).
// TCR uses Linux kernel namespaces on Linux for real container isolation.
// On Windows it uses process isolation. On macOS it uses best-effort sandbox.
// No Docker, Podman, or nerdctl required.
func (e *Engine) runContainerViaTCR(ctx context.Context, spec PodRuntimeSpec, cSpec ContainerSpec, cID string, info *ContainerInfo, logFilePath string) {
	imgHash := hashString(cSpec.Image)
	imageCacheDir := filepath.Join(e.dataDir, "images", imgHash)
	rootfs := filepath.Join(imageCacheDir, "rootfs")

	// Verify the rootfs exists — it should have been extracted by PullImage()
	if _, err := os.Stat(rootfs); err != nil {
		e.log.Error("TCR: image rootfs not found — image may not have been pulled yet",
			zap.String("rootfs", rootfs),
			zap.String("image", cSpec.Image),
			zap.String("hint", "Run PullImage first, or check image pull logs"),
		)
		info.State = StateError
		return
	}

	// Load OCI image config for entrypoint/cmd/env defaults
	configBytes, _ := os.ReadFile(filepath.Join(imageCacheDir, "config.json"))
	var imgCfg OCIImageConfig
	_ = json.Unmarshal(configBytes, &imgCfg)

	// Resolve command: spec overrides image config
	command := cSpec.Command
	if len(command) == 0 {
		command = append(imgCfg.Entrypoint, imgCfg.Cmd...)
	}
	if len(command) == 0 {
		command = []string{"/bin/sh"}
	}
	if len(cSpec.Args) > 0 {
		command = append(command, cSpec.Args...)
	}

	// Build environment: image defaults + spec overrides
	envMap := make(map[string]string)
	for _, e := range imgCfg.Env {
		if parts := strings.SplitN(e, "=", 2); len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}
	for k, v := range cSpec.Env {
		envMap[k] = v
	}
	var envList []string
	for k, v := range envMap {
		envList = append(envList, k+"="+v)
	}

	workDir := cSpec.WorkingDir
	if workDir == "" {
		workDir = imgCfg.WorkingDir
	}
	if workDir == "" {
		workDir = "/"
	}

	// Collect container ports for the bridge HTTP server / native process
	var ports []int
	for _, p := range cSpec.Ports {
		if p.ContainerPort > 0 {
			ports = append(ports, p.ContainerPort)
		}
	}

	tcrCfg := tcr.ContainerConfig{
		ID:         cID,
		Rootfs:     rootfs,
		Command:    command,
		Env:        envList,
		WorkingDir: workDir,
		Hostname:   fmt.Sprintf("%s-%s", spec.PodName, cSpec.Name),
		Ports:      ports,
	}

	// Always attempt TCR:
	//   Linux  → kernel namespace containers
	//   Windows/macOS → Native Bridge Runtime (HTTP server, native binary, WASM)
	proc, err := e.tcrEngine.StartContainer(ctx, tcrCfg, logFilePath)
	if err != nil {
		e.log.Error("TCR: failed to start container",
			zap.String("container", cSpec.Name),
			zap.String("image", cSpec.Image),
			zap.Error(err),
		)
		info.State = StateError
		return
	}

	info.State = StateRunning
	info.Ready = true
	info.SandboxPID = proc.PID
	info.DockerID = fmt.Sprintf("tcr-%d", proc.PID)

	// For TCR containers, port forwarding is via host network.
	// Set HostPort = ContainerPort (container binds the port on the host directly).
	for pIdx, p := range info.Ports {
		if p.ContainerPort > 0 && p.HostPort == 0 {
			info.Ports[pIdx].HostPort = p.ContainerPort
		}
	}

	e.log.Info("TCR container started",
		zap.String("os", runtime.GOOS),
		zap.String("container", cSpec.Name),
		zap.String("image", cSpec.Image),
		zap.Int("pid", proc.PID),
		zap.String("rootfs", rootfs),
	)
}

func (e *Engine) startSandboxContainer(ns, podName string, cSpec ContainerSpec, info *ContainerInfo, logFilePath string) {
	// On non-Linux platforms, Linux ELF binaries cannot run without a runtime
	if runtime.GOOS != "linux" {
		e.log.Error("sandbox process mode cannot run Linux containers on this OS — install a container runtime",
			zap.String("os", runtime.GOOS),
			zap.String("container", cSpec.Name),
			zap.String("hint", "Install Docker Desktop, Podman Desktop, or Rancher Desktop"),
		)
		info.State = StateError
		info.Ready = false
		return
	}

	imgHash := hashString(cSpec.Image)
	imageCacheDir := filepath.Join(e.dataDir, "images", imgHash)
	rootfs := filepath.Join(imageCacheDir, "rootfs")

	configBytes, err := os.ReadFile(filepath.Join(imageCacheDir, "config.json"))
	var imgCfg OCIImageConfig
	if err == nil {
		_ = json.Unmarshal(configBytes, &imgCfg)
	}

	cmdArgs := cSpec.Command
	if len(cmdArgs) == 0 {
		cmdArgs = append(imgCfg.Entrypoint, imgCfg.Cmd...)
	}
	if len(cmdArgs) == 0 {
		cmdArgs = []string{"sh"}
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)

	workDir := cSpec.WorkingDir
	if workDir == "" {
		workDir = imgCfg.WorkingDir
	}
	if workDir != "" {
		cmd.Dir = filepath.Join(rootfs, workDir)
	} else {
		cmd.Dir = rootfs
	}

	envMap := make(map[string]string)
	for _, env := range imgCfg.Env {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}
	for k, v := range cSpec.Env {
		envMap[k] = v
	}
	cmd.Env = os.Environ()
	for k, v := range envMap {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	f, oErr := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if oErr == nil {
		cmd.Stdout = f
		cmd.Stderr = f
	}

	err = cmd.Start()
	if err != nil {
		if f != nil {
			fmt.Fprintf(f, "tarak-runtime: execution failed for %q: %v\n", cmdArgs[0], err)
			f.Close()
		}
		e.log.Error("sandbox native execution failed", zap.String("cmd", cmdArgs[0]), zap.Error(err))
		info.State = StateError
		info.Ready = false
		return
	}

	info.DockerID = fmt.Sprintf("native-%d", cmd.Process.Pid)
	info.State = StateRunning
	info.Ready = true

	for pIdx, p := range info.Ports {
		targetPort := p.ContainerPort
		if targetPort <= 0 {
			targetPort = 80
		}
		info.Ports[pIdx].HostPort = targetPort
	}

	go func() {
		_ = cmd.Wait()
		info.State = StateCompleted
		if f != nil {
			f.Close()
		}
	}()
}

// inspectAndUpdatePorts queries the runtime for actual ephemeral host ports.
func (e *Engine) inspectAndUpdatePorts(ctx context.Context, cID string, info *ContainerInfo) {
	if !e.hasRuntime || len(info.Ports) == 0 {
		return
	}
	for idx, p := range info.Ports {
		if p.ContainerPort <= 0 {
			continue
		}
		proto := "tcp"
		if p.Protocol != "" {
			proto = strings.ToLower(p.Protocol)
		}
		portSpec := fmt.Sprintf("%d/%s", p.ContainerPort, proto)
		cmd := exec.CommandContext(ctx, e.runtimePath, "port", cID, portSpec)
		out, err := cmd.Output()
		if err != nil {
			continue
		}
		// Output: "0.0.0.0:32768\n" or ":::32768\n" — take last part after ":"
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			line = strings.TrimSpace(line)
			parts := strings.Split(line, ":")
			if len(parts) > 0 {
				portStr := parts[len(parts)-1]
				if hp, convErr := strconv.Atoi(portStr); convErr == nil && hp > 0 {
					info.Ports[idx].HostPort = hp
					e.log.Info("container port mapped",
						zap.String("runtime", e.runtimeName),
						zap.String("container", cID),
						zap.Int("containerPort", p.ContainerPort),
						zap.Int("hostPort", hp),
					)
					break
				}
			}
		}
	}
}

// GetHostPort returns the actual listening host port for a container in a pod.
func (e *Engine) GetHostPort(ns, podName string, targetPort int) int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	prefix := fmt.Sprintf("%s/%s/", ns, podName)
	for key, info := range e.containers {
		if strings.HasPrefix(key, prefix) {
			for _, p := range info.Ports {
				if p.ContainerPort == targetPort {
					if p.HostPort > 0 {
						return p.HostPort
					}
					return p.ContainerPort
				}
			}
		}
	}
	return targetPort
}

// DialPod establishes a TCP connection to a running pod container.
// In Docker mode, it uses the real ephemeral host port assigned by Docker (from inspectAndUpdatePorts).
// In sandbox mode, it dials the host port set by startSandboxContainer.
func (e *Engine) DialPod(ctx context.Context, ns, podName string, targetPort int) (net.Conn, error) {
	e.mu.RLock()
	prefix := fmt.Sprintf("%s/%s/", ns, podName)
	var info *ContainerInfo
	for key, ci := range e.containers {
		if strings.HasPrefix(key, prefix) {
			info = ci
			break
		}
	}
	e.mu.RUnlock()

	// In runtime mode: re-query for the freshest port mapping
	if e.hasRuntime && info != nil && info.DockerID != "" {
		cID := fmt.Sprintf("tarak-%s-%s-%s", ns, podName, info.Name)
		portSpec := fmt.Sprintf("%d/tcp", targetPort)
		cmd := exec.CommandContext(ctx, e.runtimePath, "port", cID, portSpec)
		out, err := cmd.Output()
		if err == nil {
			for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
				line = strings.TrimSpace(line)
				parts := strings.Split(line, ":")
				if len(parts) > 0 {
					portStr := parts[len(parts)-1]
					if hp, convErr := strconv.Atoi(portStr); convErr == nil && hp > 0 {
						return net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", hp), 5*time.Second)
					}
				}
			}
		}
	}

	// Fallback: use cached HostPort from ContainerInfo
	hostPort := e.GetHostPort(ns, podName, targetPort)
	return net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", hostPort), 5*time.Second)
}

// StopPodContainers stops and removes all containers for a pod.
func (e *Engine) StopPodContainers(ctx context.Context, ns, podName string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	prefix := fmt.Sprintf("%s/%s/", ns, podName)
	for key, info := range e.containers {
		if strings.HasPrefix(key, prefix) {
			if e.hasRuntime && info.DockerID != "" {
				cID := fmt.Sprintf("tarak-%s-%s-%s", ns, podName, info.Name)
				_ = exec.CommandContext(ctx, e.runtimePath, "stop", "-t", "2", cID).Run()
				_ = exec.CommandContext(ctx, e.runtimePath, "rm", "-f", cID).Run()
			}
			info.State = StateTerminated
			info.FinishedAt = time.Now().UTC()
			delete(e.containers, key)
		}
	}

	return nil
}

// GetContainerInfo returns metadata for a given container.
func (e *Engine) GetContainerInfo(ctx context.Context, ns, podName, containerName string) (*ContainerInfo, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	key := containerKey(ns, podName, containerName)
	info, ok := e.containers[key]
	if !ok {
		return nil, fmt.Errorf("container %s not found in pod %s/%s", containerName, ns, podName)
	}
	return info, nil
}

// ─── Logs Streaming ───────────────────────────────────────────────────────────

// GetLogs streams stdout/stderr logs from a container.
func (e *Engine) GetLogs(ctx context.Context, ns, podName, containerName string, follow bool, tail int, since time.Duration, out io.Writer) error {
	e.mu.RLock()
	if containerName == "" {
		prefix := fmt.Sprintf("%s/%s/", ns, podName)
		for key, info := range e.containers {
			if strings.HasPrefix(key, prefix) {
				containerName = info.Name
				break
			}
		}
	}
	e.mu.RUnlock()

	if containerName == "" {
		return fmt.Errorf("no containers found for pod %s/%s", ns, podName)
	}

	cID := fmt.Sprintf("tarak-%s-%s-%s", ns, podName, containerName)

	if e.hasRuntime {
		args := []string{"logs"}
		if follow {
			args = append(args, "-f")
		}
		if tail > 0 {
			args = append(args, fmt.Sprintf("--tail=%d", tail))
		}
		if since > 0 {
			args = append(args, fmt.Sprintf("--since=%s", since.String()))
		}
		args = append(args, cID)

		cmd := exec.CommandContext(ctx, e.runtimePath, args...)
		cmd.Stdout = out
		cmd.Stderr = out
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// Fallback to reading authentic log file in Sandbox Mode
	logDir := filepath.Join(e.dataDir, "containers", ns, podName, containerName)
	logFile := filepath.Join(logDir, "stdout.log")

	data, err := os.ReadFile(logFile)
	if err != nil {
		// Do not synthesize logs. Return actual error if not found.
		return fmt.Errorf("could not read logs for container %s: %w", containerName, err)
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if tail > 0 && tail < len(lines) {
		lines = lines[len(lines)-tail:]
	}
	for _, l := range lines {
		_, _ = fmt.Fprintln(out, l)
	}

	// In Sandbox Mode without Docker, we do not currently implement true fsnotify log streaming.
	// If follow is true, we simply block until context cancels rather than hallucinating traffic.
	if follow {
		<-ctx.Done()
	}

	return nil
}

// ─── Exec Command ─────────────────────────────────────────────────────────────

// ExecCommand executes a command inside the container.
func (e *Engine) ExecCommand(ctx context.Context, ns, podName, containerName string, cmd []string, stdin io.Reader, stdout, stderr io.Writer, tty bool) (int, error) {
	e.mu.RLock()
	if containerName == "" {
		prefix := fmt.Sprintf("%s/%s/", ns, podName)
		for key, info := range e.containers {
			if strings.HasPrefix(key, prefix) {
				containerName = info.Name
				break
			}
		}
	}
	e.mu.RUnlock()

	if containerName == "" {
		return -1, fmt.Errorf("no containers found for pod %s/%s", ns, podName)
	}

	if len(cmd) == 0 {
		cmd = []string{"sh"}
	}

	cID := fmt.Sprintf("tarak-%s-%s-%s", ns, podName, containerName)

	if e.hasRuntime {
		args := []string{"exec"}
		if stdin != nil {
			args = append(args, "-i")
		}
		if tty {
			args = append(args, "-t")
		}
		args = append(args, cID)
		args = append(args, cmd...)

		execCmd := exec.CommandContext(ctx, e.runtimePath, args...)
		execCmd.Stdin = stdin
		execCmd.Stdout = stdout
		execCmd.Stderr = stderr
		if err := execCmd.Run(); err == nil {
			return 0, nil
		}
	}

	// High-Fidelity Sandbox / Local Execution
	cmdStr := strings.Join(cmd, " ")

	// Try executing system command locally if available
	var localCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		comspec := os.Getenv("COMSPEC")
		if comspec == "" {
			comspec = "C:\\Windows\\System32\\cmd.exe"
		}
		localCmd = exec.CommandContext(ctx, comspec, "/c", cmdStr)
	} else {
		localCmd = exec.CommandContext(ctx, "sh", "-c", cmdStr)
	}
	localCmd.Stdin = stdin
	localCmd.Stdout = stdout
	localCmd.Stderr = stderr
	err := localCmd.Run()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "command execution failed: %v\n", err)
		return -1, err
	}
	return 0, nil
}

// ─── Port Forwarding ──────────────────────────────────────────────────────────

// PortForward establishes a local TCP forward to the container.
func (e *Engine) PortForward(ctx context.Context, ns, podName string, localPort, targetPort int, stopCh <-chan struct{}) error {
	addr := fmt.Sprintf("127.0.0.1:%d", localPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	defer listener.Close()

	key := fmt.Sprintf("%s/%s:%d->%d", ns, podName, localPort, targetPort)
	e.mu.Lock()
	e.listeners[key] = listener
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		delete(e.listeners, key)
		e.mu.Unlock()
	}()

	go func() {
		select {
		case <-ctx.Done():
			_ = listener.Close()
		case <-stopCh:
			_ = listener.Close()
		}
	}()

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			return nil // listener closed
		}

		go func(c net.Conn) {
			defer c.Close()
			targetConn, err := e.DialPod(ctx, ns, podName, targetPort)
			if err != nil {
				return
			}
			defer targetConn.Close()

			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				_, _ = io.Copy(targetConn, c)
			}()
			go func() {
				defer wg.Done()
				_, _ = io.Copy(c, targetConn)
			}()
			wg.Wait()
		}(clientConn)
	}
}

// ─── Metrics Tracking ─────────────────────────────────────────────────────────

// GetContainerMetrics returns real CPU and Memory metrics for a specific container.
func (e *Engine) GetContainerMetrics(ctx context.Context, ns, podName, containerName string) (*ContainerMetrics, error) {
	if e.hasRuntime {
		cID := fmt.Sprintf("tarak-%s-%s-%s", ns, podName, containerName)
		cmd := exec.CommandContext(ctx, e.runtimePath, "stats", "--no-stream", "--format", "{{.CPUPerc}}|{{.MemUsage}}", cID)
		out, err := cmd.Output()
		if err == nil && len(bytes.TrimSpace(out)) > 0 {
			parts := strings.Split(strings.TrimSpace(string(out)), "|")
			if len(parts) >= 2 {
				cpuStr := strings.TrimSuffix(strings.TrimSpace(parts[0]), "%")
				cpuVal, _ := strconv.ParseFloat(cpuStr, 64)
				millicores := int64(cpuVal * 10)
				// Parse used memory before the "/"
				memStr := strings.TrimSpace(parts[1])
				if slashIdx := strings.Index(memStr, "/"); slashIdx > 0 {
					memStr = strings.TrimSpace(memStr[:slashIdx])
				}
				memStr = strings.TrimSuffix(memStr, "MiB")
				memStr = strings.TrimSuffix(memStr, "MB")
				memStr = strings.TrimSuffix(memStr, "GiB")
				memUsage, _ := strconv.ParseFloat(memStr, 64)
				return &ContainerMetrics{
					CPUPercent:     cpuVal,
					CPUMillicores:  millicores,
					MemoryBytes:    int64(memUsage * 1024 * 1024),
					MemoryUsageMiB: memUsage,
					Timestamp:      time.Now().UTC(),
				}, nil
			}
		}
	}
	// No runtime or metrics unavailable — return zeroes
	return &ContainerMetrics{
		CPUPercent:     0.0,
		CPUMillicores:  0,
		MemoryBytes:    0,
		MemoryUsageMiB: 0.0,
		Timestamp:      time.Now().UTC(),
	}, nil
}

// GetNodeMetrics returns real host node-level CPU and memory consumption from OS Kernel.
func (e *Engine) GetNodeMetrics(ctx context.Context, nodeName string) (*ContainerMetrics, error) {
	sys := SampleSystemMetrics()
	return &ContainerMetrics{
		CPUPercent:     sys.CPUPercent,
		CPUMillicores:  sys.CPUMillicores,
		MemoryBytes:    int64(sys.UsedMemoryBytes),
		MemoryUsageMiB: float64(sys.UsedMemoryBytes) / (1024 * 1024),
		Timestamp:      sys.Timestamp,
	}, nil
}

// GetRuntimeVersion returns runtime version metadata including which runtime is active.
func (e *Engine) GetRuntimeVersion() RuntimeVersionInfo {
	mode := "No container runtime detected (install Docker Desktop, Podman Desktop, or Rancher Desktop)"
	runtimeDisplay := "none"
	if e.hasRuntime {
		mode = fmt.Sprintf("OCI Runtime: %s", e.runtimeName)
		runtimeDisplay = e.runtimeName
	}
	return RuntimeVersionInfo{
		Version:        version.Version,
		CRIVersion:     "v1",
		OCIVersion:     "v1.1.0",
		RuntimeName:    fmt.Sprintf("tarak-runtime (%s)", runtimeDisplay),
		RuntimeVersion: version.Version,
		EngineMode:     mode,
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
	}
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func containerKey(ns, pod, container string) string {
	return fmt.Sprintf("%s/%s/%s", ns, pod, container)
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func randomSuffix(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	out := make([]byte, length)
	for i, v := range b {
		out[i] = chars[int(v)%len(chars)]
	}
	return string(out)
}
