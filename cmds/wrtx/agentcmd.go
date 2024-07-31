package main

import (
	"bytes"
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
		status := instances.NewStatus()
		if err := status.Load(conf.StatusFile); err != nil {
			return fmt.Errorf("found status file but failed to load: %v", err)
		}
		if status.AgentPid > 0 {
			pid := status.AgentPid
			if _, err := os.Stat(fmt.Sprintf("/proc/%d", pid)); err == nil {
				buf, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
				if err == nil {
					bufs := bytes.Split(buf, []byte{0})
					if string(bufs[0]) == "wrtd" {
						return fmt.Errorf("wrtx agent is already running with pid %d", pid)
					}
				}
			}
			os.RemoveAll(conf.StatusFile)
		}
	}
	return instances.StartInstance(conf)
}
