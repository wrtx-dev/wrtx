package main

import (
	"fmt"
	"os"
	"wrtx/internal/agent"
	"wrtx/internal/config"

	"github.com/urfave/cli/v2"
)

var startCmd = cli.Command{
	Name:      "start",
	Usage:     "start wrtx's instance",
	ArgsUsage: "instance name",
	Action:    start,
}

func start(ctx *cli.Context) error {
	globalConfPath := ctx.String("conf")
	globalConfLoaded := false
	if globalConfPath == "" {
		globalConfPath = config.DefaultConfPath
	}
	globalConfig := config.NewGlobalConf()
	if err := globalConfig.Load(globalConfPath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("load global config error: %v", err)
		}
	} else {
		globalConfLoaded = true
	}
	if !globalConfLoaded {
		globalConfig = config.DefaultGlobalConf()
	}

	name := ctx.Args().First()
	if name == "" {
		name = "openwrt"
	}
	return agent.StartWrtxInstance(fmt.Sprintf("%s/%s/config.json", globalConfig.InstancesPath, name))

}
