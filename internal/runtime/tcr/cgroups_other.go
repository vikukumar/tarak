//go:build !linux

package tcr

// CgroupLimits configures resource constraints for container processes.
type CgroupLimits struct {
	CPUMilliCores int64
	MemoryBytes   int64
	PIDsMax       int64
}

// CgroupManager stub for non-Linux platforms.
type CgroupManager struct{}

func NewCgroupManager(containerID string) (*CgroupManager, error) {
	return &CgroupManager{}, nil
}

func (cm *CgroupManager) ApplyLimits(limits CgroupLimits) error {
	return nil
}

func (cm *CgroupManager) AttachProcess(pid int) error {
	return nil
}

func (cm *CgroupManager) Cleanup() {}
