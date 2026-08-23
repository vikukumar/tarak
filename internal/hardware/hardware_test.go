package hardware

import (
	"testing"
)

func TestDetectHost(t *testing.T) {
	info := DetectHost()

	if info.CPUCores <= 0 {
		t.Fatalf("expected CPUCores > 0, got %d", info.CPUCores)
	}

	if info.TotalMemoryMB <= 0 {
		t.Fatalf("expected TotalMemoryMB > 0, got %d", info.TotalMemoryMB)
	}

	t.Logf("Detected Host Hardware:\n- CPU: %d cores (%s)\n- RAM: %d MB (%s)\n- GPU: %d GPUs (models: %v)\n- Disk: %d GB total (%d GB free)\n- OS: %s (%s)",
		info.CPUCores, info.CPUModel, info.TotalMemoryMB, info.TotalMemoryGB, info.GPUCount, info.GPUModels, info.TotalDiskGB, info.FreeDiskGB, info.OSName, info.OSVersion)

	// Test 100% full allocation default
	alloc := ComputeAllocation(info, "", "", "")
	if !alloc.IsFullHost {
		t.Errorf("expected IsFullHost=true by default")
	}
	if alloc.CPUCores == "" || alloc.MemoryMi == "" {
		t.Errorf("invalid alloc: %+v", alloc)
	}

	// Test user overrides
	userAlloc := ComputeAllocation(info, "4", "8192Mi", "1")
	if userAlloc.IsFullHost {
		t.Errorf("expected IsFullHost=false with custom limits")
	}
	if userAlloc.CPUCores != "4" || userAlloc.MemoryMi != "8192Mi" || userAlloc.GPU != "1" {
		t.Errorf("unexpected userAlloc: %+v", userAlloc)
	}
}
