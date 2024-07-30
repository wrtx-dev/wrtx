package agent

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"wrtx/internal/config"
)

func StartWrtxInstance(confFile string) error {
	_, err := LoadInstanceConfig(confFile)
	if err != nil {
		return err
	}
	cmd := exec.Cmd{
		Path:        "/proc/self/exe",
		Args:        []string{"wrtxd", "agent", "--config", confFile},
		SysProcAttr: &syscall.SysProcAttr{Setsid: true},
	}
	return cmd.Start()
}

func LoadInstanceConfig(confPath string) (*config.WrtxConfig, error) {
	conf := config.WrtxConfig{}
	if _, err := os.Stat(confPath); err != nil {
		if os.IsNotExist(err) {
			return &conf, err
		}
		return &conf, fmt.Errorf("load config file: %s error: %v", confPath, err)
	}
	if err := conf.Load(confPath); err != nil {
		return nil, fmt.Errorf("load config file: %s error: %v", confPath, err)
	}
	return &conf, nil
}
