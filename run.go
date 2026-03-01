package main

import (
	"log/slog"
	"os"
	"tiny-docker/cgroups"
	"tiny-docker/cgroups/subsystems"
	"tiny-docker/container"
)

func Run(tty bool, command string, res *subsystems.ResourceConfig) {
	parent := container.NewParentProcess(tty, command)
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
	parent.Wait()
	os.Exit(1)
}
