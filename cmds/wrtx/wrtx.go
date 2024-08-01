package main

import (
	"fmt"
	"os"
	_ "wrtx/internal/init"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var WrtxVersion string

func main() {
	app := cli.App{
		Name:  "wrtx",
		Usage: fmt.Sprintf("Run openwrt quickly and easily in linux namespaces, version: %s", WrtxVersion),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "conf",
				Usage:   "config file path",
				EnvVars: []string{"WRTX_CONFIG"},
			},
		},
		Commands: []*cli.Command{
			&runcmd,
			&importCmd,
			&execCmd,
			&shellCmd,
			&stopCmd,
			&proxyCmd,
			&startCmd,
			&agentCmd,
			&rmImageCmd,
			&lsiCmd,
			&pscmd,
		},
	}
	logrus.SetFormatter(&logrus.TextFormatter{})
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
