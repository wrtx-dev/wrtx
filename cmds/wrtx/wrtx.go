package main

import (
	"os"
	_ "wrtx/internal/init"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type pidMsg struct {
	ChildPID      int `json:"childpid"`
	GrandChildPid int `json:"grandchildpid"`
}

const stdioCount = 3

func main() {
	app := cli.App{
		Name:  "wrtx",
		Usage: "namespace example",
		Commands: []*cli.Command{
			&runcmd,
			// {
			// 	Name:  "cg",
			// 	Usage: "cgroup test",
			// 	Action: func(ctx *cli.Context) error {
			// 		return nil

			// 	},
			// },
		},
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
