package main

import (
	"log/slog"
	"os"
	"strings"
	"tiny-docker/cgroups"
	"tiny-docker/cgroups/subsystems"
	"tiny-docker/container"
)

// 创建容器, 设置namespace和cgroup
func Run(tty bool, image string, comArray []string, res *subsystems.ResourceConfig, volume string, containerName string) {
	containerID := container.GenerateId()
	parent, writePipe, err := container.NewParentProcess(tty, volume, containerID, image)
	if parent == nil || err != nil {
		slog.Error("parent process create failed")
		return
	}
	if err := parent.Start(); err != nil {
		slog.Error("container startup failed", "error", err)
	}
	slog.Info("container created", "pid", parent.Process.Pid)

	container.RecordContainerInfo(containerID, parent.Process.Pid, comArray, containerName, volume)

	// 设置cgroup
	cgroupsManager := cgroups.NewCgroupManager("tiny-docker-cgroup")
	defer cgroupsManager.Destory()
	cgroupsManager.Set(res)
	cgroupsManager.Apply(parent.Process.Pid)

	// 通过管道向init进程发送容器命令
	sendInitCommand(comArray, writePipe)

	if tty {
		parent.Wait()
		// container.DeleteWorkSpace("containerID", volume)
	}

}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	writePipe.WriteString(command)
	writePipe.Close()
}
