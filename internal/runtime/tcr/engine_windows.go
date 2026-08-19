//go:build windows

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
	"syscall"
	"time"
)

func platformSupported() bool { return true }

func platformNote() string {
	return "Windows: TCR uses the Native Bridge Runtime.\n" +
		"  - Web server images (nginx, apache, caddy): built-in Go HTTP server\n" +
		"  - Multi-arch images (windows/amd64): native process execution\n" +
		"  - WASM images: embedded WASM runner\n" +
		"  - Linux-only images: install Docker Desktop (Tarak auto-detects it)"
}

// newOSCommand creates an exec.Cmd with Windows-appropriate settings.
func newOSCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	// Create new process group so we can kill cleanly
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
	return cmd
}

// platformStart starts a container on Windows.
// Priority:
//  1. Bridge runtime (nginx→HTTP server, static→HTTP server, WASM→WASM runner)
//  2. Windows-native binary from rootfs
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

	// Detect image type FIRST — if it's a web server or WASM, use bridge
	imgType := DetectImageType(cfg.Rootfs, cfg.Command)
	switch imgType {
	case ImageTypeNginx, ImageTypeApache, ImageTypeCaddy, ImageTypeLighttpd, ImageTypeStaticSite:
		return StartBridgeContainer(ctx, cfg, ports, logFilePath)
	case ImageTypeWASM:
		return StartBridgeContainer(ctx, cfg, ports, logFilePath)
	}

	// Try to find a Windows-native executable in the rootfs
	if cfg.Rootfs != "" {
		if proc, err := tryWindowsNativeExec(ctx, cfg, logFilePath); err == nil {
			return proc, nil
		}
	}

	// Try the bridge for any remaining cases (may find static files, WASM, etc.)
	if proc, err := StartBridgeContainer(ctx, cfg, ports, logFilePath); err == nil {
		return proc, nil
	}

	// Nothing worked — surface a clear message
	return nil, fmt.Errorf(
		"TCR Windows: unable to run container natively.\n"+
			"Image: %s\nRootfs: %s\n\n"+
			"Linux ELF binaries require a Linux kernel (WSL2 or Linux host).\n\n"+
			"Solutions for Windows:\n"+
			"  1. Use a web server image (nginx, caddy) — TCR serves it via built-in HTTP server\n"+
			"  2. Use a multi-arch image that includes windows/amd64 layers\n"+
			"  3. Install Docker Desktop — Tarak auto-detects it on next start\n"+
			"  4. Deploy to a Linux node — full native containers supported\n\n"+
			"Current: %s/%s",
		cfg.ID, cfg.Rootfs, runtime.GOOS, runtime.GOARCH,
	)
}

// tryWindowsNativeExec attempts to run a Windows-native binary from the rootfs.
func tryWindowsNativeExec(ctx context.Context, cfg ContainerConfig, logFilePath string) (*Process, error) {
	logFile, _ := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)

	entryCmd := cfg.Command
	if len(entryCmd) == 0 {
		return nil, fmt.Errorf("no command")
	}
	entryBin := filepath.Base(entryCmd[0])
	searchDirs := []string{
		cfg.Rootfs,
		filepath.Join(cfg.Rootfs, "bin"),
		filepath.Join(cfg.Rootfs, "usr", "bin"),
		filepath.Join(cfg.Rootfs, "usr", "local", "bin"),
	}

	winExts := []string{".exe", ".cmd", ".bat", ""}
	for _, dir := range searchDirs {
		for _, ext := range winExts {
			candidate := filepath.Join(dir, entryBin+ext)
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				// Check it's not a Linux ELF
				if isLinuxELFFile(candidate) {
					continue
				}
				return startNativeProcess(ctx, cfg.ID, candidate, entryCmd[1:], cfg.Env, cfg.WorkingDir, logFile)
			}
		}
	}
	if logFile != nil {
		logFile.Close()
	}
	return nil, fmt.Errorf("no Windows-native binary found in rootfs")
}

func platformWait(proc *Process) {
	for proc.IsRunning() {
		time.Sleep(500 * time.Millisecond)
	}
}

func platformStop(proc *Process) error {
	// For goroutine-based containers (HTTP server, WASM runner)
	if proc.cancel != nil {
		proc.cancel()
	}
	// For OS-process containers
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
		cfg.Command = []string{"cmd.exe"}
	}

	comspec := os.Getenv("COMSPEC")
	if comspec == "" {
		comspec = "C:\\Windows\\System32\\cmd.exe"
	}
	args := append([]string{"/c"}, cfg.Command...)
	cmd := exec.CommandContext(ctx, comspec, args...)
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
