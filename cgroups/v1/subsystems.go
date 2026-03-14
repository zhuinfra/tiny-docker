package v1

import "tiny-docker/cgroups"

// Subsystem接口, 每个subsystem实现该接口
// 根据ResourceConfig在每一种subsystem中设置资源限制
type Subsystem interface {
	Name() string
	Set(cgroupPath string, res *cgroups.ResourceConfig) error
	Apply(cgroupPath string, pid int) error
	Remove(cgroupPath string) error
}
