//go:build !linux

package network

import "go.uber.org/zap"

type IPTablesManager struct{}

func NewIPTablesManager(log *zap.Logger) *IPTablesManager {
	return &IPTablesManager{}
}

func (m *IPTablesManager) InitChains() error {
	return nil
}

func (m *IPTablesManager) SyncServiceNAT(clusterIP string, port int, backends []string) {}

func (m *IPTablesManager) Cleanup() {}
