package v2

import "tiny-docker/cgroups"

var SubsystemsIns = []cgroups.Subsystem{
	&MemorySubSystem{},
	&CpuSubSystem{},
	&CpusetSubSystem{},
}
