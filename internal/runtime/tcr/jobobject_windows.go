//go:build windows

package tcr

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	kernel32                     = syscall.NewLazyDLL("kernel32.dll")
	procCreateJobObjectW         = kernel32.NewProc("CreateJobObjectW")
	procSetInformationJobObject  = kernel32.NewProc("SetInformationJobObject")
	procAssignProcessToJobObject = kernel32.NewProc("AssignProcessToJobObject")
	procCloseHandle              = kernel32.NewProc("CloseHandle")
)

const (
	JobObjectBasicLimitInformation               = 2
	JobObjectBasicUIRestrictions                 = 4
	JobObjectSecurityLimitInformation            = 5
	JobObjectEndOfJobTimeInformation             = 6
	JobObjectAssociateCompletionPortInformation  = 7
	JobObjectBasicAndIoAccountingInformation     = 8
	JobObjectExtendedLimitInformation            = 9
	JobObjectGroupInformation                    = 11
	JobObjectCpuRateControlInformation           = 15

	JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE          = 0x00002000
	JOB_OBJECT_LIMIT_PROCESS_MEMORY             = 0x00000100
	JOB_OBJECT_LIMIT_JOB_MEMORY                 = 0x00000200
	JOB_OBJECT_LIMIT_ACTIVE_PROCESS             = 0x00000008
	JOB_OBJECT_CPU_RATE_CONTROL_ENABLE          = 0x00000001
	JOB_OBJECT_CPU_RATE_CONTROL_HARD_CAP        = 0x00000004

	PROCESS_SET_QUOTA = 0x0100
	PROCESS_TERMINATE = 0x0001
)

type IO_COUNTERS struct {
	ReadOperationCount  uint64
	WriteOperationCount uint64
	OtherOperationCount uint64
	ReadTransferCount   uint64
	WriteTransferCount  uint64
	OtherTransferCount  uint64
}

type JOBOBJECT_BASIC_LIMIT_INFORMATION struct {
	PerProcessUserTimeLimit int64
	PerJobUserTimeLimit     int64
	LimitFlags              uint32
	MinimumWorkingSetSize   uintptr
	MaximumWorkingSetSize   uintptr
	ActiveProcessLimit      uint32
	Affinity                uintptr
	PriorityClass           uint32
	SchedulingClass         uint32
}

type JOBOBJECT_EXTENDED_LIMIT_INFORMATION struct {
	BasicLimitInformation JOBOBJECT_BASIC_LIMIT_INFORMATION
	IoInfo                IO_COUNTERS
	ProcessMemoryLimit    uintptr
	JobMemoryLimit        uintptr
	PeakProcessMemoryUsed uintptr
	PeakJobMemoryUsed     uintptr
}

type JOBOBJECT_CPU_RATE_CONTROL_INFORMATION struct {
	ControlFlags uint32
	Value        uint32 // CPU rate in 1/10000 of 100%
}

// WindowsJobObject manages Microsoft Windows Kernel Job Object resource sandboxing.
type WindowsJobObject struct {
	handle syscall.Handle
	id     string
}

// CreateWindowsJobObject creates a new Windows Job Object with KILL_ON_JOB_CLOSE flag.
func CreateWindowsJobObject(containerID string) (*WindowsJobObject, error) {
	name, _ := syscall.UTF16PtrFromString(fmt.Sprintf("tarak_job_%s", containerID))
	h, _, err := procCreateJobObjectW.Call(0, uintptr(unsafe.Pointer(name)))
	if h == 0 {
		return nil, fmt.Errorf("create job object: %w", err)
	}

	job := &WindowsJobObject{handle: syscall.Handle(h), id: containerID}

	// Always ensure processes are terminated when the job handle closes
	var info JOBOBJECT_EXTENDED_LIMIT_INFORMATION
	info.BasicLimitInformation.LimitFlags = JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	_, _, _ = procSetInformationJobObject.Call(
		uintptr(h),
		JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uintptr(unsafe.Sizeof(info)),
	)

	return job, nil
}

// SetMemoryLimit sets the maximum physical RAM limit for all processes inside the job.
func (j *WindowsJobObject) SetMemoryLimit(memoryBytes int64) error {
	if memoryBytes <= 0 {
		return nil
	}

	var info JOBOBJECT_EXTENDED_LIMIT_INFORMATION
	info.BasicLimitInformation.LimitFlags = JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE | JOB_OBJECT_LIMIT_JOB_MEMORY
	info.JobMemoryLimit = uintptr(memoryBytes)

	r, _, err := procSetInformationJobObject.Call(
		uintptr(j.handle),
		JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uintptr(unsafe.Sizeof(info)),
	)
	if r == 0 {
		return fmt.Errorf("set job memory limit: %w", err)
	}
	return nil
}

// SetCPULimit sets the hard CPU rate cap (e.g. 5000 = 50% CPU).
func (j *WindowsJobObject) SetCPULimit(cpuPercent float64) error {
	if cpuPercent <= 0 || cpuPercent > 100 {
		return nil
	}

	var cpuInfo JOBOBJECT_CPU_RATE_CONTROL_INFORMATION
	cpuInfo.ControlFlags = JOB_OBJECT_CPU_RATE_CONTROL_ENABLE | JOB_OBJECT_CPU_RATE_CONTROL_HARD_CAP
	cpuInfo.Value = uint32(cpuPercent * 100) // 10000 = 100%

	r, _, err := procSetInformationJobObject.Call(
		uintptr(j.handle),
		JobObjectCpuRateControlInformation,
		uintptr(unsafe.Pointer(&cpuInfo)),
		uintptr(unsafe.Sizeof(cpuInfo)),
	)
	if r == 0 {
		return fmt.Errorf("set job cpu limit: %w", err)
	}
	return nil
}

// AssignProcess binds an OS process PID into this Windows Job Object.
func (j *WindowsJobObject) AssignProcess(pid int) error {
	hProc, err := syscall.OpenProcess(PROCESS_SET_QUOTA|PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		return fmt.Errorf("open process %d: %w", pid, err)
	}
	defer syscall.CloseHandle(hProc)

	r, _, err := procAssignProcessToJobObject.Call(uintptr(j.handle), uintptr(hProc))
	if r == 0 {
		return fmt.Errorf("assign process to job: %w", err)
	}
	return nil
}

// Close terminates and cleans up the Windows Job Object.
func (j *WindowsJobObject) Close() {
	if j.handle != 0 {
		_, _, _ = procCloseHandle.Call(uintptr(j.handle))
		j.handle = 0
	}
}
