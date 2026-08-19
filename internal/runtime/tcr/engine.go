// Package tcr implements the Tarak Container Runtime — a built-from-scratch
// container execution engine that requires NO external tools (no Docker, Podman,
// or nerdctl).
//
// On Linux: uses kernel namespaces directly (PID, Mount, UTS, User) for real
// container isolation equivalent to Docker's runc.
//
// On Windows: uses Windows Job Objects for process isolation. Linux images
// (ELF binaries) require a Linux kernel and cannot run natively.
//
// On macOS: uses process isolation. Linux images require a Linux VM.
//
// The reexec pattern:
//
//	Parent (tarak server) spawns: tarak __tcr_init__ [with namespace flags on Linux]
//	Child reads TARAK_TCR_CONFIG, mounts /proc /sys /dev, chroots, exec's the app.
//
// This is exactly how Docker's runc works internally — Tarak implements it natively.
package tcr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"
)

// Special argv[1] values that trigger internal TCR modes.
const (
	InitArg = "__tcr_init__"
	ExecArg = "__tcr_exec__"
)

// Environment variable keys for passing config between parent and child.
const (
	EnvConfig     = "TARAK_TCR_CONFIG"
	EnvExecConfig = "TARAK_TCR_EXEC_CONFIG"
)

// ContainerConfig is the complete specification for launching a container.
type ContainerConfig struct {
	ID         string   `json:"id"`
	Rootfs     string   `json:"rootfs"`
	Command    []string `json:"command"`
	Env        []string `json:"env"`
	WorkingDir string   `json:"workingDir"`
	Hostname   string   `json:"hostname"`
	// Ports is the list of container ports to expose (used by the native bridge HTTP server).
	Ports []int `json:"ports"`
}

// ExecConfig specifies a command to run inside an already-running container.
type ExecConfig struct {
	TargetPID  int      `json:"targetPid"`
	Rootfs     string   `json:"rootfs"`
	Command    []string `json:"command"`
	Env        []string `json:"env"`
	WorkingDir string   `json:"workingDir"`
	Tty        bool     `json:"tty"`
}

// Process tracks a running container process.
type Process struct {
	ID        string
	PID       int
	Rootfs    string
	StartedAt time.Time
	mu        sync.Mutex
	state     string
	exitCode  int
	// cancel is set for goroutine-based containers (built-in HTTP server, WASM).
	// Calling it stops the container cleanly without needing an OS PID.
	cancel context.CancelFunc
}

func (p *Process) SetState(state string, code int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = state
	p.exitCode = code
}

func (p *Process) State() (string, int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state, p.exitCode
}

func (p *Process) IsRunning() bool {
	s, _ := p.State()
	return s == "running"
}

// Engine is the Tarak Container Runtime engine.
type Engine struct {
	mu        sync.RWMutex
	processes map[string]*Process
}

func New() *Engine {
	return &Engine{processes: make(map[string]*Process)}
}

func (e *Engine) register(id string, proc *Process) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.processes[id] = proc
}

func (e *Engine) Get(id string) (*Process, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	proc, ok := e.processes[id]
	return proc, ok
}

func (e *Engine) remove(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.processes, id)
}

// SupportedOnThisOS returns true if TCR can run Linux containers on this OS.
func SupportedOnThisOS() bool { return platformSupported() }

// OSNote returns a human-readable description of TCR capabilities on the current OS.
func OSNote() string { return platformNote() }

// StartContainer starts a container using the native runtime.
func (e *Engine) StartContainer(ctx context.Context, cfg ContainerConfig, logFilePath string) (*Process, error) {
	proc, err := platformStart(ctx, cfg, logFilePath)
	if err != nil {
		return nil, err
	}
	e.register(cfg.ID, proc)
	go func() {
		platformWait(proc)
		e.remove(cfg.ID)
	}()
	return proc, nil
}

// StopContainer stops a running container.
func (e *Engine) StopContainer(id string) error {
	proc, ok := e.Get(id)
	if !ok {
		return nil
	}
	err := platformStop(proc)
	e.remove(id)
	return err
}

// ExecInContainer runs a command inside a running container's namespace.
func (e *Engine) ExecInContainer(ctx context.Context, id string, cmd []string, env []string, workdir string, stdin io.Reader, stdout, stderr io.Writer, tty bool) (int, error) {
	proc, ok := e.Get(id)
	if !ok {
		return -1, fmt.Errorf("container %s: not found or not running", id)
	}
	cfg := ExecConfig{
		TargetPID:  proc.PID,
		Rootfs:     proc.Rootfs,
		Command:    cmd,
		Env:        env,
		WorkingDir: workdir,
		Tty:        tty,
	}
	return platformExec(ctx, proc, cfg, stdin, stdout, stderr, tty)
}

func MarshalConfig(cfg ContainerConfig) (string, error) {
	b, err := json.Marshal(cfg)
	return string(b), err
}

func UnmarshalConfig(s string) (ContainerConfig, error) {
	var cfg ContainerConfig
	return cfg, json.Unmarshal([]byte(s), &cfg)
}

func MarshalExecConfig(cfg ExecConfig) (string, error) {
	b, err := json.Marshal(cfg)
	return string(b), err
}

func UnmarshalExecConfig(s string) (ExecConfig, error) {
	var cfg ExecConfig
	return cfg, json.Unmarshal([]byte(s), &cfg)
}
