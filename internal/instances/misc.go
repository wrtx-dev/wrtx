package instances

import (
	"fmt"
	"os"
	"path/filepath"
	"wrtx/internal/config"
)

func CheckImages(globalConf *config.GlobalConfig, name string) (bool, error) {
	imgPath := fmt.Sprintf("%s/%s", globalConf.ImagePath, name)
	if _, err := os.Stat(imgPath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func GetAllInstancesConfig(path string) ([]*config.WrtxConfig, error) {
	confPath := path
	if confPath == "" {
		confPath = config.DefaultConfPath
	}
	glConfig := config.NewGlobalConf()
	if err := glConfig.Load(confPath); err != nil {
		return nil, err
	}

	instancesConfig := make([]*config.WrtxConfig, 0)
	dirs, err := os.ReadDir(glConfig.InstancesPath)
	if err != nil {
		return nil, err
	}
	for _, dir := range dirs {
		if dir.IsDir() {
			instanceConf := config.NewConf()
			if err := instanceConf.Load(fmt.Sprintf("%s/%s/config.json", glConfig.InstancesPath, dir.Name())); err != nil {
				continue
			}
			instancesConfig = append(instancesConfig, instanceConf)
		}
	}
	return instancesConfig, nil
}

func GetAllImageInUsed(path string) ([]string, error) {
	confs, err := GetAllInstancesConfig(path)
	if err != nil {
		return nil, err
	}

	images := make([]string, 0)
	for _, conf := range confs {
		images = append(images, filepath.Base(conf.ImgPath))
	}
	return images, nil
}
