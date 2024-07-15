package main

import "github.com/urfave/cli/v2"

var importCmd = cli.Command{
	Name:  "import",
	Usage: "import openwrt's zip package",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "named image",
		},
	},
	Action: importAction,
}

func importAction(ctx *cli.Context) error {
	return nil
}
