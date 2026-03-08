package network

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/vishvananda/netlink"
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
	Delete(network Network) error
	Connect(network *Network, endpoint *Endpoint) error
	Disconnect(network *Network, endpoint *Endpoint) error
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
		// 如果是目录则跳过
		if info.IsDir() {
			return nil
		}
		//  加载文件名作为网络名
		_, netName := path.Split(netPath)
		net := &Network{
			Name: netName,
		}
		// 调用前面介绍的 Network.load 方法加载网络的配置信息
		if err = net.load(netPath); err != nil {
			slog.Error("error load network", "err", err)
		}
		// 将网络的配置信息加入到 networks 字典中
		networks[netName] = net
		return nil
	})
	return networks, err
}

func Init() error {
	// 1.注册驱动
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver

	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(defaultNetworkPath, 0644)
		} else {
			return err
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

	return nil
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
	if err := drivers[nw.Driver].Delete(*nw); err != nil {
		return fmt.Errorf("Error Remove Network DriverError: %s", err)
	}

	// 3. 删除网络信息
	return nw.remove(defaultNetworkPath)
}
