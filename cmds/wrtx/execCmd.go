package main

import "github.com/urfave/cli/v2"

var execCmd = cli.Command{
	Name:   "exec",
	Usage:  "execute a command in openwrt",
	Action: execAction,
}

func execAction(ctx *cli.Context) error {
	return nil
}
