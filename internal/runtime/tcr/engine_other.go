//go:build !linux && !windows

package tcr

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func platformSupported() bool { return true }

func platformNote() string {
	return fmt.Sprintf(
		"%s/%s: TCR uses the Native Bridge Runtime.\n"+
			"  - Web server images (nginx, apache, caddy): built-in Go HTTP server\n"+
			"  - Multi-arch images (%s/%s): native process execution\n"+
			"  - WASM images: embedded WASM runner\n"+
			"  - Linux-only images: install Docker Desktop or Rancher Desktop (auto-detected)",
		runtime.GOOS, runtime.GOARCH, runtime.GOOS, runtime.GOARCH,
	)
}

// newOSCommand creates an exec.Cmd appropriate for the current non-Linux/Windows OS.
func newOSCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

// platformStart starts a container on macOS (or other non-Linux/Windows platforms).
// Priority:
//  1. Bridge runtime (nginx→HTTP server, static→HTTP server, WASM→WASM runner)
//  2. macOS/native binary from rootfs (Mach-O or platform-native binary)
//  3. Clear error with actionable instructions
func platformStart(ctx context.Context, cfg ContainerConfig, logFilePath string) (*Process, error) {
	if len(cfg.Command) == 0 && cfg.Rootfs == "" {
		return nil, fmt.Errorf("no command and no rootfs specified")
	}

	// Container ports come from cfg.Ports (set by runContainerViaTCR in runtime.go)
	ports := cfg.Ports
	if len(ports) == 0 {
		ports = []int{80}
	}

	// Check image type — route web/WASM images to the bridge
	imgType := DetectImageType(cfg.Rootfs, cfg.Command)
	switch imgType {
	case ImageTypeNginx, ImageTypeApache, ImageTypeCaddy, ImageTypeLighttpd, ImageTypeStaticSite:
		return StartBridgeContainer(ctx, cfg, ports, logFilePath)
	case ImageTypeWASM:
		return StartBridgeContainer(ctx, cfg, ports, logFilePath)
	}

	// Try to run a native binary from the rootfs
	if cfg.Rootfs != "" {
		if proc, err := tryNativeBinaryOnPlatform(ctx, cfg, logFilePath); err == nil {
			return proc, nil
		}
	}

	// Try the bridge for any remaining image types
	if proc, err := StartBridgeContainer(ctx, cfg, ports, logFilePath); err == nil {
		return proc, nil
	}

	// Surface a clear error
	return nil, fmt.Errorf(
		"TCR %s: unable to run container natively.\n"+
			"Image: %s\nRootfs: %s\n\n"+
			"Linux ELF binaries require a Linux kernel.\n\n"+
			"Solutions for %s:\n"+
			"  1. Use a web server image (nginx, caddy) — TCR serves it natively\n"+
			"  2. Use a multi-arch image with %s/%s layers\n"+
			"  3. Install Docker Desktop or Rancher Desktop — Tarak auto-detects it\n"+
			"  4. Deploy on Linux — full native TCR container support\n\n"+
			"TCR note: %s",
		runtime.GOOS, cfg.ID, cfg.Rootfs, runtime.GOOS, runtime.GOOS, runtime.GOARCH, platformNote(),
	)
}

// tryNativeBinaryOnPlatform looks for a Mach-O or platform-native binary in the rootfs.
func tryNativeBinaryOnPlatform(ctx context.Context, cfg ContainerConfig, logFilePath string) (*Process, error) {
	logFile, _ := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)

	if len(cfg.Command) == 0 {
		if logFile != nil {
			logFile.Close()
		}
		return nil, fmt.Errorf("no command")
	}

	entryBin := filepath.Base(cfg.Command[0])
	searchDirs := []string{
		cfg.Rootfs,
		filepath.Join(cfg.Rootfs, "bin"),
		filepath.Join(cfg.Rootfs, "usr", "bin"),
		filepath.Join(cfg.Rootfs, "usr", "local", "bin"),
	}

	for _, dir := range searchDirs {
		candidate := filepath.Join(dir, entryBin)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			// Don't try Linux ELF on macOS
			if isLinuxELFFile(candidate) {
				continue
			}
			// Check executable permission
			if info.Mode()&0111 != 0 {
				return startNativeProcess(ctx, cfg.ID, candidate, cfg.Command[1:], cfg.Env, cfg.WorkingDir, logFile)
			}
		}
	}

	if logFile != nil {
		logFile.Close()
	}
	return nil, fmt.Errorf("no native binary found")
}

func platformWait(proc *Process) {
	for proc.IsRunning() {
		time.Sleep(500 * time.Millisecond)
	}
}

func platformStop(proc *Process) error {
	// Goroutine-based containers (HTTP server, WASM)
	if proc.cancel != nil {
		proc.cancel()
	}
	// OS-process containers
	if proc.PID > 0 {
		p, err := os.FindProcess(proc.PID)
		if err == nil {
			_ = p.Kill()
		}
	}
	proc.SetState("exited", 0)
	return nil
}

func platformExec(ctx context.Context, proc *Process, cfg ExecConfig, stdin io.Reader, stdout, stderr io.Writer, tty bool) (int, error) {
	if len(cfg.Command) == 0 {
		cfg.Command = []string{"sh"}
	}

	shell := "/bin/sh"
	if s := os.Getenv("SHELL"); s != "" {
		shell = s
	}

	cmd := exec.CommandContext(ctx, shell, append([]string{"-c"}, strings.Join(cfg.Command, " "))...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if len(cfg.Env) > 0 {
		cmd.Env = cfg.Env
	}
	if cfg.WorkingDir != "" {
		cmd.Dir = cfg.WorkingDir
	}
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), err
		}
		return 1, err
	}
	return 0, nil
}

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

func DetectRootfs(dataDir, imageHash string) string {
	return filepath.Join(dataDir, "images", imageHash, "rootfs")
}
