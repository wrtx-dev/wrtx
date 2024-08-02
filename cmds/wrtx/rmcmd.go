package main

import (
	"fmt"
	"os"
	"path/filepath"
	"wrtx/internal/config"

	"github.com/urfave/cli/v2"
)

var rmCmd = cli.Command{
	Name:    "rm",
	Aliases: []string{"remove"},
	Usage:   "Remove a instance",

	ArgsUsage: "instance_name",
	Action:    rm,
}

func rm(ctx *cli.Context) error {
	args := ctx.Args().Slice()
	if len(args) < 1 {
		return fmt.Errorf("missing instance name")
	}
	gConf, err := config.GetGlobalConfig(ctx.String("conf"))
	if err != nil {
		return fmt.Errorf("failed to get global config: %v", err)
	}
	instances, err := os.ReadDir(gConf.InstancesPath)
	if err != nil {
		return fmt.Errorf("failed to read instances directory: %v", err)
	}
	for _, arg := range args {
		found := false
		for _, instance := range instances {
			if instance.Name() == arg {
				found = true
				statusFile := filepath.Join(gConf.InstancesPath, instance.Name(), "status.json")
				if _, err := os.Stat(statusFile); err != nil {
					if os.IsNotExist(err) {
						if err := os.RemoveAll(filepath.Join(gConf.InstancesPath, instance.Name())); err != nil {
							fmt.Printf("failed to remove instance directory: %v\n", err)
						} else {
							fmt.Println("Instance", instance.Name(), "removed successfully")
						}
					} else {
						fmt.Printf("failed to check status file: %v, skipping remove this instance\n", err)
					}
				} else {
					fmt.Printf("Instance %s is running, please stop it first\n", instance.Name())
				}
			}

		}
		if !found {
			fmt.Printf("Instance %s not found\n", arg)
		}
	}

	return nil
}
