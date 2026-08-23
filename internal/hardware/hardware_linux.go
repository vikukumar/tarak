//go:build linux

package hardware

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func detectPlatformHardware(info *HostInfo) {
	// 1. Memory from /proc/meminfo
	if f, err := os.Open("/proc/meminfo"); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "MemTotal:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					if kb, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
						info.TotalMemoryBytes = kb * 1024
						info.TotalMemoryMB = kb / 1024
						info.TotalMemoryGB = fmt.Sprintf("%.1f GiB", float64(info.TotalMemoryMB)/1024.0)
					}
				}
				break
			}
		}
	}

	// 2. CPU Model from /proc/cpuinfo
	if f, err := os.Open("/proc/cpuinfo"); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "model name") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					info.CPUModel = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}
	if info.CPUModel == "" {
		info.CPUModel = fmt.Sprintf("Linux Host (%d Cores)", info.CPUCores)
	}

	// 3. Disk Space via statfs
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err == nil {
		totalBytes := stat.Blocks * uint64(stat.Bsize)
		freeBytes := stat.Bavail * uint64(stat.Bsize)
		info.TotalDiskGB = totalBytes / (1024 * 1024 * 1024)
		info.FreeDiskGB = freeBytes / (1024 * 1024 * 1024)
	}

	// 4. GPU Detection via nvidia-smi or lspci
	if out, err := exec.Command("nvidia-smi", "--query-gpu=name,memory.total", "--format=csv,noheader").Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		var gpus []string
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l != "" {
				gpus = append(gpus, l)
			}
		}
		if len(gpus) > 0 {
			info.GPUCount = len(gpus)
			info.GPUModels = gpus
			return
		}
	}

	if out, err := exec.Command("lspci").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		var gpus []string
		for _, l := range lines {
			if strings.Contains(strings.ToLower(l), "vga") || strings.Contains(strings.ToLower(l), "3d controller") {
				parts := strings.Split(l, ":")
				if len(parts) >= 3 {
					gpus = append(gpus, strings.TrimSpace(parts[2]))
				}
			}
		}
		if len(gpus) > 0 {
			info.GPUCount = len(gpus)
			info.GPUModels = gpus
		}
	}
}
