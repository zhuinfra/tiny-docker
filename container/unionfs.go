package container

import (
	"log/slog"
	"os"
	"os/exec"
)

func NewWorkSpace(containerID string, imageName string, volume string) {
	creatLower(imageName)
	createDirs(containerID)
	mountOverlayFS(containerID, imageName)
	if volume != "" {
		mntPath := GetMerged("containerID")
		hostPath, containerPath, err := volumeExtract(volume)
		if err != nil {
			slog.Error("volumeExtract error", "error", err)
			os.Exit(1)
		}
		mountVolume(mntPath, hostPath, containerPath)
	}
}

func DeleteWorkSpace(containerID string, volume string) {
	if volume != "" {
		_, containerPath, err := volumeExtract(volume)
		if err != nil {
			slog.Error("volumeExtract error", "error", err)
			return
		}
		mntPath := GetMerged(containerID)
		umountVolume(mntPath, containerPath)
	}
	umountOverlayFS(containerID)
	deleteDirs(containerID)
}

// 创建只读层,　存放镜像(程序运行所需要的文件集合)
// 可复用
func creatLower(imageName string) error {
	lowerPath := GetLower(imageName)
	ImagePath := GetImage(imageName)
	exist, err := PathExists(lowerPath)
	if err != nil {
		return err
	}
	if !exist {
		if err := os.MkdirAll(lowerPath, 0622); err != nil {
			slog.Error("Mkdir unTarFolderUrl error.", "err", err)
			return err
		}

		if _, err := exec.Command("tar", "-xvf", ImagePath, "-C", lowerPath).CombinedOutput(); err != nil {
			slog.Error("Untar image error.", "err", err)
			return err
		}
	}
	return nil
}

// 创建可写层用于overlayfs 挂载
func createDirs(containerID string) {
	dirs := []string{
		GetWorker(containerID),
		GetUpper(containerID),
		GetMerged(containerID),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0622); err != nil {
			slog.Error("Mkdir dir error.", "err", err)
		}
	}
}

// 创建overlayfs挂载点
func mountOverlayFS(containerName string, imageName string) error {
	lowerPath := GetLower(imageName)
	upperPath := GetUpper(containerName)
	workPath := GetWorker(containerName)
	options := GetOverlayFSDirs(lowerPath, upperPath, workPath)
	mergedPath := GetMerged(containerName)

	// 执行命令
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", options, mergedPath)
	slog.Info("Executing mount command", "command", cmd.String())
	if err := cmd.Run(); err != nil {
		slog.Error("Mount overlayfs error.", "err", err)
		return err
	}
	return nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func umountOverlayFS(containerID string) {
	mntPath := GetMerged(containerID)
	cmd := exec.Command("umount", mntPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	slog.Info("umountOverlayFS")
	if err := cmd.Run(); err != nil {
		slog.Error(err.Error())
	}
}

func deleteDirs(containerID string) {
	dirs := []string{
		GetMerged(containerID),
		GetUpper(containerID),
		GetWorker(containerID),
		GetLower(containerID),
		GetRoot(containerID), // root 目录也要删除
	}

	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			slog.Error("Remove dir error", "dir", dir, "err", err)
		}
	}
}
