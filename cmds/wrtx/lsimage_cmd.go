package main

import (
	"fmt"
	"os"
	"wrtx/internal/config"

	"github.com/urfave/cli/v2"
)

var lsiCmd = cli.Command{
	Name:    "lsi",
	Usage:   "List all images in the registry",
	Aliases: []string{"list-images"},
	Action:  listImages,
}

func listImages(ctx *cli.Context) error {
	gConfPath := ctx.String("conf")
	conf, err := config.GetGlobalConfig(gConfPath)
	if err != nil {
		return err
	}
	dirs, err := os.ReadDir(conf.ImagePath)
	if err != nil {
		return err
	}
	for _, dir := range dirs {
		if dir.IsDir() {
			fmt.Println(dir.Name())
		}
	}
	return nil
}
