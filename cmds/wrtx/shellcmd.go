package main

import (
	"fmt"
	"os"
	"os/exec"
	"wrtx/internal/config"

	"github.com/urfave/cli/v2"
)

var shellCmd = cli.Command{
	Name:      "shell",
	Usage:     "run openwrt shell, /bin/ash --login",
	ArgsUsage: " instance_name",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "instance name",
		},
	},
	Action: shellAction,
}

func shellAction(ctx *cli.Context) error {
	instanceName := ctx.Args().First()
	gConf := ctx.String("conf")
	args := []string{"exec"}
	globalConfig, err := config.GetGlobalConfig(gConf)
	if err != nil {
		return fmt.Errorf("failed to get global config: %v", err)
	}
	if instanceName == "" {
		instanceName = globalConfig.DefaultInstanceName
	}
	if gConf != "" {
		args = append(args, "--conf", gConf)
	}
	args = append(args, []string{instanceName, "/bin/ash", "--login"}...)

	cmd := exec.Command("/proc/self/exe", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Run()
	return nil
}
