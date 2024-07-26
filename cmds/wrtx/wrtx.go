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
		Commands: []*cli.Command{
			&runcmd,
			&importCmd,
			&execCmd,
			&shellCmd,
			&stopCmd,
			&proxyCmd,
			&startCmd,
		},
	}
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
