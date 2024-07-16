package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.App{
		Name:  "wrtxproxy",
		Usage: "openwrt web ui proxy",
		Commands: []*cli.Command{
			&serveCmd,
			&proxyCmd,
		},
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
