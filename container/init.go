package container

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func InitContainer() error {
	slog.Info("InitContainer")
	cmdArray := readUserCommand()
	if len(cmdArray) == 0 {
		return fmt.Errorf("readUserCommand error, cmdArray is nil")
	}

	// 切断 mount 传播, mount --make-rprivate /
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return err
	}

	setupMount()

	env := containerEnv()

	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		slog.Error("LookPath", "err", err)
		return err
	}
	if err := syscall.Exec(path, cmdArray, env); err != nil {
		slog.Error("Exec", "err", err)
	}
	return nil
}

func containerEnv() []string {

	env := []string{
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		"HOME=/root",
		"TERM=xterm",
	}

	// 可选：继承宿主 TERM
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "TERM=") {
			env = append(env, e)
		}
	}

	return env
}
func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	defer pipe.Close()
	msg, err := io.ReadAll(pipe)
	if err != nil {
		slog.Error("ReadAll", "err", err)
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

func setupMount() {
	pwd, err := os.Getwd()
	if err != nil {
		slog.Error("Get current location error", "err", err)
		return
	}
	slog.Info("setupMount", "pwd", pwd)
	pivotRoot(pwd)

	// MS_NOEXEC在本文件系统中不允许运行其他程序
	// MS_NOSUID在本系统中运行程序的时候，不允许set-user-ID或set-group-ID
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	// 挂载proc文件系统, ps命令依赖proc文件系统
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	// 挂载一个安全、隔离的 tmpfs 文件系统到 /dev 目录，为容器提供一个干净的设备文件环境，同时通过挂载选项增强安全性和权限控制
	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
}

func pivotRoot(rootfs string) error {
	// pivot_root需要满足new_root和put_old为不同文件系统，否则会失败
	if err := syscall.Mount(rootfs, rootfs, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount rootfs to itself error: %v", err)
	}
	// 创建rootfs/.pivot_root存储旧rootfs
	pivotDir := filepath.Join(rootfs, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return fmt.Errorf("mkdir .pivot_root error: %v", err)
	}
	// pivot_root 到新的rootfs, 挂载点此时还能在mount中看见
	if err := syscall.PivotRoot(rootfs, pivotDir); err != nil {
		return fmt.Errorf("pivot_root error: %v", err)
	}
	// 修改工作目录
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / error: %v", err)
	}
	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir error: %v", err)
	}
	// 删除.pivot_root
	return os.Remove(pivotDir)
}
