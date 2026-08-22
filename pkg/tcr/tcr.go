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

// ExecConfig specifies a command to run inside an existing container.
type ExecConfig = tcr.ExecConfig

// Process tracks a running container process.
type Process = tcr.Process

// RunContainerInit invokes the native container init sequence.
func RunContainerInit() error {
	return tcr.RunContainerInit()
}

// RunContainerExec executes a command inside a running container.
func RunContainerExec() error {
	return tcr.RunContainerExec()
}
