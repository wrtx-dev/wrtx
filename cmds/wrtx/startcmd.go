package main

import (
	"fmt"
	"os"
	"wrtx/internal/config"
	"wrtx/internal/instances"

	"github.com/urfave/cli/v2"
)

var startCmd = cli.Command{
	Name:      "start",
	Usage:     "start wrtx's instance",
	ArgsUsage: "instance name",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "config",
			Usage: "path to config file",
		},
	},
	Action: start,
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

	conf, err := loadInstanceConfig(fmt.Sprintf("%s/%s/config.json", globalConfig.InstancesPath, name))
	if err != nil {
		return err
	}
	if _, err := os.Stat(conf.StatusFile); err == nil {
		return fmt.Errorf("instance %s is already running", name)
	}
	return instances.StartInstance(conf)
}

func loadInstanceConfig(confPath string) (*config.WrtxConfig, error) {
	conf := config.WrtxConfig{}
	if _, err := os.Stat(confPath); err != nil {
		if os.IsNotExist(err) {
			return &conf, err
		}
		return &conf, fmt.Errorf("load config file: %s error: %v", confPath, err)
	}
	if err := conf.Load(confPath); err != nil {
		return nil, fmt.Errorf("load config file: %s error: %v", confPath, err)
	}
	return &conf, nil
}
