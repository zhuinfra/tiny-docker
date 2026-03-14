package cgroups

// 资源限制配置
type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	Cpus        float64
	CpuSet      string
}

var CgroupDir = "tiny-docker"
