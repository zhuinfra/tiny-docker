package types

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"tiny-docker/cgroups"
)

const cgroupV2Root = "/sys/fs/cgroup"

type CgroupManagerV2 struct {
	Path     string
	Resource *cgroups.ResourceConfig
}

func NewCgroupManagerV2(path string) *CgroupManagerV2 {
	return &CgroupManagerV2{
		Path: path,
	}
}

/*
初始化 cgroup v2

/sys/fs/cgroup

	└── tiny-docker
	        └── containerID
*/
func InitCgroupV2() error {

	// enable root controllers
	if err := enableControllers(cgroupV2Root); err != nil {
		return err
	}

	// create tiny-docker dir
	tdPath := path.Join(cgroupV2Root, cgroups.CgroupDir)
	if err := os.MkdirAll(tdPath, 0755); err != nil {
		return err
	}

	// enable controllers for tiny-docker
	if err := enableControllers(tdPath); err != nil {
		return err
	}

	return nil
}

func enableControllers(cgroupPath string) error {

	data, err := os.ReadFile(path.Join(cgroupPath, "cgroup.controllers"))
	if err != nil {
		return err
	}

	controllers := strings.Fields(string(data))
	if len(controllers) == 0 {
		return nil
	}

	var enable []string
	for _, ctrl := range controllers {
		enable = append(enable, "+"+ctrl)
	}

	return os.WriteFile(
		path.Join(cgroupPath, "cgroup.subtree_control"),
		[]byte(strings.Join(enable, " ")),
		0644,
	)
}

func (c *CgroupManagerV2) getCgroupPath() (string, error) {

	cgroupPath := path.Join(cgroupV2Root, c.Path)

	if err := os.MkdirAll(cgroupPath, 0755); err != nil {
		return "", err
	}

	return cgroupPath, nil
}

func (c *CgroupManagerV2) Set(res *cgroups.ResourceConfig) error {

	c.Resource = res

	cgroupPath, err := c.getCgroupPath()
	if err != nil {
		return err
	}

	// memory limit
	if res.MemoryLimit != "" {

		if err := os.WriteFile(
			filepath.Join(cgroupPath, "memory.max"),
			[]byte(res.MemoryLimit),
			0644,
		); err != nil {
			return err
		}
	}

	// cpu limit
	if res.Cpus > 0 {

		period := 100000
		quota := int(res.Cpus * float64(period))

		value := fmt.Sprintf("%d %d", quota, period)

		if err := os.WriteFile(
			filepath.Join(cgroupPath, "cpu.max"),
			[]byte(value),
			0644,
		); err != nil {
			return err
		}
	}

	// cpu share (v1 cpu.shares -> v2 cpu.weight)
	if res.CpuShare != "" {

		shares, err := strconv.Atoi(res.CpuShare)
		if err != nil {
			return err
		}

		// convert shares to weight
		weight := 1 + (shares-2)*9999/262142

		if weight < 1 {
			weight = 1
		}
		if weight > 10000 {
			weight = 10000
		}

		if err := os.WriteFile(
			filepath.Join(cgroupPath, "cpu.weight"),
			[]byte(strconv.Itoa(weight)),
			0644,
		); err != nil {
			return err
		}
	}

	// cpuset
	if res.CpuSet != "" {

		if err := os.WriteFile(
			filepath.Join(cgroupPath, "cpuset.cpus"),
			[]byte(res.CpuSet),
			0644,
		); err != nil {
			return err
		}
	}

	return nil
}

func (c *CgroupManagerV2) Apply(pid int) error {

	cgroupPath, err := c.getCgroupPath()
	if err != nil {
		return err
	}

	return os.WriteFile(
		filepath.Join(cgroupPath, "cgroup.procs"),
		[]byte(strconv.Itoa(pid)),
		0644,
	)
}

func (c *CgroupManagerV2) Destroy() error {

	cgroupPath := path.Join(cgroupV2Root, c.Path)

	return os.RemoveAll(cgroupPath)
}
