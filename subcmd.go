// subcom.go 处理tiny-docker的所有子命令解析与分发
// 包括init/run/ps/logs/exec/stop/rm/commit/network等命令的参数解析和执行入口。
package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
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
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		args := cmd.Args()
		if args.Len() < 1 {
			return fmt.Errorf("missing command")
		}
		comArray := args.Slice()
		command := strings.Join(comArray, " ")
		tty := cmd.Bool("it")
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: cmd.String("m"),
			CpuSet:      cmd.String("cpus"),
			CpuShare:    cmd.String("c"),
		}
		Run(tty, command, resConf) // 传递剩余参数
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
		args := cmd.Args()
		command := args.First()
		err := container.InitContainer(command)
		return err
	},
}
