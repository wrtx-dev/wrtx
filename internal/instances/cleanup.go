package instances

import (
	"fmt"
	"syscall"
	"wrtx/internal/config"

	"github.com/sirupsen/logrus"
)

func cleanup(conf *config.WrtxConfig) {
	if conf.NetConfigFile != "" {
		if err := syscall.Unmount(fmt.Sprintf("%s/etc/config/network", conf.MergeDir), 0); err != nil {
			logrus.Errorf("unmount path: %s error: %v", fmt.Sprintf("%s/etc/config/network", conf.MergeDir), err)
		}
	}
	if err := syscall.Unmount(conf.MergeDir, 0); err != nil {
		logrus.Errorf("unmount path: %s error: %v", conf.MergeDir, err)
	}
	if err := syscall.Unlink(conf.StatusFile); err != nil {
		logrus.Errorf("unlink file: %s error: %v", conf.StatusFile, err)
	}
}
