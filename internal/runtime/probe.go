package runtime

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
)

// RuntimeType identifies the underlying container execution driver.
type RuntimeType string

const (
	RuntimeTypeDocker       RuntimeType = "docker"
	RuntimeTypePodman       RuntimeType = "podman"
	RuntimeTypeNerdctl      RuntimeType = "nerdctl"
	RuntimeTypeContainerd   RuntimeType = "containerd"
	RuntimeTypeWindowsHCS   RuntimeType = "windows-hcs"
	RuntimeTypeWSL2         RuntimeType = "wsl2"
	RuntimeTypeTCRNative    RuntimeType = "tcr-native"
)

// HostRuntimeReport contains full diagnostic information on the container engine on the host.
type HostRuntimeReport struct {
	Type        RuntimeType `json:"type"`
	Name        string      `json:"name"`
	Version     string      `json:"version"`
	BinaryPath  string      `json:"binaryPath"`
	IsAvailable bool        `json:"isAvailable"`
	Description string      `json:"description"`
	OS          string      `json:"os"`
	Arch        string      `json:"arch"`
}

// ProbeHostRuntimes scans the system in priority order and returns the active runtime report.
func ProbeHostRuntimes(log *zap.Logger) HostRuntimeReport {
	if log == nil {
		log = zap.NewNop()
	}

	// 1. Check Docker (Docker Desktop / Docker Engine)
	if rep, ok := probeCommand("docker", []string{"info", "--format", "{{.ServerVersion}}"}, RuntimeTypeDocker, "Docker Engine / Desktop"); ok {
		log.Info("detected container runtime", zap.String("type", string(rep.Type)), zap.String("version", rep.Version))
		return rep
	}

	// 2. Check Podman (Podman Desktop / Podman Engine)
	if rep, ok := probeCommand("podman", []string{"info", "--format", "{{.version.Version}}"}, RuntimeTypePodman, "Podman Engine / Desktop"); ok {
		log.Info("detected container runtime", zap.String("type", string(rep.Type)), zap.String("version", rep.Version))
		return rep
	}

	// 3. Check nerdctl (containerd / Rancher Desktop)
	if rep, ok := probeCommand("nerdctl", []string{"info"}, RuntimeTypeNerdctl, "nerdctl (containerd)"); ok {
		log.Info("detected container runtime", zap.String("type", string(rep.Type)), zap.String("version", rep.Version))
		return rep
	}

	// 4. Check Windows Containers / HCS if on Windows
	if runtime.GOOS == "windows" {
		if rep, ok := probeWindowsHCS(); ok {
			log.Info("detected Windows container runtime", zap.String("type", string(rep.Type)), zap.String("version", rep.Version))
			return rep
		}
	}

	// 5. Fallback to TCR (Tarak Container Runtime - Built-in Native Engine)
	tcrDesc := "Tarak Native Container Engine (Built-in OCI Sandbox & Process Isolation)"
	if runtime.GOOS == "linux" {
		tcrDesc = "Tarak Native Container Engine (Linux Kernel Namespaces & cgroups)"
	}

	return HostRuntimeReport{
		Type:        RuntimeTypeTCRNative,
		Name:        "Tarak Container Runtime (TCR)",
		Version:     "v1.0.6",
		BinaryPath:  "builtin://tcr",
		IsAvailable: true,
		Description: tcrDesc,
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
	}
}

func findRuntimeBinary(binary string) string {
	if binPath, err := exec.LookPath(binary); err == nil {
		return binPath
	}
	var candidates []string
	if runtime.GOOS == "windows" {
		candidates = []string{
			filepath.Join(os.Getenv("ProgramFiles"), "Docker", "Docker", "resources", "bin", binary+".exe"),
			filepath.Join(os.Getenv("ProgramFiles"), "Docker", "Docker", "resources", binary+".exe"),
			filepath.Join(os.Getenv("ProgramData"), "DockerDesktop", "version-bin", binary+".exe"),
			filepath.Join(os.Getenv("ProgramFiles"), "RedHat", "Podman", binary+".exe"),
			filepath.Join(os.Getenv("ProgramFiles"), "containerd", binary+".exe"),
			filepath.Join(os.Getenv("ProgramFiles"), "containerd", "bin", binary+".exe"),
			filepath.Join(os.Getenv("ProgramFiles"), "Rancher Desktop", "resources", "resources", "win32", "bin", binary+".exe"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Docker", "resources", "bin", binary+".exe"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Podman", binary+".exe"),
		}
	} else {
		candidates = []string{
			"/usr/bin/" + binary,
			"/usr/local/bin/" + binary,
			"/opt/homebrew/bin/" + binary,
			"/usr/local/opt/" + binary,
			"/usr/sbin/" + binary,
			"/run/current-system/sw/bin/" + binary,
		}
	}
	for _, cand := range candidates {
		if cand == "" {
			continue
		}
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
	}
	return ""
}

func probeCommand(binary string, testArgs []string, rtType RuntimeType, name string) (HostRuntimeReport, bool) {
	binPath := findRuntimeBinary(binary)
	if binPath == "" {
		return HostRuntimeReport{}, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binPath, testArgs...)
	out, err := cmd.Output()
	if err != nil {
		return HostRuntimeReport{}, false
	}

	verStr := strings.TrimSpace(string(out))
	if len(verStr) == 0 {
		verStr = "detected"
	}
	if len(verStr) > 40 {
		verStr = verStr[:40]
	}

	return HostRuntimeReport{
		Type:        rtType,
		Name:        name,
		Version:     verStr,
		BinaryPath:  binPath,
		IsAvailable: true,
		Description: fmt.Sprintf("%s active and responsive", name),
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
	}, true
}

func probeWindowsHCS() (HostRuntimeReport, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", "Get-Service vmcompute -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Status")
	out, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(out)) == "Running" {
		return HostRuntimeReport{
			Type:        RuntimeTypeWindowsHCS,
			Name:        "Windows Host Compute Service (HCS)",
			Version:     "Native Windows Container",
			BinaryPath:  "System32/vmcompute.dll",
			IsAvailable: true,
			Description: "Windows Native Container Service (vmcompute) Active",
			OS:          "windows",
			Arch:        runtime.GOARCH,
		}, true
	}
	return HostRuntimeReport{}, false
}
