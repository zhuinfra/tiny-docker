package cgroups

// 资源限制配置
type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	Cpus        float64
	CpuSet      string
}

// Subsystem接口, 每个subsystem实现该接口
// 根据ResourceConfig在每一种subsystem中设置资源限制
type Subsystem interface {
	Name() string
	Set(cgroupPath string, res *ResourceConfig) error
	Apply(cgroupPath string, pid int) error
	Remove(cgroupPath string) error
}

var CgroupDir = "tiny-docker"
