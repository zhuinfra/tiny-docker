package v1

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"strconv"
	"tiny-docker/cgroups"
)

// cpu控制器
type CpuSubSystem struct {
}

func (s *CpuSubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		if err := os.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
		return nil
	} else {
		return err
	}
}
func (s *CpuSubSystem) Set(cgroupPath string, res *cgroups.ResourceConfig) error {
	subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true)
	if err != nil {
		return err
	}

	// cpu.shares
	if res.CpuShare != "" {
		if err := os.WriteFile(
			path.Join(subsysCgroupPath, "cpu.shares"),
			[]byte(res.CpuShare),
			0644,
		); err != nil {
			slog.Error("set cpu.shares fail", "err", err)
			return err
		}
	}

	// --cpus
	if res.Cpus > 0 {

		period := 100000
		quota := int(res.Cpus * float64(period))

		// cpu.cfs_period_us
		if err := os.WriteFile(
			path.Join(subsysCgroupPath, "cpu.cfs_period_us"),
			[]byte(strconv.Itoa(period)),
			0644,
		); err != nil {
			return err
		}

		// cpu.cfs_quota_us
		if err := os.WriteFile(
			path.Join(subsysCgroupPath, "cpu.cfs_quota_us"),
			[]byte(strconv.Itoa(quota)),
			0644,
		); err != nil {
			return err
		}
	}

	return nil
}

func (s *CpuSubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return err
	}
}

func (s *CpuSubSystem) Name() string {
	return "cpu"
}
