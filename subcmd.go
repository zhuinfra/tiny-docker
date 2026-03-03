// subcom.go 处理tiny-docker的所有子命令解析与分发
// 包括init/run/ps/logs/exec/stop/rm/commit/network等命令的参数解析和执行入口。
package main

import (
	"context"
	"fmt"
	"log/slog"
	"tiny-docker/cgroups/subsystems"
	"tiny-docker/container"

	"github.com/urfave/cli/v3"
)

var runCommand = &cli.Command{
	Name:  "run",
	Usage: "Create a container with namespaces and cgroups limit\nUsage: tiny-docker run [OPTIONS] -- COMMAND [ARGS...]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "it",
			Usage: "enable tty",
		},
		&cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		},
		&cli.StringFlag{
			Name:    "m",
			Aliases: []string{"memory"},
			Usage:   "memory limit",
		},
		&cli.IntFlag{
			Name:    "c",
			Aliases: []string{"cpu-shares"},
			Usage:   "CPU shares (relative weight)",
		},
		&cli.FloatFlag{
			Name:  "cpus",
			Usage: "Number of CPUs",
		},
		&cli.StringFlag{
			Name:  "v",
			Usage: "Bind a directory on the host to the container",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		slog.Info("runCommand start")
		args := cmd.Args()
		if args.Len() < 1 {
			return fmt.Errorf("missing command")
		}
		comArray := args.Slice()

		tty := cmd.Bool("it")
		detach := cmd.Bool("d")
		if tty && detach {
			return fmt.Errorf("it is invalid to use -it and -d together")
		}

		resConf := &subsystems.ResourceConfig{
			MemoryLimit: cmd.String("m"),
			CpuSet:      cmd.String("cpus"),
			CpuShare:    cmd.String("c"),
		}
		volume := cmd.String("v")
		Run(tty, comArray, resConf, volume) // 传递剩余参数
		return nil
	},
}

var initCommand = &cli.Command{
	/*
		1.执行容器初始化操作, 实则通过exec替换进程
	*/
	Name:  "init",
	Usage: "Init container process running user program. Do not call it outside",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		slog.Info("initCommand start")
		err := container.InitContainer()
		return err
	},
}

var exportCommand = &cli.Command{
	/*
		1.导出容器的rootfs
	*/
	Name:  "export",
	Usage: "export a container",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		slog.Info("exportCommand start")
		args := cmd.Args()
		if args.Len() < 2 {
			return fmt.Errorf("missing container name and image name")
		}
		if err := ExportContainer(args.Get(0), args.Get(1)); err != nil {
			return err
		}
		return nil
	},
}
