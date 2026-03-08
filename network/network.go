package network

import (
	"net"

	"github.com/vishvananda/netlink"
)

// 划分网络, 同一网络容器可相互通信
type Network struct {
	Name    string
	IpRange *net.IPNet
	Driver  string
}

// 容器的端点信息, 连接容器与网络
type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"dev"`
	IPAddress   net.IP           `json:"ip"`
	MacAddress  net.HardwareAddr `json:"mac"`
	PortMapping []string
	Network     *Network
}

// 网络驱动接口, 不同的驱动对网络的创建、连接、销毁的策略不同
type NetworkDriver interface {
	Name() string
	Create(subnet string, name string) (*Network, error)
	Delete(network Network) error
	Connect(network *Network, endpoint *Endpoint) error
	Disconnect(network *Network, endpoint *Endpoint) error
}
