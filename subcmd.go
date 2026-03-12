// subcom.go 处理tiny-docker的所有子命令解析与分发
// 包括init/run/ps/logs/exec/stop/rm/commit/network等命令的参数解析和执行入口。
package main

import (
	"context"
	"fmt"
	"log/slog"
	"tiny-docker/cgroups"
	"tiny-docker/container"
	"tiny-docker/network"

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
		&cli.StringFlag{
			Name:  "name",
			Usage: "Assign a name to the container",
		},
		&cli.StringSliceFlag{
			Name:  "e",
			Usage: "set environment variables",
		},
		&cli.StringFlag{
			Name:  "network",
			Usage: "Connect a container to a network",
		},
		&cli.StringSliceFlag{
			Name:  "p",
			Usage: "set port mapping,e.g. -p 8080:80 -p 30336:3306",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		slog.Info("runCommand start")
		args := cmd.Args()
		if args.Len() < 2 {
			return fmt.Errorf("missing image or command")
		}
		image := args.Get(0)
		comArray := args.Slice()[1:]

		tty := cmd.Bool("it")
		detach := cmd.Bool("d")
		if tty && detach {
			return fmt.Errorf("it is invalid to use -it and -d together")
		}

		resConf := &cgroups.ResourceConfig{
			MemoryLimit: cmd.String("m"),
			CpuSet:      cmd.String("cpus"),
			CpuShare:    cmd.String("c"),
		}
		volume := cmd.String("v")
		containerName := cmd.String("name")
		envSlice := cmd.StringSlice("e")

		network := cmd.String("network")
		portMapping := cmd.StringSlice("p")

		Run(tty, volume, network, containerName, image, comArray, envSlice, portMapping, resConf)
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
			return fmt.Errorf("missing container id and image name")
		}
		if err := ExportContainer(args.Get(0), args.Get(1)); err != nil {
			return err
		}
		return nil
	},
}

var psCommand = &cli.Command{
	Name:  "ps",
	Usage: "list all the containers",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		slog.Info("psCommand start")
		ListContainers()
		return nil
	},
}

var logsCommand = &cli.Command{
	Name:  "logs",
	Usage: "print logs of a container",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		slog.Info("logCommand start")
		args := cmd.Args()
		if args.Len() < 1 {
			return fmt.Errorf("missing container id")
		}
		logContainer(args.Get(0))
		return nil
	},
}

var execCommand = &cli.Command{
	Name:  "exec",
	Usage: "exec a command into a container",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		slog.Info("execCommand start")
		args := cmd.Args()
		if args.Len() < 2 {
			return fmt.Errorf("missing container id or command")
		}
		containerID := args.Get(0)
		cmdArray := args.Slice()[1:]
		ExecContainer(containerID, cmdArray)
		return nil
	},
}

var stopCommand = &cli.Command{
	Name:  "stop",
	Usage: "stop a container",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		slog.Info("stopCommand start")
		args := cmd.Args()
		if args.Len() < 1 {
			return fmt.Errorf("missing container id")
		}
		stopContainer(args.Get(0))
		return nil
	},
}

var rmCommand = &cli.Command{
	Name:  "rm",
	Usage: "remove unused containers",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "f",
			Usage: "force remove container",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		slog.Info("rmCommand start")
		args := cmd.Args()
		if args.Len() < 1 {
			return fmt.Errorf("missing container id")
		}
		isForce := cmd.Bool("f")
		removeContainer(args.Get(0), isForce)
		return nil
	},
}

var networkCommand = &cli.Command{
	Name:  "network",
	Usage: "container network commands",
	Commands: []*cli.Command{
		{
			Name:  "create",
			Usage: "create a container network",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "driver",
					Usage: "network driver",
					Value: "bridge",
				},
				&cli.StringFlag{
					Name:     "subnet",
					Usage:    "network subnet (e.g., 172.18.0.0/16)",
					Required: true,
				},
			},
			Action: func(ctx context.Context, cmd *cli.Command) error {
				slog.Info("network create start")
				args := cmd.Args()
				if args.Len() < 1 {
					return fmt.Errorf("missing network name")
				}
				driver := cmd.String("driver")
				subnet := cmd.String("subnet")
				network.CreateNetwork(driver, subnet, args.Get(0))
				return nil
			},
		},
		{
			Name:  "ls",
			Usage: "list all container networks",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				slog.Info("network list start")
				network.ListNetwork()
				return nil
			},
		},
		{
			Name:  "rm",
			Usage: "delete a container network",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "f",
					Usage: "force delete network",
				},
			},
			Action: func(ctx context.Context, cmd *cli.Command) error {
				slog.Info("network delete start")
				args := cmd.Args()
				if args.Len() < 1 {
					return fmt.Errorf("missing network name")
				}
				network.DeleteNetwork(args.Get(0))
				return nil
			},
		},
	},
}
