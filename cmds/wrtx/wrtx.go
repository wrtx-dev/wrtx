package main

import (
	"fmt"
	"os"
	_ "wrtx/internal/init"

	_ "embed"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var WrtxVersion string

type pidMsg struct {
	ChildPID      int `json:"childpid"`
	GrandChildPid int `json:"grandchildpid"`
}

const stdioCount = 3

func main() {
	app := cli.App{
		Name:  "wrtx",
		Usage: fmt.Sprintf("simple to run openwrtx in namespace, version: %s", WrtxVersion),
		Commands: []*cli.Command{
			&runcmd,
			// {
			// 	Name:  "cg",
			// 	Usage: "cgroup test",
			// 	Action: func(ctx *cli.Context) error {
			// 		return nil

			// 	},
			// },
			&importCmd,
			&execCmd,
			&shellCmd,
		},
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
