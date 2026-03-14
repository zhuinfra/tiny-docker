package v2

import (
	"tiny-docker/cgroups"
)

type MemorySubSystem struct {
}

func (s *MemorySubSystem) Set(cgroupPath string, res *cgroups.ResourceConfig) error {
	return nil
}

func (s *MemorySubSystem) Remove(cgroupPath string) error {
	return nil
}
func (s *MemorySubSystem) Apply(cgroupPath string, pid int) error {
	return nil
}
func (s *MemorySubSystem) Name() string {
	return "memory"
}
