package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"tiny-docker/container"
)

func logContainer(containerID string) {
	logDir := fmt.Sprintf(container.DefaultInfoLocation, containerID)
	logPath := logDir + container.ContainerLogFile
	file, err := os.Open(logPath)
	if err != nil {
		slog.Error("Open log file error.")
		return
	}
	content, err := io.ReadAll(file)
	if err != nil {
		slog.Error("Log container read file error", "err", err)
		return
	}
	fmt.Fprint(os.Stdout, string(content))
}
