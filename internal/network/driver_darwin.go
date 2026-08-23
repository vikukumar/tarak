//go:build darwin

package network

import (
	"os/exec"

	"go.uber.org/zap"
)

func (d *Driver) setupHostBridgePlatform() error {
	d.log.Info("configuring macOS loopback aliases for cluster network",
		zap.String("podCIDR", d.cfg.PodCIDR),
	)

	// Add loopback alias
	_ = exec.Command("ifconfig", "lo0", "alias", "10.244.0.1", "255.255.0.0").Run()
	_ = exec.Command("route", "add", "-net", "10.244.0.0/16", "10.244.0.1").Run()
	return nil
}

func (d *Driver) teardownHostBridgePlatform() {
	_ = exec.Command("ifconfig", "lo0", "-alias", "10.244.0.1").Run()
	_ = exec.Command("route", "delete", "-net", "10.244.0.0/16").Run()
}
