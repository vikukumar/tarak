//go:build linux

package tcr

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// CgroupLimits configures resource constraints for container processes.
type CgroupLimits struct {
	CPUMilliCores int64 // CPU quota in millicores (e.g. 500 = 0.5 CPU)
	MemoryBytes   int64 // Hard memory limit in bytes (e.g. 512MiB)
	PIDsMax       int64 // Maximum processes/threads inside container
}

// CgroupManager configures Linux kernel cgroups (supporting both cgroups v2 and v1).
type CgroupManager struct {
	cgroupPath string
	isV2       bool
}

// NewCgroupManager detects cgroup version and creates container cgroup slice.
func NewCgroupManager(containerID string) (*CgroupManager, error) {
	// 1. Check for cgroups v2 unified hierarchy
	v2Mount := "/sys/fs/cgroup"
	if _, err := os.Stat(filepath.Join(v2Mount, "cgroup.controllers")); err == nil {
		cPath := filepath.Join(v2Mount, "tarak", containerID)
		if err := os.MkdirAll(cPath, 0755); err != nil {
			return nil, fmt.Errorf("create cgroup v2 path: %w", err)
		}
		return &CgroupManager{cgroupPath: cPath, isV2: true}, nil
	}

	// 2. Fallback to cgroups v1
	v1Path := filepath.Join("/sys/fs/cgroup/cpu,cpuacct/tarak", containerID)
	_ = os.MkdirAll(v1Path, 0755)
	_ = os.MkdirAll(filepath.Join("/sys/fs/cgroup/memory/tarak", containerID), 0755)
	_ = os.MkdirAll(filepath.Join("/sys/fs/cgroup/pids/tarak", containerID), 0755)

	return &CgroupManager{cgroupPath: v1Path, isV2: false}, nil
}

// ApplyLimits writes CPU, Memory, and PIDs limits to the cgroup filesystem.
func (cm *CgroupManager) ApplyLimits(limits CgroupLimits) error {
	if cm.isV2 {
		// cgroups v2: cpu.max ("quota period")
		if limits.CPUMilliCores > 0 {
			quota := limits.CPUMilliCores * 100 // 100,000us per core
			_ = os.WriteFile(filepath.Join(cm.cgroupPath, "cpu.max"), []byte(fmt.Sprintf("%d 100000", quota)), 0644)
		}
		// cgroups v2: memory.max
		if limits.MemoryBytes > 0 {
			_ = os.WriteFile(filepath.Join(cm.cgroupPath, "memory.max"), []byte(strconv.FormatInt(limits.MemoryBytes, 10)), 0644)
		}
		// cgroups v2: pids.max
		if limits.PIDsMax > 0 {
			_ = os.WriteFile(filepath.Join(cm.cgroupPath, "pids.max"), []byte(strconv.FormatInt(limits.PIDsMax, 10)), 0644)
		}
	} else {
		// cgroups v1: cpu.cfs_quota_us / cpu.cfs_period_us
		if limits.CPUMilliCores > 0 {
			quota := limits.CPUMilliCores * 100
			_ = os.WriteFile(filepath.Join(cm.cgroupPath, "cpu.cfs_period_us"), []byte("100000"), 0644)
			_ = os.WriteFile(filepath.Join(cm.cgroupPath, "cpu.cfs_quota_us"), []byte(strconv.FormatInt(quota, 10)), 0644)
		}
	}
	return nil
}

// AttachProcess adds a process PID to the container cgroup.
func (cm *CgroupManager) AttachProcess(pid int) error {
	pidStr := strconv.Itoa(pid)
	if cm.isV2 {
		return os.WriteFile(filepath.Join(cm.cgroupPath, "cgroup.procs"), []byte(pidStr), 0644)
	}
	_ = os.WriteFile(filepath.Join(cm.cgroupPath, "cgroup.procs"), []byte(pidStr), 0644)
	return nil
}

// Cleanup removes the container cgroup slice.
func (cm *CgroupManager) Cleanup() {
	if cm.cgroupPath != "" {
		_ = os.Remove(cm.cgroupPath)
	}
}
