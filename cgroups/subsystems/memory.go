package subsystems

import (
	"fmt"
	"os"
)

type MemorySubSystem struct {
}

func (s *MemorySubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if err := os.WriteFile(subsysCgroupPath+"/memory.limit_in_bytes", []byte(res.MemoryLimit), 0644); err != nil {
			return fmt.Errorf("set cgroup memory fail %v", err)
		}
		return nil
	} else {
		return err
	}
}
func (s *MemorySubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		if err := os.WriteFile(subsysCgroupPath+"/tasks", []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
		return nil
	} else {
		return err
	}
}
func (s *MemorySubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return err
	}
}
func (s *MemorySubSystem) Name() string {
	return "memory"
}
