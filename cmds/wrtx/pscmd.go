package main

import (
	"fmt"
	"os"
	"path/filepath"
	"wrtx/internal/config"
	"wrtx/internal/instances"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var pscmd = cli.Command{
	Name:   "ps",
	Usage:  "List all instances",
	Action: psAction,
}

func psAction(ctx *cli.Context) error {
	gconf, err := config.GetGlobalConfig(ctx.String("conf"))
	if err != nil {
		return errors.WithMessagef(err, "Failed to get global config: %v", err)
	}
	allInstances, err := os.ReadDir(gconf.InstancesPath)
	if err != nil {
		return errors.WithMessagef(err, "Failed to read instances directory: %v", err)
	}
	for _, instance := range allInstances {
		p := filepath.Join(gconf.InstancesPath, instance.Name())
		running := true
		statusPath := filepath.Join(p, "status.json")
		if _, err := os.Stat(statusPath); os.IsNotExist(err) {
			fmt.Println("instance:", instance.Name(), "status: not running")
			running = false
		}
		if running {
			if status, err := instances.GetInstanceStatus(statusPath); err == nil {
				fmt.Println("instance:", instance.Name(), "status:", status.Status)
			} else {
				fmt.Println("instance:", instance.Name(), "status: reading status err:", err)
			}
		}
	}
	return nil
}
