package types

import (
	"tiny-docker/cgroups"
	v1 "tiny-docker/cgroups/v1"
)

type CgroupManagerV1 struct {
	Path     string
	Resource *cgroups.ResourceConfig
}

func NewCgroupManagerV1(path string) *CgroupManagerV1 {
	return &CgroupManagerV1{
		Path: path,
	}
}

func (c *CgroupManagerV1) Apply(pid int) error {
	for _, subSysIns := range v1.SubsystemsIns {
		subSysIns.Apply(c.Path, pid)
	}
	return nil
}

func (c *CgroupManagerV1) Set(res *cgroups.ResourceConfig) error {
	for _, subSysIns := range v1.SubsystemsIns {
		subSysIns.Set(c.Path, res)
	}
	return nil
}

func (c *CgroupManagerV1) Destroy() error {
	for _, subSysIns := range v1.SubsystemsIns {
		subSysIns.Remove(c.Path)
	}
	return nil
}
