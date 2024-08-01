package main

import (
	"fmt"
	"os"
	"wrtx/internal/config"
	"wrtx/internal/instances"

	"github.com/urfave/cli/v2"
)

var rmImageCmd = cli.Command{
	Name:      "rmi",
	Aliases:   []string{"rmimage"},
	Usage:     "Remove an image from the registry",
	ArgsUsage: " IMAGE...",
	Action:    rmImage,
}

func rmImage(c *cli.Context) error {
	args := c.Args().Slice()
	if len(args) < 1 {
		return fmt.Errorf("missing image name")
	}
	conf := c.String("conf")
	globalConfig, err := config.GetGlobalConfig(conf)
	if err != nil {
		return fmt.Errorf("failed to get global config: %v", err)
	}
	usingImages, err := instances.GetAllImageInUsed(conf)
	if err != nil {
		return fmt.Errorf("failed to get all images in used: %v", err)
	}
	willDels := []string{}
	for _, rmImg := range args {
		used := false
		for _, usedImg := range usingImages {

			if usedImg == rmImg {
				used = true
				break
			}
		}
		if used {
			fmt.Printf("Image %s is in used, skip it\n", rmImg)
			continue
		}
		willDels = append(willDels, rmImg)
	}
	if len(willDels) == 0 {
		return nil
	}
	for _, rmImg := range willDels {
		if _, err := os.Stat(fmt.Sprintf("%s/%s", globalConfig.ImagePath, rmImg)); os.IsNotExist(err) {
			fmt.Printf("Image %s not found, skip it\n", rmImg)
			continue
		}
		if err := os.RemoveAll(fmt.Sprintf("%s/%s", globalConfig.ImagePath, rmImg)); err != nil {
			fmt.Printf("failed to remove image %s: %v", rmImg, err)
		}
		fmt.Printf("Removed image %s\n", rmImg)
	}
	return nil
}
