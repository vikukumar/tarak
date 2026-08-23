//go:build !windows

package runtime

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

var (
	lastLinuxIdle  uint64
	lastLinuxTotal uint64
	hasLinuxPrev   bool
)

func getSystemMemory(total, used, avail *uint64, pct *float64) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return
	}
	defer file.Close()

	var memTotal, memAvail, memFree uint64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		val, _ := strconv.ParseUint(fields[1], 10, 64)
		valBytes := val * 1024 // /proc/meminfo is in kB

		switch fields[0] {
		case "MemTotal:":
			memTotal = valBytes
		case "MemAvailable:":
			memAvail = valBytes
		case "MemFree:":
			memFree = valBytes
		}
	}

	if memTotal > 0 {
		*total = memTotal
		if memAvail > 0 {
			*avail = memAvail
		} else {
			*avail = memFree
		}
		*used = *total - *avail
		*pct = float64(*used) / float64(*total) * 100.0
	}
}

func getSystemCPU(numCPU int, cpuPct *float64, cpuMillis *int64) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return
	}
	fields := strings.Fields(scanner.Text())
	if len(fields) < 5 || fields[0] != "cpu" {
		return
	}

	var user, nice, system, idle, iowait, irq, softirq, steal uint64
	user, _ = strconv.ParseUint(fields[1], 10, 64)
	nice, _ = strconv.ParseUint(fields[2], 10, 64)
	system, _ = strconv.ParseUint(fields[3], 10, 64)
	idle, _ = strconv.ParseUint(fields[4], 10, 64)
	if len(fields) > 5 {
		iowait, _ = strconv.ParseUint(fields[5], 10, 64)
	}
	if len(fields) > 6 {
		irq, _ = strconv.ParseUint(fields[6], 10, 64)
	}
	if len(fields) > 7 {
		softirq, _ = strconv.ParseUint(fields[7], 10, 64)
	}
	if len(fields) > 8 {
		steal, _ = strconv.ParseUint(fields[8], 10, 64)
	}

	total := user + nice + system + idle + iowait + irq + softirq + steal
	idleTotal := idle + iowait

	if hasLinuxPrev {
		totalDelta := total - lastLinuxTotal
		idleDelta := idleTotal - lastLinuxIdle
		if totalDelta > 0 {
			busy := totalDelta
			if busy > idleDelta {
				busy -= idleDelta
			} else {
				busy = 0
			}
			*cpuPct = (float64(busy) / float64(totalDelta)) * 100.0
			*cpuMillis = int64(*cpuPct * float64(numCPU) * 10.0)
		}
	} else {
		hasLinuxPrev = true
		*cpuPct = 2.0
		*cpuMillis = int64(20 * numCPU)
	}

	lastLinuxTotal = total
	lastLinuxIdle = idleTotal
}
