// Package tcr provides public access to the Tarak Container Runtime primitives.
package tcr

import (
	"github.com/vikukumar/tarak/internal/runtime/tcr"
)

// Re-export common TCR constants and execution helpers
const (
	InitArg = tcr.InitArg
	ExecArg = tcr.ExecArg
)

// ContainerConfig holds container launch configuration.
type ContainerConfig = tcr.ContainerConfig

// ProcessState represents a running container process.
type ProcessState = tcr.ProcessState

// RunContainerInit invokes the native container init sequence.
func RunContainerInit() error {
	return tcr.RunContainerInit()
}

// RunContainerExec executes a command inside a running container.
func RunContainerExec() error {
	return tcr.RunContainerExec()
}
