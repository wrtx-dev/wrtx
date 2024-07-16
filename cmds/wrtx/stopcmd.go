package main

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
	"time"
	"wrtx/internal/config"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var stopCmd = cli.Command{
	Name:   "stop",
	Usage:  "stop openwrt in namespace",
	Action: stopWrt,
}

func stopWrt(ctx *cli.Context) error {
	conf := config.NewConf()
	conf.Load(config.DefaultConfPath)
	pidfile := config.DefaultWrtxRunPidFile
	buf, err := os.ReadFile(pidfile)
	if err != nil {
		return fmt.Errorf("read file %s error: %v", pidfile, err)
	}
	pid, err := strconv.Atoi(string(buf))
	if err != nil {
		return fmt.Errorf("covert %s to int error", string(buf))
	}
	err = syscall.Kill(pid, syscall.SIGTERM)
	if err != nil {
		return fmt.Errorf("send SIGTERM to %d error: %v", pid, err)
	}
	flag := false
	for range 120 {
		if !flag {
			fmt.Printf("waiting pid: %d exit.", pid)
			flag = true
		} else {
			fmt.Print(".")
		}
		time.Sleep(1 * time.Second)
		if checkPidExist(string(buf)) {
			continue
		}
		fmt.Println()
		if err := syscall.Unlink(config.DefaultWrtxRunPidFile); err != nil {
			logrus.Errorf("unlink file: %s error: %v", config.DefaultWrtxRunPidFile, err)
		}
		if err := syscall.Unmount(conf.MergeDir, 0); err != nil {
			logrus.Errorf("unmount path: %s error: %v", conf.MergeDir, err)
		} else {
			fmt.Printf("unmount path: %s successed", conf.MergeDir)
		}
		break
	}
	return nil
}

func checkPidExist(pid string) bool {
	pidDir := fmt.Sprintf("/proc/%s", pid)
	if _, err := os.Stat(pidDir); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
