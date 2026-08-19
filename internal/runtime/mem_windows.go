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

func getSystemMemory(total, used, avail *uint64, pct *float64) {
	mod := syscall.NewLazyDLL("kernel32.dll")
	proc := mod.NewProc("GlobalMemoryStatusEx")
	var ms memoryStatusEx
	ms.cbSize = uint32(unsafe.Sizeof(ms))
	r1, _, _ := proc.Call(uintptr(unsafe.Pointer(&ms)))
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
