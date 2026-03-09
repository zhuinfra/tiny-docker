package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"text/tabwriter"
	"tiny-docker/container"
)

func ListContainers() {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, "")
	if _, err := os.Stat(dirURL); os.IsNotExist(err) {
		os.MkdirAll(dirURL, 0622)
	}
	files, err := os.ReadDir(dirURL)
	if err != nil {
		slog.Error("Read dir error", "err", err)
		return
	}

	var containers []*container.ContainerInfo
	for _, file := range files {
		tmpContainer, err := getContainerInfo(file)
		if err != nil {
			slog.Error("Get container info error", "err", err)
			continue
		}
		containers = append(containers, tmpContainer)
	}

	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreatedTime)
	}
	if err := w.Flush(); err != nil {
		slog.Error("flush error", "err", err)
		return
	}
}

func getContainerInfo(file os.DirEntry) (*container.ContainerInfo, error) {
	containerName := file.Name()
	configFileDir := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFileDir = configFileDir + container.ConfigName
	content, err := os.ReadFile(configFileDir)
	if err != nil {
		slog.Error("ReadFile error.", "err", err)
		return nil, err
	}
	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		slog.Error("Json unmarshal error")
		return nil, err
	}

	return &containerInfo, nil
}
