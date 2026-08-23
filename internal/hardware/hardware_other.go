//go:build !windows && !linux && !darwin

package hardware

import "fmt"

func detectPlatformHardware(info *HostInfo) {
	info.CPUModel = fmt.Sprintf("Generic Host (%d Cores)", info.CPUCores)
	info.TotalMemoryMB = 16384
	info.TotalMemoryGB = "16.0 GiB"
	info.TotalDiskGB = 250
	info.FreeDiskGB = 100
}
