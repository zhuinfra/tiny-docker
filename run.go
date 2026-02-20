package main

import (
	"log/slog"
	"os"
	"tiny-docker/container"
)

func Run(tty bool, command string) {
	parent := container.NewParentProcess(tty, command)
	if err := parent.Start(); err != nil {
		slog.Error("container startup failed", "error", err)
	}
	parent.Wait()
	os.Exit(1)
}
