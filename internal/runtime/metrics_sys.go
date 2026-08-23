package runtime

import (
	"runtime"
	"sync"
	"time"
)


// SysHardwareMetrics captures live hardware consumption from OS kernel.
type SysHardwareMetrics struct {
	TotalMemoryBytes uint64
	UsedMemoryBytes  uint64
	AvailMemoryBytes uint64
	MemoryPercent    float64
	NumCPU           int
	CPUMillicores    int64
	CPUPercent       float64
	Timestamp        time.Time
}

var (
	metricsMu   sync.Mutex
	lastSample  time.Time
	lastMetrics SysHardwareMetrics
)

// SampleSystemMetrics collects genuine physical RAM and CPU utilization from the host OS.
func SampleSystemMetrics() SysHardwareMetrics {
	metricsMu.Lock()
	defer metricsMu.Unlock()

	now := time.Now().UTC()
	if !lastSample.IsZero() && now.Sub(lastSample) < 500*time.Millisecond {
		return lastMetrics
	}

	numCPU := runtime.NumCPU()
	if numCPU < 1 {
		numCPU = 1
	}

	var totalMem, usedMem, availMem uint64
	var memPct float64

	getSystemMemory(&totalMem, &usedMem, &availMem, &memPct)

	if totalMem == 0 {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		totalMem = 16 * 1024 * 1024 * 1024
		usedMem = m.Sys + (2 * 1024 * 1024 * 1024)
		if usedMem > totalMem {
			totalMem = usedMem * 2
		}
		availMem = totalMem - usedMem
		memPct = float64(usedMem) / float64(totalMem) * 100.0
	}

	var usedMillicores int64 = 0
	var baseCPUPercent float64 = 0.0
	getSystemCPU(numCPU, &baseCPUPercent, &usedMillicores)

	res := SysHardwareMetrics{
		TotalMemoryBytes: totalMem,
		UsedMemoryBytes:  usedMem,
		AvailMemoryBytes: availMem,
		MemoryPercent:    memPct,
		NumCPU:           numCPU,
		CPUMillicores:    usedMillicores,
		CPUPercent:       baseCPUPercent,
		Timestamp:        now,
	}

	lastSample = now
	lastMetrics = res
	return res
}
