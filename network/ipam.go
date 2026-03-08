package network

import (
	"encoding/json"
	"log/slog"
	"net"
	"os"
	"path"
	"strings"
)

const ipamDefaultAllocatorPath = "/var/run/tiny-docker/network/ipam/subnet.json"

// 网段分配器
type IPAM struct {
	// 分配文件路径
	SubnetAllocatorPath string
	// key为网段, value为位图
	Subnets *map[string]string
}

var ipAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

// 加载网段分配信息
func (ipam *IPAM) load() error {
	// 存储文件如果不存在, 说明之前未分配, 无需加载
	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
	subnetConfigFile, err := os.Open(ipam.SubnetAllocatorPath)
	if err != nil {
		return err
	}
	defer subnetConfigFile.Close()
	subnetJson := make([]byte, 2048)
	n, err := subnetConfigFile.Read(subnetJson)
	if err != nil {
		return err
	}

	err = json.Unmarshal(subnetJson[:n], ipam.Subnets)
	if err != nil {
		return err
	}
	return nil
}

// 存储网段地址分配信息
func (ipam *IPAM) dump() error {
	// 创建存储文件目录
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(ipamConfigFileDir); os.IsNotExist(err) {
		if err := os.MkdirAll(ipamConfigFileDir, 0644); err != nil {
			return err
		}
	}

	subnetConfigFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer subnetConfigFile.Close()
	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return err
	}
	_, err = subnetConfigFile.Write(ipamConfigJson)
	return err
}

// 从指定网段分配一个未使用的ip
func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	ipam.Subnets = &map[string]string{}

	// 加载网段分配信息
	if err := ipam.load(); err != nil {
		slog.Error("load ipam config error")
	}

	// 标准化子网格式
	_, subnet, _ = net.ParseCIDR(subnet.String())
	one, size := subnet.Mask.Size()

	// 初始化子网的标记串(首次分配)
	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		// 生成 2^(size-one) 个 "0", 创建位图
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(size-one))
	}

	// 遍历位图, 寻找第一个未使用的ip
	for c := range (*ipam.Subnets)[subnet.String()] {
		// 找到第一个标记为 0 的位置（未分配）
		if (*ipam.Subnets)[subnet.String()][c] == '0' {
			// 1. 把该位置标记为 1（已分配）
			ipalloc := []byte((*ipam.Subnets)[subnet.String()])
			ipalloc[c] = '1'
			(*ipam.Subnets)[subnet.String()] = string(ipalloc)

			// 2. 计算该位置对应的实际 IP 地址
			ip = subnet.IP // 子网起始 IP（比如 192.168.1.0）
			// 按"字节"计算偏移量（把索引 c 转成 IP 字节）
			// 通过除法+取余同样能实现，但是位运算对计算机更友好
			for t := uint(4); t > 0; t -= 1 {
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}
			ip[3] += 1 // 最后一个字节+1（比如 192.168.1.0 → 192.168.1.1）

			break // 找到第一个就退出循环
		}
	}

	ipam.dump()
	return
}

// 释放指定网段下的指定ip
func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	ipam.Subnets = &map[string]string{}

	_, subnet, _ = net.ParseCIDR(subnet.String())

	err := ipam.load()
	if err != nil {
		return err
	}

	// 计算ip在网段位图中的索引
	c := 0
	releaseIP := ipaddr.To4()
	releaseIP[3] -= 1
	for t := uint(4); t > 0; t -= 1 {
		c += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}

	// 将分配的位图数组中索引位置的值置为０
	ipalloc := []byte((*ipam.Subnets)[subnet.String()])
	ipalloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipalloc)

	// 存储网段分配信息
	ipam.dump()
	return nil
}
