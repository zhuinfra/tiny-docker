package network

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"

	"tiny-docker/container"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

var (
	defaultNetworkPath = "/var/run/tiny-docker/network/network/"
	drivers            = map[string]NetworkDriver{}
	networks           = map[string]*Network{}
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
	Delete(network *Network) error
	Connect(networkName string, endpoint *Endpoint) error
	Disconnect(endpointID string) error
}

// 持久化保存网络信息
func (nw *Network) dump(dumpPath string) error {
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(dumpPath, 0644)
		} else {
			return err
		}
	}

	// 每个network单独存放
	nwPath := path.Join(dumpPath, nw.Name)
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer nwFile.Close()

	nwJson, err := json.Marshal(nw)
	if err != nil {
		return err
	}

	_, err = nwFile.Write(nwJson)
	if err != nil {
		return err
	}
	return nil
}

func (nw *Network) remove(dumpPath string) error {
	if _, err := os.Stat(path.Join(dumpPath, nw.Name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else {
		return os.Remove(path.Join(dumpPath, nw.Name))
	}
}

func (nw *Network) load(dumpPath string) error {
	nwConfigFile, err := os.Open(dumpPath)
	if err != nil {
		return err
	}
	defer nwConfigFile.Close()
	nwJson := make([]byte, 2000)
	n, err := nwConfigFile.Read(nwJson)
	if err != nil {
		return err
	}

	err = json.Unmarshal(nwJson[:n], nw)
	if err != nil {
		return err
	}
	return nil
}

func loadNetwork() (map[string]*Network, error) {
	networks := map[string]*Network{}

	// 检查网络配置目录中的所有文件,并执行第二个参数中的函数指针去处理目录下的每一个文件
	err := filepath.Walk(defaultNetworkPath, func(netPath string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			return nil
		}
		// 加载文件名作为网络名
		_, netName := path.Split(netPath)
		net := &Network{
			Name: netName,
		}
		// 调用前面介绍的 Network.load 方法加载网络的配置信息
		if err = net.load(netPath); err != nil {
			slog.Error("error load network", "err", err)
			return err
		}
		// 将网络的配置信息加入到 networks 字典中
		networks[netName] = net
		return nil
	})
	return networks, err
}

func init() {
	// 1.注册驱动
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver

	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(defaultNetworkPath, 0644)
		} else {
			return
		}
	}

	// 2.加载已存在的网络
	filepath.Walk(defaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		if strings.HasSuffix(nwPath, "/") {
			return nil
		}
		_, nwName := path.Split(nwPath)
		nw := &Network{
			Name: nwName,
		}

		if err := nw.load(nwPath); err != nil {
			slog.Error("error load network", "err", err)
		}

		networks[nwName] = nw
		return nil
	})
}

func CreateNetwork(driver, subnet, name string) error {
	_, cidr, _ := net.ParseCIDR(subnet)
	ip, err := ipAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = ip

	nw, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}
	networks[name] = nw
	return nw.dump(defaultNetworkPath)
}

func ListNetwork() {
	networks, err := loadNetwork()
	if err != nil {
		slog.Error("load network from file failed", "err", err)
		return
	}
	// 通过tabwriter库把信息打印到屏幕上
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "NAME\tIpRange\tDriver\n")
	for _, net := range networks {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			net.Name,
			net.IpRange.String(),
			net.Driver,
		)
	}
	if err = w.Flush(); err != nil {
		slog.Error("Flush error", "err", err)
		return
	}
}

func DeleteNetwork(networkName string) error {
	networks, err := loadNetwork()
	if err != nil {
		return err
	}
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("No Such Network: %s", networkName)
	}

	// 1. 释放网关ip
	if err := ipAllocator.Release(nw.IpRange, &nw.IpRange.IP); err != nil {
		return fmt.Errorf("Error Remove Network gateway ip: %s", err)
	}

	// 2. 删除网络驱动
	if err := drivers[nw.Driver].Delete(nw); err != nil {
		return fmt.Errorf("Error Remove Network DriverError: %s", err)
	}

	// 3. 删除网络信息
	return nw.remove(defaultNetworkPath)
}

// 创建网络端点并加入指定的网络(驱动类型相关)
func Connect(networkName string, info *container.ContainerInfo) (net.IP, error) {
	// 1.加载记录的网络信息
	networks, err := loadNetwork()
	if err != nil {
		return nil, err
	}
	network, ok := networks[networkName]
	if !ok {
		return nil, fmt.Errorf("no Such Network: %s", networkName)
	}

	// 2.分配容器IP地址
	ip, err := ipAllocator.Allocate(network.IpRange)
	if err != nil {
		return ip, err
	}
	// 3.创建网络端点
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", info.Id, networkName),
		IPAddress:   ip,
		Network:     network,
		PortMapping: info.PortMapping,
	}

	// 4.调用网络驱动挂载和配置网络端点
	if err = drivers[network.Driver].Connect(network.Name, ep); err != nil {
		return ip, err
	}

	// 5.到容器的namespace配置容器网络设备IP地址
	if err = configEndpointIpAddressAndRoute(ep, info); err != nil {
		return ip, err
	}
	// 6.配置端口映射信息
	return ip, addPortMapping(ep)
}

// Disconnect 将容器中指定网络中移除
func Disconnect(networkName string, info *container.ContainerInfo) error {
	networks, err := loadNetwork()
	if err != nil {
		return err
	}
	// 从networks字典中取到容器连接的网络的信息，networks字典中保存了当前己经创建的网络
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no Such Network: %s", networkName)
	}
	// veth 从 bridge 解绑并删除 veth-pair 设备对
	drivers[network.Driver].Disconnect(fmt.Sprintf("%s-%s", info.Id, networkName))

	// 清理端口映射添加的 iptables 规则
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", info.Id, networkName),
		IPAddress:   net.ParseIP(info.IP),
		Network:     network,
		PortMapping: info.PortMapping,
	}
	return deletePortMapping(ep)
}

func enterContainerNetNS(enLink *netlink.Link, info *container.ContainerInfo) func() {
	// /proc/[pid]/ns/net 打开这个文件的文件描述符就可以来操作Net Namespace
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", info.Pid), os.O_RDONLY, 0)
	if err != nil {
		slog.Error("error open container net namespace")
	}
	nsFD := f.Fd()

	// 调用runtime.LockOSThread()锁定当前程序执行的线程, 防止goroutine被调度到其他线程, 保证一直在所需的网络空间
	runtime.LockOSThread()

	// 将网络端点的一端veth移动到容器的Net Namespace
	if err = netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		slog.Error("error set link netns")
	}

	// 获取当前的网络namespace
	origns, err := netns.Get()
	if err != nil {
		slog.Error("error get current netns")
	}

	// 调用 netns.Set方法，将当前进程加入容器的Net Namespace
	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		slog.Error("error set netns")
	}

	// 在容器的网络空间中执行完容器配置之后调用此函数就可以将程序恢复到原生的Net Namespace
	return func() {
		netns.Set(origns)
		origns.Close()
		runtime.UnlockOSThread()
		f.Close()
	}
}

// 配置容器网络端点的地址和路由
func configEndpointIpAddressAndRoute(ep *Endpoint, info *container.ContainerInfo) error {
	// 根据名字找到对应Veth设备
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return err
	}

	// defer 只延迟函数调用，不延迟函数参数计算。
	// 进入容器的网络空间, 配置完成后恢复到原生网络空间
	defer enterContainerNetNS(&peerLink, info)()

	interfaceIP := *ep.Network.IpRange
	interfaceIP.IP = ep.IPAddress
	// 设置容器内Veth端点的IP
	if err = setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return fmt.Errorf("%v,%s", ep.Network, err)
	}
	// 启动容器内的Veth端点和环回接口
	if err = setInterfaceUP(ep.Device.PeerName); err != nil {
		return err
	}
	if err = setInterfaceUP("lo"); err != nil {
		return err
	}

	// 设置容器内默认路由，route add -net 0.0.0.0/0 gw (Bridge网桥地址) dev （容器内的Veth端点设备)
	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IpRange.IP,
		Dst:       cidr,
	}
	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return err
	}

	return nil
}
func addPortMapping(ep *Endpoint) error {
	return configPortMapping(ep, false)
}

func deletePortMapping(ep *Endpoint) error {
	return configPortMapping(ep, true)
}

// 通过iptables配置主机和容器的端口映射
func configPortMapping(ep *Endpoint, isDelete bool) error {
	action := "-A"
	if isDelete {
		action = "-D"
	}

	var err error
	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			slog.Error("port mapping format error")
			continue
		}
		// iptables -t nat -A PREROUTING ! -i testbridge -p tcp -m tcp --dport 8080 -j DNAT --to-destination 10.0.0.4:80
		iptablesCmd := fmt.Sprintf("-t nat %s PREROUTING ! -i %s -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			action, ep.Network.Name, portMapping[0], ep.IPAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		slog.Info("配置端口映射 DNAT", "cmd", cmd.String())
		output, err := cmd.Output()
		if err != nil {
			slog.Error("iptables Output", "output", output)
			continue
		}
	}
	return err
}
