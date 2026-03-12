package cgroups

type CgroupManager interface {
	Apply(pid int) error
	Set(res *ResourceConfig) error
	Destory() error
}
