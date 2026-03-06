package container

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"syscall"
)

func NewParentProcess(tty bool, volume string, containerID string, image string) (*exec.Cmd, *os.File, error) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		slog.Error("New pipe error", "error", err)
		return nil, nil, err
	}
	// /proc/self/exe指向当前程序, 相当于执行docker init command
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS |
			syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNET,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		dirURL := fmt.Sprintf(DefaultInfoLocation, containerID)
		if err := os.MkdirAll(dirURL, 0622); err != nil {
			slog.Error("NewParentProcess mkdir error", "err", err)
			return nil, nil, err
		}
		stdLogFilePath := dirURL + ContainerLogFile
		stdLogFile, err := os.Create(stdLogFilePath)
		if err != nil {
			slog.Error("NewParentProcess create file error", "err", err)
			return nil, nil, err
		}
		cmd.Stdout = stdLogFile
	}

	cmd.ExtraFiles = []*os.File{readPipe}

	NewWorkSpace(containerID, image, volume)
	cmd.Dir = GetMerged(containerID)

	return cmd, writePipe, nil
}

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}
