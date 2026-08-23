//go:build windows

package network

import (
	"os/exec"

	"go.uber.org/zap"
)

func (d *Driver) setupHostBridgePlatform() error {
	// 1. On Windows, ensure route for Pod CIDR 10.244.0.0/16 points to loopback/local interface so host can connect directly
	// e.g. route add 10.244.0.0 MASK 255.255.0.0 127.0.0.1 metric 1
	d.log.Info("configuring Windows host routing table for direct cluster access",
		zap.String("podCIDR", d.cfg.PodCIDR),
		zap.String("serviceCIDR", d.cfg.ServiceCIDR),
	)

	// Attempt adding host routes (fails gracefully if already present or running non-admin)
	cmd := exec.Command("route", "add", "10.244.0.0", "mask", "255.255.0.0", "127.0.0.1", "metric", "1")
	_ = cmd.Run()

	cmdSvc := exec.Command("route", "add", "10.96.0.0", "mask", "255.240.0.0", "127.0.0.1", "metric", "1")
	_ = cmdSvc.Run()

	return nil
}

func (d *Driver) teardownHostBridgePlatform() {
	_ = exec.Command("route", "delete", "10.244.0.0").Run()
	_ = exec.Command("route", "delete", "10.96.0.0").Run()
}
