package instances

import (
	"fmt"
	"os"
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
