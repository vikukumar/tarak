//go:build linux

package network

import (
	"fmt"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

// IPTablesManager configures Linux kernel iptables rules for high-speed ClusterIP and NodePort routing.
type IPTablesManager struct {
	log *zap.Logger
}

// NewIPTablesManager initializes the iptables manager.
func NewIPTablesManager(log *zap.Logger) *IPTablesManager {
	if log == nil {
		log = zap.NewNop()
	}
	return &IPTablesManager{log: log.Named("iptables")}
}

// InitChains creates the TARAK-SERVICES and TARAK-POSTROUTING custom iptables chains.
func (m *IPTablesManager) InitChains() error {
	commands := [][]string{
		{"-t", "nat", "-N", "TARAK-SERVICES"},
		{"-t", "nat", "-N", "TARAK-POSTROUTING"},
		{"-t", "nat", "-C", "PREROUTING", "-j", "TARAK-SERVICES"},
		{"-t", "nat", "-A", "PREROUTING", "-j", "TARAK-SERVICES"},
		{"-t", "nat", "-C", "OUTPUT", "-j", "TARAK-SERVICES"},
		{"-t", "nat", "-A", "OUTPUT", "-j", "TARAK-SERVICES"},
		{"-t", "nat", "-C", "POSTROUTING", "-j", "TARAK-POSTROUTING"},
		{"-t", "nat", "-A", "POSTROUTING", "-j", "TARAK-POSTROUTING"},
	}

	for _, args := range commands {
		cmd := exec.Command("iptables", args...)
		_ = cmd.Run() // Errors are non-fatal if chains/rules already exist
	}

	m.log.Info("initialized Linux kernel iptables TARAK chains")
	return nil
}

// SyncServiceNAT creates DNAT rules redirecting ClusterIP traffic to backend pod endpoints.
func (m *IPTablesManager) SyncServiceNAT(clusterIP string, port int, backends []string) {
	if len(backends) == 0 {
		return
	}

	// Direct DNAT to primary backend
	primaryBackend := backends[0]
	parts := strings.Split(primaryBackend, ":")
	if len(parts) != 2 {
		return
	}
	dstIP, dstPort := parts[0], parts[1]

	rule := []string{
		"-t", "nat", "-A", "TARAK-SERVICES",
		"-d", clusterIP, "-p", "tcp", "--dport", fmt.Sprintf("%d", port),
		"-j", "DNAT", "--to-destination", fmt.Sprintf("%s:%s", dstIP, dstPort),
	}

	cmd := exec.Command("iptables", rule...)
	if out, err := cmd.CombinedOutput(); err != nil {
		m.log.Debug("iptables rule sync note", zap.String("out", string(out)), zap.Error(err))
	}
}

// Cleanup flushes TARAK custom chains.
func (m *IPTablesManager) Cleanup() {
	_ = exec.Command("iptables", "-t", "nat", "-F", "TARAK-SERVICES").Run()
	_ = exec.Command("iptables", "-t", "nat", "-F", "TARAK-POSTROUTING").Run()
}
