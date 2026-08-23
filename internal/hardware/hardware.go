package hardware

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

// HostInfo contains full detected physical hardware metrics for the host machine.
type HostInfo struct {
	// CPU
	CPUCores     int    `json:"cpuCores"`     // Total logical CPU cores (e.g. 16)
	CPUModel     string `json:"cpuModel"`     // Model description (e.g. Intel Core i7 / AMD Ryzen 9)
	Architecture string `json:"architecture"` // Architecture (amd64, arm64)

	// Memory
	TotalMemoryBytes uint64 `json:"totalMemoryBytes"` // Total Physical RAM in bytes
	TotalMemoryMB    uint64 `json:"totalMemoryMB"`    // Total Physical RAM in MiB
	TotalMemoryGB    string `json:"totalMemoryGB"`    // Total Physical RAM in GiB string (e.g. "32.00 GiB")

	// GPU
	GPUCount  int      `json:"gpuCount"`  // Number of detected GPUs
	GPUModels []string `json:"gpuModels"` // Names of detected GPUs
	HasGPU    bool     `json:"hasGpu"`    // Whether at least 1 GPU is active

	// Ephemeral Storage / Disk
	TotalDiskGB uint64 `json:"totalDiskGB"` // Total root disk capacity in GiB
	FreeDiskGB  uint64 `json:"freeDiskGB"`  // Available root disk in GiB

	// Operating System
	OSName    string `json:"osName"`    // Windows, Linux, Darwin
	OSVersion string `json:"osVersion"` // OS release / build version
}

// ResourceAllocation defines the allocated compute limits for the cluster node.
type ResourceAllocation struct {
	CPUCores    string // e.g. "16"
	MemoryMi    string // e.g. "32768Mi"
	GPU         string // e.g. "1" or "0"
	DiskGi      string // e.g. "500Gi"
	IsFullHost  bool   // true if occupying 100% of host machine hardware
}

// DetectHost returns the detected physical hardware of the host system.
func DetectHost() HostInfo {
	info := HostInfo{
		CPUCores:     runtime.NumCPU(),
		Architecture: runtime.GOARCH,
		OSName:       runtime.GOOS,
		OSVersion:    runtime.Version(),
	}

	// Platform-specific physical memory, GPU, disk, and CPU model detection
	detectPlatformHardware(&info)

	if info.TotalMemoryMB == 0 && info.TotalMemoryBytes > 0 {
		info.TotalMemoryMB = info.TotalMemoryBytes / (1024 * 1024)
	}
	if info.TotalMemoryGB == "" && info.TotalMemoryMB > 0 {
		info.TotalMemoryGB = fmt.Sprintf("%.1f GiB", float64(info.TotalMemoryMB)/1024.0)
	}

	info.HasGPU = info.GPUCount > 0
	return info
}

// ComputeAllocation computes the cluster node resources.
// If user overrides are empty, it defaults to 100% of the host machine's total CPU, RAM, GPU, and Storage!
func ComputeAllocation(info HostInfo, userCPULimit, userMemLimit, userGPULimit string) ResourceAllocation {
	alloc := ResourceAllocation{
		IsFullHost: true,
	}

	// 1. CPU
	if userCPULimit != "" {
		alloc.CPUCores = strings.TrimSpace(userCPULimit)
		alloc.IsFullHost = false
	} else {
		alloc.CPUCores = strconv.Itoa(info.CPUCores)
	}

	// 2. Memory
	if userMemLimit != "" {
		alloc.MemoryMi = strings.TrimSpace(userMemLimit)
		if !strings.HasSuffix(alloc.MemoryMi, "Mi") && !strings.HasSuffix(alloc.MemoryMi, "Gi") && !strings.HasSuffix(alloc.MemoryMi, "M") && !strings.HasSuffix(alloc.MemoryMi, "G") {
			alloc.MemoryMi += "Mi"
		}
		alloc.IsFullHost = false
	} else {
		if info.TotalMemoryMB > 0 {
			alloc.MemoryMi = fmt.Sprintf("%dMi", info.TotalMemoryMB)
		} else {
			alloc.MemoryMi = "16384Mi"
		}
	}

	// 3. GPU
	if userGPULimit != "" {
		alloc.GPU = strings.TrimSpace(userGPULimit)
		alloc.IsFullHost = false
	} else {
		alloc.GPU = strconv.Itoa(info.GPUCount)
	}

	// 4. Storage
	if info.TotalDiskGB > 0 {
		alloc.DiskGi = fmt.Sprintf("%dGi", info.TotalDiskGB)
	} else {
		alloc.DiskGi = "250Gi"
	}

	return alloc
}
