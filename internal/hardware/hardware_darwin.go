//go:build darwin

package hardware

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func detectPlatformHardware(info *HostInfo) {
	// 1. Memory via sysctl hw.memsize
	if out, err := exec.Command("sysctl", "-n", "hw.memsize").Output(); err == nil {
		if bytes, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64); err == nil {
			info.TotalMemoryBytes = bytes
			info.TotalMemoryMB = bytes / (1024 * 1024)
			info.TotalMemoryGB = fmt.Sprintf("%.1f GiB", float64(info.TotalMemoryMB)/1024.0)
		}
	}

	// 2. CPU Model via sysctl machdep.cpu.brand_string
	if out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
		model := strings.TrimSpace(string(out))
		if model != "" {
			info.CPUModel = model
		}
	}
	if info.CPUModel == "" {
		info.CPUModel = fmt.Sprintf("Apple Silicon/Mac (%d Cores)", info.CPUCores)
	}

	// 3. Disk Space via statfs
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err == nil {
		totalBytes := stat.Blocks * uint64(stat.Bsize)
		freeBytes := stat.Bavail * uint64(stat.Bsize)
		info.TotalDiskGB = totalBytes / (1024 * 1024 * 1024)
		info.FreeDiskGB = freeBytes / (1024 * 1024 * 1024)
	}

	// 4. GPU Detection via system_profiler SPDisplaysDataType
	if out, err := exec.Command("system_profiler", "SPDisplaysDataType").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		var gpus []string
		for _, l := range lines {
			if strings.Contains(l, "Chipset Model:") {
				parts := strings.Split(l, ":")
				if len(parts) >= 2 {
					gpus = append(gpus, strings.TrimSpace(parts[1]))
				}
			}
		}
		if len(gpus) > 0 {
			info.GPUCount = len(gpus)
			info.GPUModels = gpus
		}
	}
}
