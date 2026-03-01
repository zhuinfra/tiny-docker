package subsystems

// 资源限制配置
type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	CpuSet      string
}

// Subsystem接口, 每个subsystem实现该接口
// cgroup的增删, 限制配置
type Subsystem interface {
	Name() string
	Set(cgroupPath string, res *ResourceConfig) error
	Apply(cgroupPath string, pid int) error
	Remove(cgroupPath string) error
}

var (
	SubsystemsIns = []Subsystem{
		&MemorySubSystem{},
		&CpuSubSystem{},
		&CpusetSubSystem{},
	}
)
