package main

import (
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

var shellCmd = cli.Command{
	Name:   "shell",
	Usage:  "run openwrt shell, /bin/ash --login",
	Action: shellAction,
}

func shellAction(ctx *cli.Context) error {
	cmd := exec.Command("/proc/self/exe", []string{"exec", "/bin/ash", "--login"}...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Run()
	return nil
}
