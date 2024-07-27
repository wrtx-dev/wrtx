package main

import (
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

var shellCmd = cli.Command{
	Name:  "shell",
	Usage: "run openwrt shell, /bin/ash --login",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "conf",
			Usage: "global config file path",
		},
		&cli.StringFlag{
			Name:  "name",
			Usage: "instance name",
		},
	},
	Action: shellAction,
}

func shellAction(ctx *cli.Context) error {
	instanceName := ctx.String("name")
	gConf := ctx.String("conf")
	args := []string{"exec"}
	if instanceName != "" {
		args = append(args, "--name", instanceName)
	}
	if gConf != "" {
		args = append(args, "--conf", gConf)
	}
	args = append(args, []string{"/bin/ash", "--login"}...)

	cmd := exec.Command("/proc/self/exe", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Run()
	return nil
}
