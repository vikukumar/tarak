//go:build windows

package hardware

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

type memoryStatusEx struct {
	cbSize                  uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

var (
	kernel32              = syscall.NewLazyDLL("kernel32.dll")
	globalMemoryStatusEx  = kernel32.NewProc("GlobalMemoryStatusEx")
	getDiskFreeSpaceExW   = kernel32.NewProc("GetDiskFreeSpaceExW")
)

func detectPlatformHardware(info *HostInfo) {
	// 1. Physical RAM
	var memStatus memoryStatusEx
	memStatus.cbSize = uint32(unsafe.Sizeof(memStatus))
	ret, _, _ := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memStatus)))
	if ret != 0 {
		info.TotalMemoryBytes = memStatus.ullTotalPhys
		info.TotalMemoryMB = memStatus.ullTotalPhys / (1024 * 1024)
		info.TotalMemoryGB = fmt.Sprintf("%.1f GiB", float64(memStatus.ullTotalPhys)/(1024.0*1024.0*1024.0))
	}

	// 2. Disk Space on system root (C:\)
	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64
	rootPath, err := syscall.UTF16PtrFromString("C:\\")
	if err == nil {
		r, _, _ := getDiskFreeSpaceExW.Call(
			uintptr(unsafe.Pointer(rootPath)),
			uintptr(unsafe.Pointer(&freeBytesAvailable)),
			uintptr(unsafe.Pointer(&totalNumberOfBytes)),
			uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
		)
		if r != 0 {
			info.TotalDiskGB = totalNumberOfBytes / (1024 * 1024 * 1024)
			info.FreeDiskGB = freeBytesAvailable / (1024 * 1024 * 1024)
		}
	}

	// 3. GPU Detection via nvidia-smi or powershell WMI
	detectWindowsGPU(info)

	// 4. CPU Model
	detectWindowsCPU(info)
}

func detectWindowsGPU(info *HostInfo) {
	// Check nvidia-smi first
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

	// Fallback to PowerShell Get-CimInstance Win32_VideoController
	psCmd := `Get-CimInstance Win32_VideoController | Where-Object { $_.Name -notlike "*Basic*" -and $_.Name -notlike "*Remote*" } | Select-Object -ExpandProperty Name`
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		lines := strings.Split(strings.TrimSpace(out.String()), "\r\n")
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

	// Default integrated or virtual GPU check
	info.GPUCount = 0
	info.GPUModels = nil
}

func detectWindowsCPU(info *HostInfo) {
	psCmd := `(Get-CimInstance Win32_Processor | Select-Object -First 1).Name`
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		model := strings.TrimSpace(out.String())
		if model != "" {
			info.CPUModel = model
			return
		}
	}
	info.CPUModel = fmt.Sprintf("Host CPU (%d Cores)", info.CPUCores)
}
