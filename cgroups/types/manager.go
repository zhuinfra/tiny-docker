package types

import (
	"path"
	cgroups "tiny-docker/cgroups"
)

func NewCgroupManager(containerID string) cgroups.CgroupManager {
	path := path.Join(cgroups.CgroupDir, containerID)
	return &CgroupManagerV1{Path: path}
}
