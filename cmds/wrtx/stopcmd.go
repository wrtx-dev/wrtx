package main

import (
	"fmt"
	"os"
	"syscall"
	"time"
	_ "wrtx/internal/terminate"
	"wrtx/internal/utils"

	"github.com/urfave/cli/v2"
)

var stopCmd = cli.Command{
	Name:      "stop",
	Usage:     "stop openwrt in namespace",
	ArgsUsage: "instance_name",
	Action:    stopWrt,
}

func stopWrt(ctx *cli.Context) error {
	globalConfPath := ctx.String("conf")
	instanceName := ctx.Args().First()
	if instanceName == "" {
		instanceName = "openwrt"
	}

	conf, err := utils.GetInstancesConfig(globalConfPath, instanceName)
	if err != nil {
		return fmt.Errorf("get instance config error: %v", err)
	}

	status, err := utils.GetStatusByWrtxConfig(conf)
	if err != nil {
		return fmt.Errorf("get status error: %v", err)
	}

	pid := status.AgentPid
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("kill agent error: %v", err)
	}

	fmt.Printf("wait agent exited, pid: %d", pid)
	for i := 0; i < 120; i++ {
		if checkPidExist(pid) {
			fmt.Print(".")
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	fmt.Println()

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
