package v1

import (
	cgroups "tiny-docker/cgroups"
)

func NewCgroupManager(path string) cgroups.CgroupManager {
	return &CgroupManagerV1{}
}
