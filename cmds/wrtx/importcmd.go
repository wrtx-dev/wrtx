package main

import (
	"fmt"
	"os"
	"wrtx/internal/config"
	"wrtx/package/packageextract"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

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
	conf := ctx.String("conf")
	if conf == "" {
		conf = config.DefaultConfPath
	}
	globalConfig := config.NewGlobalConf()
	if err := globalConfig.Load(conf); err != nil {
		return errors.WithMessagef(err, "load config file: %s error", conf)
	}
	if globalConfig.ImagePath == "" {
		globalConfig.ImagePath = config.DefaultImagePath
	}
	imagePath := ctx.Args().First()
	name := ctx.String("name")
	if name == "" {
		name = "openwrt"
	}
	if stat, err := os.Stat(imagePath); err == nil {
		if stat.IsDir() {
			return errors.WithMessagef(err, "%s isn't an tar.tz file", imagePath)
		}
	} else {
		return errors.WithMessagef(err, "check path: %s error", imagePath)
	}
	savePath := fmt.Sprintf("%s/%s", globalConfig.ImagePath, name)
	if _, err := os.Stat(savePath); err == nil {
		return fmt.Errorf("image saved path: %s alreay exist", savePath)
	}
	if err := os.Mkdir(savePath, os.ModePerm); err != nil {
		return errors.WithMessagef(err, "create path: %s error", savePath)
	}
	if err := packageextract.UnTarGZ(imagePath, savePath); err != nil {
		return errors.WithMessagef(err, "extract image package to %s error", savePath)
	}
	fmt.Printf("save image to %s\n\n", savePath)
	return nil
}
