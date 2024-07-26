package utils

import (
	"fmt"
	"os"
	"wrtx/internal/config"
	"wrtx/internal/instances"
)

func GetInstancesConfig(globalPath, instanceName string) (*config.WrtxConfig, error) {
	globalConfPath := globalPath
	globalConfLoaded := false

	if instanceName == "" {
		instanceName = "openwrt"
	}

	if globalConfPath == "" {
		globalConfPath = config.DefaultConfPath
	}
	globalConfig := config.NewGlobalConf()
	if err := globalConfig.Load(globalConfPath); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("load global config error: %v", err)
		}
	} else {
		globalConfLoaded = true
	}
	if !globalConfLoaded {
		return nil, fmt.Errorf("global config not found: %s", globalConfPath)
	}

	conf := config.NewConf()
	instanceConfig := fmt.Sprintf("%s/%s/config.json", globalConfig.InstancesPath, instanceName)
	if err := conf.Load(instanceConfig); err != nil {
		return nil, fmt.Errorf("load instance config error: %v", err)
	}
	if err := conf.Load(instanceConfig); err != nil {
		return nil, fmt.Errorf("load instance config error: %v", err)
	}
	return conf, nil
}

func GetStatusByWrtxConfig(wrtxConfig *config.WrtxConfig) (*instances.Status, error) {
	if wrtxConfig == nil {
		return nil, fmt.Errorf("wrtxConfig is nil or empty")
	}
	status := instances.NewStatus()
	if err := status.Load(wrtxConfig.StatusFile); err != nil {
		return nil, fmt.Errorf("load status error: %v", err)
	}
	return status, nil
}

func GetInstancesPid(globalPth string, instanceName string) (int, error) {
	conf, err := GetInstancesConfig(globalPth, instanceName)
	if err != nil {
		return 0, err
	}
	status, err := GetStatusByWrtxConfig(conf)
	if err != nil {
		return 0, err
	}
	return status.PID, nil
}
