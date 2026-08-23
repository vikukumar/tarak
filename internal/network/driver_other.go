//go:build !windows && !linux && !darwin

package network

func (d *Driver) setupHostBridgePlatform() error {
	return nil
}

func (d *Driver) teardownHostBridgePlatform() {
}
