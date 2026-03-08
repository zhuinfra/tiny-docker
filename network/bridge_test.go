package network

import (
	"net"
	"testing"
)

var testName = "testbr"

func TestCreate(t *testing.T) {
	d := BridgeNetworkDriver{}
	n, err := d.Create("192.168.0.1/24", testName)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("create network: %v", n.Name)
}

func TestDeleteBridge(t *testing.T) {
	d := BridgeNetworkDriver{}
	_, ipRange, _ := net.ParseCIDR("192.168.0.1/24")
	n := &Network{
		Name:    testName,
		IpRange: ipRange,
	}
	err := d.deleteBridge(n)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("delete network :%v", testName)
}
