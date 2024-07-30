package main

import (
	"fmt"
	"os"
	"wrtx/internal/agent"
	"wrtx/internal/instances"

	"github.com/urfave/cli/v2"
)

var agentCmd = cli.Command{
	Name:   "agent",
	Usage:  "run the wrtx agent",
	Hidden: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "config",
			Usage: "path to the config file",
		},
	},
	Action: agentStart,
}

func agentStart(ctx *cli.Context) error {
	confPath := ctx.String("config")

	conf, err := agent.LoadInstanceConfig(confPath)
	if err != nil {
		return err
	}
	if _, err := os.Stat(conf.StatusFile); err == nil {
		return fmt.Errorf("instance %s is already running", conf.InstanceName)
	}
	return instances.StartInstance(conf)
}
