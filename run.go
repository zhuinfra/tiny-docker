package main

import (
	"log/slog"
	"os"
	"strings"
	"tiny-docker/cgroups"
	"tiny-docker/cgroups/subsystems"
	"tiny-docker/container"
)

func Run(tty bool, comArray []string, res *subsystems.ResourceConfig) {
	parent, writePipe := container.NewParentProcess(tty)
	if parent == nil {
		slog.Error("parent process create failed")
		return
	}
	if err := parent.Start(); err != nil {
		slog.Error("container startup failed", "error", err)
	}
	cgroupsManager := cgroups.NewCgroupManager("tiny-docker-cgroup")
	defer cgroupsManager.Destory()
	slog.Info("cgroup", "pid", parent.Process.Pid)
	cgroupsManager.Set(res)
	cgroupsManager.Apply(parent.Process.Pid)

	sendInitCommand(comArray, writePipe)

	parent.Wait()
	os.Exit(1)
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	slog.Info("command", "command", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
