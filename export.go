package main

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"tiny-docker/container"
)

// 镜像打包
func ExportContainer(containerId string, imageName string) error {
	mntPath := container.GetMerged(containerId)
	if _, err := os.Stat(mntPath); os.IsNotExist(err) {
		slog.Error("mntPath not exists")
		return nil
	}

	var outputPath string
	if filepath.IsAbs(imageName) {
		outputPath = imageName
	} else {
		cwd, _ := os.Getwd()
		outputPath = filepath.Join(cwd, imageName)
	}
	slog.Info("Export container", "outputPath", outputPath, "mntPath", mntPath)

	if _, err := exec.Command("tar", "-czf", outputPath, "-C", mntPath, ".").CombinedOutput(); err != nil {
		slog.Error("Tar error.", "err", err)
		return nil
	}
	return nil
}
