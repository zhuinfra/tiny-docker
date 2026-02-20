package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	// handler := slog.NewTextHandler(
	// 	os.Stdout,
	// 	&slog.HandlerOptions{

	// 		Level: slog.LevelDebug,
	// 	},
	// )

	// logger := slog.New(handler)

	// slog.SetDefault(logger)

	// slog.Info("cli start")
	// cmd := &cli.Command{
	// 	Name:  "tiny-docker",
	// 	Usage: "A custom Docker-like implementation (tiny-docker) demonstrating core container commands (init/run/ps/logs/exec/stop/rm/commit/network) with Linux namespace isolation and cgroup resource limiting.",
	// 	Action: func(context.Context, *cli.Command) error {
	// 		fmt.Println("boom! I say!")
	// 		return nil
	// 	},
	// }

	// if err := cmd.Run(context.Background(), os.Args); err != nil {
	// 	slog.Error("cli error", "error", err)
	// }
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
			{
				Name:  "run",
				Usage: "add a task to the list",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("added task: ", cmd.Args().First())
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
