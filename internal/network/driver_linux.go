//go:build linux

package network

import (
	"os/exec"

	"go.uber.org/zap"
)

func (d *Driver) setupHostBridgePlatform() error {
	d.log.Info("configuring Linux host bridge interface and iptables forwarding",
		zap.String("bridge", d.cfg.BridgeName),
		zap.String("podCIDR", d.cfg.PodCIDR),
	)

	// Create bridge if not exists
	_ = exec.Command("ip", "link", "add", "name", d.cfg.BridgeName, "type", "bridge").Run()
	_ = exec.Command("ip", "addr", "add", "10.244.0.1/16", "dev", d.cfg.BridgeName).Run()
	_ = exec.Command("ip", "link", "set", "dev", d.cfg.BridgeName, "up").Run()

	// Enable IP forwarding
	_ = exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1").Run()

	// Enable iptables masquerade for outbound traffic
	_ = exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", d.cfg.PodCIDR, "!", "-o", d.cfg.BridgeName, "-j", "MASQUERADE").Run()
	_ = exec.Command("iptables", "-A", "FORWARD", "-i", d.cfg.BridgeName, "-j", "ACCEPT").Run()
	_ = exec.Command("iptables", "-A", "FORWARD", "-o", d.cfg.BridgeName, "-j", "ACCEPT").Run()

	return nil
}

func (d *Driver) teardownHostBridgePlatform() {
	_ = exec.Command("ip", "link", "set", "dev", d.cfg.BridgeName, "down").Run()
	_ = exec.Command("ip", "link", "delete", "dev", d.cfg.BridgeName).Run()
	_ = exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING", "-s", d.cfg.PodCIDR, "!", "-o", d.cfg.BridgeName, "-j", "MASQUERADE").Run()
}
