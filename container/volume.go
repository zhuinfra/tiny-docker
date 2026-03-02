package container

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strings"
)

func mountVolume(mntPath, hostPath, containerPath string) {
	if err := os.Mkdir(hostPath, 0777); err != nil {
		slog.Info("os.Mkdir error", "error", err)
	}
	containerPathInHost := path.Join(mntPath, containerPath)
	if err := os.Mkdir(containerPathInHost, 0777); err != nil {
		slog.Info("os.Mkdir error", "error", err)
	}
	cmd := exec.Command("mount", "-o", "bind", hostPath, containerPathInHost)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		slog.Error("mount error", "error", err)
	}
}

func umountVolume(mntPath, containerPath string) {
	containerPathInHost := path.Join(mntPath, containerPath)
	cmd := exec.Command("umount", containerPathInHost)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		slog.Error("umount error", "error", err)
	}
}

func volumeExtract(volume string) (hostPath, containerPath string, err error) {
	parts := strings.Split(volume, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("volume input error")
	}
	hostPath = parts[0]
	containerPath = parts[1]
	if hostPath == "" || containerPath == "" {
		return "", "", fmt.Errorf("volume input error")
	}
	return hostPath, containerPath, nil
}
