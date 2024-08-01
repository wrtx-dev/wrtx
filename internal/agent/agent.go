package agent

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"wrtx/internal/config"
)

func StartWrtxInstance(globalConf, confFile string) error {
	_, err := LoadInstanceConfig(confFile)
	if err != nil {
		return err
	}
	args := []string{"wrtxd"}
	if globalConf != "" {
		args = append(args, "--conf", globalConf)
	}
	args = append(args, "agent", "--iconf", confFile)

	cmd := exec.Cmd{
		Path: "/proc/self/exe",
		Args: args,
		SysProcAttr: &syscall.SysProcAttr{
			Setsid: true,
		},
		// Stdout: os.Stdout,
		// Stderr: os.Stderr,
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
