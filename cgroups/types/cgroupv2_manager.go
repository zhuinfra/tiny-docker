package types

import (
	"tiny-docker/cgroups"
	v2 "tiny-docker/cgroups/v2"
)

type CgroupManagerV2 struct {
	Path     string
	Resource *cgroups.ResourceConfig
}

func NewCgroupManagerV2(path string) *CgroupManagerV2 {
	return &CgroupManagerV2{
		Path: path,
	}
}

func (c *CgroupManagerV2) Apply(pid int) error {
	for _, subSysIns := range v2.SubsystemsIns {
		subSysIns.Apply(c.Path, pid)
	}
	return nil
}

func (c *CgroupManagerV2) Set(res *cgroups.ResourceConfig) error {
	for _, subSysIns := range v2.SubsystemsIns {
		subSysIns.Set(c.Path, res)
	}
	return nil
}

func (c *CgroupManagerV2) Destory() error {
	for _, subSysIns := range v2.SubsystemsIns {
		subSysIns.Remove(c.Path)
	}
	return nil
}
