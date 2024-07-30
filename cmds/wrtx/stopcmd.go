package main

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"wrtx/internal/config"
	_ "wrtx/internal/terminate"
	"wrtx/internal/utils"

	"github.com/urfave/cli/v2"
)

var stopCmd = cli.Command{
	Name:      "stop",
	Usage:     "stop openwrt in namespace",
	ArgsUsage: "instance_name",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "conf",
			Usage: "conf file's path, default: " + config.DefaultConfPath,
		},
	},
	Action: stopWrt,
}

func stopWrt(ctx *cli.Context) error {
	globalConfPath := ctx.String("conf")

	var instanceName string
	if len(ctx.Args().Slice()) == 0 {
		instanceName = "openwrt"
	} else {
		instanceName = ctx.Args().First()
	}
	conf, err := utils.GetInstancesConfig(globalConfPath, instanceName)
	if err != nil {
		return fmt.Errorf("get instance config error:%v", err)
	}

	status, err := utils.GetStatusByWrtxConfig(conf)
	if err != nil {
		return fmt.Errorf("get status error:%v", err)
	}

	pid := status.AgentPid
	err = syscall.Kill(pid, syscall.SIGTERM)
	if err != nil {
		return fmt.Errorf("kill agent error:%v", err)
	}

	fmt.Printf("wait agent eixted, pid: %d", pid)
	for range 120 {
		if checkPidExist(pid) {
			fmt.Printf(".")
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}

func checkPidExist(pid int) bool {
	pidDir := fmt.Sprintf("/proc/%d", pid)
	if _, err := os.Stat(pidDir); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
