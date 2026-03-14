package v2

import (
	"tiny-docker/cgroups"
)

type CpusetSubSystem struct {
}

func (s *CpusetSubSystem) Set(cgroupPath string, res *cgroups.ResourceConfig) error {
	return nil
}

func (s *CpusetSubSystem) Remove(cgroupPath string) error {
	return nil
}

func (s *CpusetSubSystem) Apply(cgroupPath string, pid int) error {
	return nil
}

func (s *CpusetSubSystem) Name() string {
	return "cpuset"
}
