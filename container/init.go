package container

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
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

	// MS_NOEXEC在本文件系统中不允许运行其他程序
	// MS_NOSUID在本系统中运行程序的时候，不允许set-user-ID或set-group-ID
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	// 挂载proc文件系统, ps命令依赖proc文件系统
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")

	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		slog.Error("LookPath", "err", err)
		return err
	}
	if err := syscall.Exec(path, cmdArray, os.Environ()); err != nil {
		slog.Error("Exec", "err", err)
	}
	return nil
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
