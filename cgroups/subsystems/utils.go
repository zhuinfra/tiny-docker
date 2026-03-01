package subsystems

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

// 查找指定subsystem的cgroup根节点路径
func FindCgroupMountPoint(subsystem string) string {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Split(txt, " ")
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				return fields[4]
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return ""
	}
	return ""
}

// 获取cgroup路径
func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRoot := FindCgroupMountPoint(subsystem)
	// 1.状态正常，返回完整路径 2.状态异常，创建为true则创建 3.状态异常，创建为false则返回错误
	if _, err := os.Stat(path.Join(cgroupRoot, cgroupPath)); err != nil {
		if os.IsNotExist(err) && autoCreate {
			if err := os.Mkdir(path.Join(cgroupRoot, cgroupPath), 0755); err != nil {
				return "", fmt.Errorf("error create cgroup %v", err)
			}
		} else {
			return "", fmt.Errorf("cgroup path error %v", err)
		}
	}
	return path.Join(cgroupRoot, cgroupPath), nil
}
