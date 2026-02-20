package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	// 定义命令参数
	cmd := &cli.Command{
		Name:  "tiny-docker",
		Usage: "A custom Docker-like implementation demonstrating core container commands",
		Description: `Tiny Docker is a lightweight container runtime that showcases fundamental container technologies.
Features:
  - Linux namespace isolation (PID, Mount, UTS, IPC, Network, User)
  - Cgroup resource limiting (CPU, memory)
  - Container lifecycle management (run, ps, logs, exec, stop, rm)
  - Image commit functionality
  - Basic network isolation`,

		Commands: []*cli.Command{
			runCommand,
			initCommand,
		},
	}

	// 程序启动前配置日志输出
	cmd.Before = func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		handler := slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		)
		logger := slog.New(handler)
		slog.SetDefault(logger)
		return ctx, nil
	}

	// 入口
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("cli error", "error", err)
		os.Exit(1)
	}
}
