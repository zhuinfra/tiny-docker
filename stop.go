package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strconv"
	"syscall"

	"tiny-docker/container"
)

func stopContainer(containerId string) {
	// 1. 根据容器Id查询容器信息
	containerInfo, err := getInfoByContainerId(containerId)
	if err != nil {
		slog.Error("Get container info error", "containerID", containerId, "err", err)
		return
	}
	pidInt, err := strconv.Atoi(containerInfo.Pid)
	if err != nil {
		slog.Error("Conver pid from string to int error", "err", err)
		return
	}
	// 2.发送SIGTERM信号
	if err = syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		slog.Error("Stop container error", "containerId", containerId, "err", err)
		return
	}
	// 3.修改容器信息，将容器置为STOP状态，并清空PID
	containerInfo.Status = container.STOP
	containerInfo.Pid = " "
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		slog.Error("Json marshal error", "containerId", containerId, "err", err)
		return
	}
	// 4.重新写回存储容器信息的文件
	dirPath := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	configFilePath := path.Join(dirPath, container.ConfigName)
	if err = os.WriteFile(configFilePath, newContentBytes, 0622); err != nil {
		slog.Error("Write file error", "configFilePath", configFilePath, "err", err)
	}
	slog.Info("Stop container success", "containerId", containerId)
}

func getInfoByContainerId(containerId string) (*container.ContainerInfo, error) {
	dirPath := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	configFilePath := path.Join(dirPath, container.ConfigName)
	contentBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}
	var containerInfo container.ContainerInfo
	if err = json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return nil, err
	}
	return &containerInfo, nil
}

func removeContainer(containerId string, force bool) {
	containerInfo, err := getInfoByContainerId(containerId)
	if err != nil {
		slog.Error("Get container info error", "containerId", containerId, "err", err)
		return
	}

	switch containerInfo.Status {
	case container.STOP:
		// 1. 删除容器信息
		if err = container.DeleteContainerInfo(containerId); err != nil {
			slog.Error("Remove container config failed", "containerId", containerId, "err", err)
			return
		}
		// 2. 删除容器工作空间
		container.DeleteWorkSpace(containerId, containerInfo.Volume)
	case container.RUNNING:
		if !force {
			slog.Error("Couldn't remove running container, Stop the container before attempting removal or force remove")
			return
		}
		stopContainer(containerId)
		removeContainer(containerId, force)
		return
	default:
		slog.Error("Couldn't remove container,invalid status", "status", containerInfo.Status)
		return
	}
	slog.Info("Remove container success", "containerId", containerId)
}
