//go:build !linux

package tcr

import "fmt"

// RunContainerInit is a no-op on non-Linux platforms.
// On Linux, this sets up namespaces, mounts /proc /sys /dev, chroots, and exec's the container process.
func RunContainerInit() error {
	return fmt.Errorf("TCR native container init is only supported on Linux; on this OS (%s) use Docker Desktop, Podman Desktop, or Rancher Desktop", platformNote())
}

// RunContainerExec is a no-op on non-Linux platforms.
func RunContainerExec() error {
	return fmt.Errorf("TCR native exec is only supported on Linux")
}
