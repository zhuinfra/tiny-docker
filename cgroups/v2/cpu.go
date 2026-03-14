package v2

import (
	"tiny-docker/cgroups"
)

type CpuSubSystem struct {
}

func (s *CpuSubSystem) Set(cgroupPath string, res *cgroups.ResourceConfig) error {
	return nil
}

func (s *CpuSubSystem) Remove(cgroupPath string) error {
	return nil
}

func (s *CpuSubSystem) Apply(cgroupPath string, pid int) error {
	return nil
}

func (s *CpuSubSystem) Name() string {
	return "cpu"
}
