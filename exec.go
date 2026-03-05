package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strings"

	"tiny-docker/container"
	_ "tiny-docker/nsenter" // 空白导入，仅执行包的 init 函数（和 Cgo 构造函数）
)

// 控制是否执行C代码里面的setns.
const (
	EnvExecPid = "mydocker_pid"
	EnvExecCmd = "mydocker_cmd"
)

func ExecContainer(containerId string, comArray []string) {
	// 根据传进来的容器名获取对应的PID
	pid, err := getPidByContainerId(containerId)
	if err != nil {
		slog.Error("Exec container getPidByContainerId error", "err", err)
		return
	}

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 把命令拼接成字符串，便于传递
	cmdStr := strings.Join(comArray, " ")
	slog.Info("attach container", "pid", pid, "command", cmdStr)
	_ = os.Setenv(EnvExecPid, pid)
	_ = os.Setenv(EnvExecCmd, cmdStr)
	// 把指定PID进程的环境变量传递给新启动的进程，实现通过exec命令也能查询到容器的环境变量
	containerEnvs := getEnvsByPid(pid)
	cmd.Env = append(os.Environ(), containerEnvs...)

	if err = cmd.Run(); err != nil {
		slog.Error("Exec container error", "err", err)
	}
}

func getPidByContainerId(containerId string) (string, error) {
	// 拼接出记录容器信息的文件路径
	dirPath := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	configFilePath := path.Join(dirPath, container.ConfigName)
	// 读取内容并解析
	contentBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}
	var containerInfo container.ContainerInfo
	if err = json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return "", err
	}
	return containerInfo.Pid, nil
}

// getEnvsByPid 读取指定PID进程的环境变量
func getEnvsByPid(pid string) []string {
	path := fmt.Sprintf("/proc/%s/environ", pid)
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		slog.Error("Read file error", "err", err)
		return nil
	}
	// env split by \u0000
	envs := strings.Split(string(contentBytes), "\u0000")
	return envs
}
