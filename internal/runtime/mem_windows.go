//go:build windows

package runtime

import (
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

type fileTime struct {
	dwLowDateTime  uint32
	dwHighDateTime uint32
}

func (ft fileTime) toUint64() uint64 {
	return (uint64(ft.dwHighDateTime) << 32) | uint64(ft.dwLowDateTime)
}

var (
	kernel32DLL        = syscall.NewLazyDLL("kernel32.dll")
	procGlobalMemoryEx = kernel32DLL.NewProc("GlobalMemoryStatusEx")
	procGetSystemTimes = kernel32DLL.NewProc("GetSystemTimes")

	lastIdleTime   uint64
	lastKernelTime uint64
	lastUserTime   uint64
	hasPrevCPUTime bool
)

func getSystemMemory(total, used, avail *uint64, pct *float64) {
	var ms memoryStatusEx
	ms.cbSize = uint32(unsafe.Sizeof(ms))
	r1, _, _ := procGlobalMemoryEx.Call(uintptr(unsafe.Pointer(&ms)))
	if r1 != 0 && ms.ullTotalPhys > 0 {
		*total = ms.ullTotalPhys
		*avail = ms.ullAvailPhys
		*used = *total - *avail
		*pct = float64(ms.dwMemoryLoad)
		if *pct <= 0 {
			*pct = float64(*used) / float64(*total) * 100.0
		}
	}
}

func getSystemCPU(numCPU int, cpuPct *float64, cpuMillis *int64) {
	var idle, kernel, user fileTime
	r1, _, _ := procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&idle)),
		uintptr(unsafe.Pointer(&kernel)),
		uintptr(unsafe.Pointer(&user)),
	)
	if r1 == 0 {
		return
	}

	curIdle := idle.toUint64()
	curKernel := kernel.toUint64()
	curUser := user.toUint64()

	if hasPrevCPUTime {
		idleDelta := curIdle - lastIdleTime
		kernelDelta := curKernel - lastKernelTime
		userDelta := curUser - lastUserTime
		totalSystem := kernelDelta + userDelta

		if totalSystem > 0 {
			busyTime := totalSystem
			if busyTime > idleDelta {
				busyTime -= idleDelta
			} else {
				busyTime = 0
			}
			*cpuPct = (float64(busyTime) / float64(totalSystem)) * 100.0
			if *cpuPct > 100.0 {
				*cpuPct = 100.0
			}
			*cpuMillis = int64(*cpuPct * float64(numCPU) * 10.0) // 100% of 1 core = 1000m
		}
	} else {
		hasPrevCPUTime = true
		*cpuPct = 2.5 // warm baseline
		*cpuMillis = int64(25 * numCPU)
	}

	lastIdleTime = curIdle
	lastKernelTime = curKernel
	lastUserTime = curUser
}
