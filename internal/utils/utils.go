package utils

import (
	"fmt"
	"wrtx/internal/config"
	"wrtx/internal/instances"
)

func GetInstancesConfig(globalPath, instanceName string) (*config.WrtxConfig, error) {
	globalConfPath := globalPath

	if globalConfPath == "" {
		globalConfPath = config.DefaultConfPath
	}
	globalConfig, err := config.GetGlobalConfig(globalConfPath)
	if err != nil {
		return nil, fmt.Errorf("load global config error: %v", err)
	}

	if instanceName == "" {
		instanceName = globalConfig.DefaultInstanceName
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
