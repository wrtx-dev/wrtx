package main

import (
	"fmt"
	"wrtx/internal/agent"
	"wrtx/internal/config"

	"github.com/pkg/errors"
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
	globalConfig, err := config.GetGlobalConfig(globalConfPath)
	if err != nil {
		return errors.WithMessage(err, "failed to load global config")
	}

	name := ctx.Args().First()
	if name == "" {
		name = globalConfig.DefaultInstanceName
	}
	return agent.StartWrtxInstance(ctx.String("conf"), fmt.Sprintf("%s/%s/config.json", globalConfig.InstancesPath, name))

}
