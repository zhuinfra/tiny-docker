package network

import (
	"net"
	"testing"
)

func TestAllocate(t *testing.T) {
	_, ipnet, _ := net.ParseCIDR("192.168.1.0/24")
	ip, _ := ipAllocator.Allocate(ipnet)
	t.Logf("allocate ip: %v", ip)
}

func TestRelease(t *testing.T) {
	_, ipnet, _ := net.ParseCIDR("192.168.1.0/24")
	ip := net.ParseIP("192.168.1.1")
	err := ipAllocator.Release(ipnet, &ip)
	if err != nil {
		t.Fatal(err)
	}
}
