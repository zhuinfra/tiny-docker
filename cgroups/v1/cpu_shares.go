package v1

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"tiny-docker/cgroups"
)

type CpuSubSystem struct {
}

func (s *CpuSubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		if err := os.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
		return nil
	} else {
		return err
	}
}
func (s *CpuSubSystem) Set(cgroupPath string, res *cgroups.ResourceConfig) error {
	fmt.Println(res.CpuShare)
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.CpuShare != "" {
			if err := os.WriteFile(path.Join(subsysCgroupPath, "cpu.shares"), []byte(res.CpuShare), 0644); err != nil {
				return fmt.Errorf("set cgroup cpu share fail %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

func (s *CpuSubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return err
	}
}

func (s *CpuSubSystem) Name() string {
	return "cpu"
}
