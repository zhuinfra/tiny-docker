package v1

import "tiny-docker/cgroups"

var SubsystemsIns = []cgroups.Subsystem{
	&MemorySubSystem{},
	&CpuSubSystem{},
	&CpusetSubSystem{},
}
