package main

import (
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

var shellCmd = cli.Command{
	Name:  "shell",
	Usage: "run openwrt shell,default:/bin/sh",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "sh",
			Usage: "path of shell's bin",
		},
	},
	Action: shellAction,
}

func shellAction(ctx *cli.Context) error {
	shell := ctx.String("sh")
	if shell == "" {
		shell = "/bin/sh"
	}

	cmd := exec.Command("/proc/self/exe", shell)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Run()
	return nil
}
