package v1

var SubsystemsIns = []Subsystem{
	&MemorySubSystem{},
	&CpuSubSystem{},
	&CpusetSubSystem{},
}
