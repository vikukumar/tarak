package network

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestDriver_IPDetectionAndAllocation(t *testing.T) {
	log := zap.NewNop()
	driver := NewDriver(BridgeConfig{
		PodCIDR:     "10.244.0.0/16",
		ServiceCIDR: "10.96.0.0/12",
	}, log)

	if err := driver.Start(context.Background()); err != nil {
		t.Fatalf("failed to start driver: %v", err)
	}
	defer driver.Stop()

	info := driver.GetHostNetworkInfo()
	t.Logf("Host Network Info:\n- LAN IP: %s\n- Public IP: %s\n- All Host IPs: %v\n- Pod CIDR: %s\n- Service CIDR: %s\n- Bridge Active: %v",
		info.PrimaryLANIP, info.PublicIP, info.AllHostIPs, info.PodCIDR, info.ServiceCIDR, info.BridgeActive)

	if info.PrimaryLANIP == "" {
		t.Errorf("expected PrimaryLANIP to not be empty")
	}

	// Test deterministic Pod IP allocation
	ip1 := driver.AllocatePodIP(0, 0)
	if ip1 != "10.244.0.2" {
		t.Errorf("expected 10.244.0.2, got %s", ip1)
	}

	ip2 := driver.AllocatePodIP(0, 1)
	if ip2 != "10.244.0.3" {
		t.Errorf("expected 10.244.0.3, got %s", ip2)
	}

	// Route registration
	driver.RegisterPodRoute("10.244.0.2", "127.0.0.1:8080")
	driver.UnregisterPodRoute("10.244.0.2")
}
