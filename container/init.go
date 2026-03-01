package container

import (
	"log/slog"
	"strings"
	"syscall"
)

func InitContainer(command string) error {
	slog.Info("InitContainer", "command", command)

	// 切断 mount 传播, mount --make-rprivate /
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return err
	}

	// MS_NOEXEC在本文件系统中不允许运行其他程序
	// MS_NOSUID在本系统中运行程序的时候，不允许set-user-ID或set-group-ID
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	// 挂载proc文件系统, ps命令依赖proc文件系统
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	argv := strings.Split(command, " ")
	command = argv[0]
	slog.Info("InitContainer", "command", command)
	if err := syscall.Exec(command, argv, []string{}); err != nil {
		slog.Error("Exec", "err", err)
	}
	return nil
}
