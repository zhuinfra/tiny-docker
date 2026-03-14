package v1

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"strconv"
	"tiny-docker/cgroups"
)

type CpusetSubSystem struct {
}

func (s *CpusetSubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		if err := os.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
		return nil
	} else {
		return err
	}
}

func (s *CpusetSubSystem) Set(cgroupPath string, res *cgroups.ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		// 1.cpuset.cpus比较特殊, 多层级目录需要先初始化cpuset.cpus
		cgroupMountPoint := FindCgroupMountPoint(s.Name())
		parentPath := path.Join(cgroupMountPoint, cgroups.CgroupDir)
		if err := initCpuset(parentPath, cgroupMountPoint); err != nil {
			return err
		}

		// 2.写入cpuset.cpus配置
		if res.CpuSet != "" {
			if err := os.WriteFile(path.Join(subsysCgroupPath, "cpuset.cpus"), []byte(res.CpuSet), 0644); err != nil {
				slog.Error("set cpuset fail", "err", err)
				return err
			}
		}
		return nil
	} else {
		return err
	}
}

func (s *CpusetSubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return err
	}
}

func (s *CpusetSubSystem) Name() string {
	return "cpuset"
}

func initCpuset(current, parent string) error {
	// 读取 parent 的配置
	cpus, err := os.ReadFile(path.Join(parent, "cpuset.cpus"))
	if err != nil {
		return err
	}

	if err := os.WriteFile(
		path.Join(current, "cpuset.cpus"),
		cpus,
		0644,
	); err != nil {
		return err
	}

	return nil
}
