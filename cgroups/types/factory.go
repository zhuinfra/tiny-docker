package types

import (
	"os"
	"path"

	cgroups "tiny-docker/cgroups"
)

func NewCgroupManager(containerID string) cgroups.CgroupManager {
	path := path.Join(cgroups.CgroupDir, containerID)
	// 判断是否是 cgroup v2
	if isCgroupV2() {
		InitCgroupV2()
		return &CgroupManagerV2{
			Path: path,
		}
	}

	// 默认 v1
	return &CgroupManagerV1{
		Path: path,
	}
}

func isCgroupV2() bool {
	_, err := os.Stat("/sys/fs/cgroup/cgroup.controllers")
	return err == nil
}
