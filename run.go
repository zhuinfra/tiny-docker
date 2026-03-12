package main

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"tiny-docker/cgroups"
	types "tiny-docker/cgroups/types"
	"tiny-docker/container"
	"tiny-docker/network"
)

// 创建容器, 设置namespace和cgroup
func Run(tty bool,
	volume, net, containerName, image string,
	comArray, envSlice, portMapping []string,
	res *cgroups.ResourceConfig) {

	// 1.创建docker init 进程
	containerID := container.GenerateId()
	parent, writePipe, err := container.NewParentProcess(tty, volume, containerID, image, envSlice)
	if parent == nil || err != nil {
		slog.Error("parent process create failed")
		return
	}
	if err := parent.Start(); err != nil {
		slog.Error("container startup failed", "error", err)
	}
	slog.Info("container created", "pid", parent.Process.Pid)

	// 2.配置容器网络
	var containerIP string
	if net != "" {
		containerInfo := &container.ContainerInfo{
			Id:          containerID,
			Pid:         strconv.Itoa(parent.Process.Pid),
			Name:        containerName,
			PortMapping: portMapping,
		}
		ip, err := network.Connect(net, containerInfo)
		if err != nil {
			slog.Error("Error Connect Network", "err", err)
			return
		}
		containerIP = ip.String()
	}

	// 3.记录容器信息
	container.RecordContainerInfo(containerID, containerName, volume, net, containerIP, parent.Process.Pid, portMapping, comArray)

	// 4.设置cgroup
	cgroupsManager := types.NewCgroupManager("tiny-docker-cgroup")
	defer cgroupsManager.Destory()
	cgroupsManager.Set(res)
	cgroupsManager.Apply(parent.Process.Pid)

	// 通过管道向init进程发送容器命令
	sendInitCommand(comArray, writePipe)

	if tty {
		parent.Wait()
		container.DeleteContainerInfo(containerID)
		container.DeleteWorkSpace(containerID, volume)
	}

}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	writePipe.WriteString(command)
	writePipe.Close()
}
