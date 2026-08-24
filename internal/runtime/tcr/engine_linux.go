//go:build linux

package tcr

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func platformSupported() bool { return true }

func platformNote() string {
	return "Linux: real container isolation via kernel namespaces (PID + Mount + UTS + User). No Docker required."
}

// newOSCommand creates an exec.Cmd appropriate for Linux.
func newOSCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

// platformStart spawns tarak __tcr_init__ with Linux namespace flags.
// The re-executed init process mounts /proc /sys /dev, chroots, and exec's the app.
func platformStart(ctx context.Context, cfg ContainerConfig, logFilePath string) (*Process, error) {
	// Resolve path to the tarak binary (the current executable)
	self, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("resolve tarak binary path: %w", err)
	}

	// Serialize the container config to pass to the init process
	cfgJSON, err := MarshalConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal container config: %w", err)
	}

	// Open (or create) the container log file
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file %s: %w", logFilePath, err)
	}

	// Build the init command: tarak __tcr_init__
	// We pass NO cobra flags — the init handler in main() checks argv[1] before cobra.
	cmd := exec.CommandContext(ctx, self, InitArg)

	// Pass container config via environment
	cmd.Env = append(os.Environ(), EnvConfig+"="+cfgJSON)

	// Redirect stdout/stderr to the container log file
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// ── Linux namespace isolation ─────────────────────────────────────────────
	//
	// CLONE_NEWPID  — Container gets its own PID namespace; container process is PID 1
	// CLONE_NEWNS   — Container gets its own mount namespace; chroot becomes real
	// CLONE_NEWUTS  — Container gets its own hostname
	// CLONE_NEWIPC  — Container gets its own IPC namespace
	// CLONE_NEWUSER — User namespace: container root = current host user (rootless!)
	//
	// With CLONE_NEWUSER + UID/GID mappings:
	//   - Container process sees itself as uid 0 (root)
	//   - On the host it runs as the current user
	//   - This allows mounting proc/sys/dev without actual root privileges
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWUSER,
		// Map container root (uid 0) to the current host user
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getgid(), Size: 1},
		},
		// Ensure container process is killed when tarak server dies
		Pdeathsig: syscall.SIGKILL,
	}

	if err := cmd.Start(); err != nil {
		logFile.Close()
		// Fallback to Native Bridge runtime if namespace clone is disallowed by kernel/sysctl
		return StartBridgeContainer(ctx, cfg, cfg.Ports, logFilePath)
	}

	proc := &Process{
		ID:        cfg.ID,
		PID:       cmd.Process.Pid,
		Rootfs:    cfg.Rootfs,
		StartedAt: time.Now().UTC(),
		state:     "running",
	}

	// Monitor the process in a goroutine and update state when it exits
	go func() {
		defer logFile.Close()
		state, err := cmd.Process.Wait()
		exitCode := 0
		if err != nil {
			exitCode = 1
		} else if state != nil && !state.Success() {
			exitCode = state.ExitCode()
		}
		proc.SetState("exited", exitCode)
	}()

	return proc, nil
}

func platformWait(proc *Process) {
	// State is managed inside platformStart's goroutine
	for proc.IsRunning() {
		time.Sleep(500 * time.Millisecond)
	}
}

func platformStop(proc *Process) error {
	// For goroutine-based containers (built-in HTTP server running on Linux)
	if proc.cancel != nil {
		proc.cancel()
	}
	if proc.PID <= 0 {
		proc.SetState("exited", 0)
		return nil
	}
	p, err := os.FindProcess(proc.PID)
	if err != nil {
		return nil
	}
	// Send SIGTERM first, then SIGKILL if needed
	_ = p.Signal(syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 20; i++ {
			if !proc.IsRunning() {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
		_ = p.Signal(syscall.SIGKILL)
	}()
	<-done
	proc.SetState("exited", 0)
	return nil
}

// platformExec enters the container's namespace and runs a command inside it.
// It uses nsenter if available (standard util-linux tool), otherwise falls
// back to re-executing tarak with __tcr_exec__.
func platformExec(ctx context.Context, proc *Process, cfg ExecConfig, stdin io.Reader, stdout, stderr io.Writer, tty bool) (int, error) {
	if len(cfg.Command) == 0 {
		cfg.Command = []string{"/bin/sh"}
	}

	// Strategy 1: nsenter (part of util-linux, available on all major Linux distros)
	// This is the most reliable way to enter all namespaces of a running container.
	nsenterPath, nsenterErr := exec.LookPath("nsenter")
	if nsenterErr == nil {
		args := []string{
			fmt.Sprintf("--target=%d", cfg.TargetPID),
			"--mount",
			"--uts",
			"--ipc",
			"--pid",
			"--",
		}
		args = append(args, cfg.Command...)

		execCmd := exec.CommandContext(ctx, nsenterPath, args...)
		execCmd.Stdin = stdin
		execCmd.Stdout = stdout
		execCmd.Stderr = stderr
		if cfg.WorkingDir != "" {
			execCmd.Dir = cfg.WorkingDir
		}
		if len(cfg.Env) > 0 {
			execCmd.Env = cfg.Env
		}
		if err := execCmd.Run(); err != nil {
			return 1, err
		}
		return 0, nil
	}

	// Strategy 2: re-execute tarak as __tcr_exec__ to enter mount namespace
	// via /proc/<pid>/ns/mnt and setns syscall.
	self, _ := os.Executable()
	execCfgJSON, _ := MarshalExecConfig(cfg)

	execCmd := exec.CommandContext(ctx, self, ExecArg)
	execCmd.Env = append(os.Environ(), EnvExecConfig+"="+execCfgJSON)
	execCmd.Stdin = stdin
	execCmd.Stdout = stdout
	execCmd.Stderr = stderr

	if err := execCmd.Run(); err != nil {
		return 1, err
	}
	return 0, nil
}

// GetLogs reads the container log file and writes to out.
func GetContainerLogs(logFilePath string, follow bool, tail int, ctx context.Context, out io.Writer) error {
	data, err := os.ReadFile(logFilePath)
	if err != nil {
		return fmt.Errorf("read container logs: %w", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if tail > 0 && tail < len(lines) {
		lines = lines[len(lines)-tail:]
	}
	for _, l := range lines {
		fmt.Fprintln(out, l)
	}
	if follow {
		<-ctx.Done()
	}
	return nil
}

// detectRootfs returns the image rootfs path for a given image hash.
func DetectRootfs(dataDir, imageHash string) string {
	return filepath.Join(dataDir, "images", imageHash, "rootfs")
}
