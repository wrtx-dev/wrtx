package main

import (
	"bytes"
	"fmt"
	"os"
	"wrtx/internal/agent"
	"wrtx/internal/config"
	"wrtx/internal/instances"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var agentCmd = cli.Command{
	Name:  "agent",
	Usage: "run the wrtx agent",
	Flags: []cli.Flag{
		&cli.StringFlag{
			EnvVars: []string{"WRTX_ICONF"},
			Usage:   "path to instance configuration file",
		},
	},
	Hidden: true,
	Action: agentStart,
}

func agentStart(ctx *cli.Context) error {
	confPath := ctx.String("iconf")
	gconfPath := ctx.String("conf")

	gconf, err := config.GetGlobalConfig(gconfPath)
	if err != nil {
		return err
	}
	fp, err := os.OpenFile(gconf.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		fmt.Println("Failed to open log file:", err)
	}

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
	logrus.SetOutput(fp)
	return instances.StartInstance(conf)
}
