package network

import (
	"fmt"
	"log/slog"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/vishvananda/netlink"
)

type BridgeNetworkDriver struct {
}

func (d *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (d *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	ip, ipRange, _ := net.ParseCIDR(subnet)
	ipRange.IP = ip
	n := &Network{
		Name:    name,
		IpRange: ipRange,
		Driver:  d.Name(),
	}
	err := d.initBridge(n)
	if err != nil {
		slog.Error("init bridge error", "err", err)
	}
	return n, err
}

func (d *BridgeNetworkDriver) Delete(network *Network) error {
	// 1.清除路由规则
	err := deleteIPRoute(network.Name, network.IpRange.String())
	if err != nil {
		slog.Error("clean route rule error", "err", err)
		return err
	}
	// 2.清除 iptables 规则
	err = deleteIPTables(network.Name, network.IpRange)
	if err != nil {
		slog.Error("clean snat iptables rule error", "err", err)
		return err
	}
	// 3.删除网桥
	err = d.deleteBridge(network)
	if err != nil {
		slog.Error("delete bridge error", "err", err)
		return err
	}
	return nil
}

// 配置veth-pair, 其中一端挂载到bridge
func (d *BridgeNetworkDriver) Connect(networkName string, endpoint *Endpoint) error {
	bridgeName := networkName
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	la := netlink.NewLinkAttrs()
	la.Name = endpoint.ID[:5]
	la.MasterIndex = br.Attrs().Index
	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + endpoint.ID[:5],
	}
	if err = netlink.LinkAdd(&endpoint.Device); err != nil {
		return fmt.Errorf("error Add Endpoint Device: %v", err)
	}
	if err = netlink.LinkSetUp(&endpoint.Device); err != nil {
		return fmt.Errorf("error Add Endpoint Device: %v", err)
	}
	return nil
}

func (d *BridgeNetworkDriver) Disconnect(endpointID string) error {
	vethNme := endpointID[:5]
	veth, err := netlink.LinkByName(vethNme)
	if err != nil {
		return err
	}
	err = netlink.LinkSetNoMaster(veth)
	if err != nil {
		return err
	}
	err = netlink.LinkDel(veth)
	if err != nil {
		return err
	}
	veth2Name := "cif-" + vethNme
	veth2, err := netlink.LinkByName(veth2Name)
	if err != nil {
		return err
	}
	err = netlink.LinkDel(veth2)
	if err != nil {
		return err
	}

	return nil
}
func (d *BridgeNetworkDriver) deleteBridge(n *Network) error {
	bridgeName := n.Name

	// get the link
	l, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf("Getting link with name %s failed: %v", bridgeName, err)
	}

	// delete the link
	if err := netlink.LinkDel(l); err != nil {
		return fmt.Errorf("Failed to remove bridge interface %s delete: %v", bridgeName, err)
	}

	return nil
}

func (d *BridgeNetworkDriver) initBridge(n *Network) error {
	// 1. 创建 bridge
	bridgeName := n.Name
	if err := createBridgeInterface(bridgeName); err != nil {
		return err
	}

	gatewayIP := *n.IpRange
	gatewayIP.IP = n.IpRange.IP

	if err := setInterfaceIP(bridgeName, gatewayIP.String()); err != nil {
		return err
	}

	if err := setInterfaceUP(bridgeName); err != nil {
		return err
	}

	// Setup iptables
	if err := setupIPTables(bridgeName, n.IpRange); err != nil {
		return err
	}

	return nil
}

func createBridgeInterface(bridgeName string) error {
	// 1.检查是否已经存在
	_, err := net.InterfaceByName(bridgeName)
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return nil
	}

	// 2.初始化netlink的Link基础对象
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName

	// 3.初始化netlink的Bridge对象
	br := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkAdd(br); err != nil {
		return err
	}
	return nil
}

func setInterfaceIP(name string, rawIP string) error {
	// 解决新创建的网络接口，内核还未同步就绪导致的临时查询失败；
	retries := 2
	var iface netlink.Link
	var err error
	for i := 0; i < retries; i++ {
		iface, err = netlink.LinkByName(name)
		if err == nil {
			break
		}
		slog.Debug("error retrieving new bridge netlink link ... retrying")
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("Abandoning retrieving the new bridge link from netlink, Run [ ip link ] to troubleshoot the error: %v", err)
	}
	ipNet, err := netlink.ParseIPNet(rawIP)
	if err != nil {
		return err
	}
	addr := &netlink.Addr{IPNet: ipNet, Peer: ipNet, Label: "", Flags: 0, Scope: 0, Broadcast: nil}
	return netlink.AddrAdd(iface, addr)
}

func setInterfaceUP(interfaceName string) error {
	iface, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return fmt.Errorf("Error retrieving a link named [ %s ]: %v", iface.Attrs().Name, err)
	}

	if err := netlink.LinkSetUp(iface); err != nil {
		return fmt.Errorf("Error enabling interface for %s: %v", interfaceName, err)
	}
	return nil
}

// 删除路由，ip addr del xxx命令
func deleteIPRoute(name string, rawIP string) error {
	retries := 2
	var iface netlink.Link
	var err error
	for i := 0; i < retries; i++ {
		// 通过LinkByName方法找到需要设置的网络接口
		iface, err = netlink.LinkByName(name)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return err
	}
	// 查询对应设备的路由并全部删除
	list, err := netlink.RouteList(iface, netlink.FAMILY_V4)
	if err != nil {
		return err
	}
	for _, route := range list {
		if route.Dst.String() == rawIP { // 根据子网进行匹配
			err = netlink.RouteDel(&route)
			if err != nil {
				slog.Error("delete route error", "err", err)
				continue
			}
		}
	}
	return nil
}

func setupIPTables(bridgeName string, subnet *net.IPNet) error {
	return configIPTables(bridgeName, subnet, false)
}

func deleteIPTables(bridgeName string, subnet *net.IPNet) error {
	return configIPTables(bridgeName, subnet, true)
}

func configIPTables(bridgeName string, subnet *net.IPNet, isDelete bool) error {
	action := "-A"
	if isDelete {
		action = "-D"
	}
	// 拼接命令
	iptablesCmd := fmt.Sprintf("-t nat %s POSTROUTING -s %s ! -o %s -j MASQUERADE", action, subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	slog.Info("配置 SNAT", "cmd", cmd.String())
	// 执行该命令
	output, err := cmd.Output()
	if err != nil {
		slog.Error("iptables Output", "output", output)
	}
	return err
}
