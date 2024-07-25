package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
	"wrtx/internal/config"
	_ "wrtx/internal/terminate"

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
	err = terminateCmd(pid)
	if err != nil {
		return fmt.Errorf("termiate openwrt error:%v", err)
	}

	nspath := fmt.Sprintf("%s/ns", config.DefaultInstancePath)
	releaseNamespace(nspath)
	if err := syscall.Unmount(fmt.Sprintf("%s/etc/config/network", conf.MergeDir), 0); err != nil {
		logrus.Errorf("unmount path: %s error: %v", fmt.Sprintf("%s/etc/config/network", conf.MergeDir), err)
	}
	if err := syscall.Unmount(conf.MergeDir, 0); err != nil {
		logrus.Errorf("unmount path: %s error: %v", conf.MergeDir, err)
	}
	if err := syscall.Unlink(config.DefaultWrtxRunPidFile); err != nil {
		logrus.Errorf("unlink file: %s error: %v", config.DefaultWrtxRunPidFile, err)
	}

	if conf.CgroupPath != "" {
		rmCgroupSubDirs(conf.CgroupPath)
	}
	return nil
}

func terminateCmd(pid int) error {
	cmd := exec.Command("/proc/self/exe", "terminate")
	cmd.Env = []string{fmt.Sprintf("NSPID=%d", pid)}
	cmd.Env = append(cmd.Env, fmt.Sprintf("NSLIST=%d", syscall.CLONE_NEWNS|syscall.CLONE_NEWNET|syscall.CLONE_NEWCGROUP))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func releaseNamespace(nspath string) {
	nses, err := os.ReadDir(nspath)
	if err != nil {
		fmt.Println("read dir:", nspath, "error:", err)
	}
	for _, ns := range nses {
		upath := fmt.Sprintf("%s/%s", nspath, ns.Name())
		if err := syscall.Unmount(upath, 0); err != nil {
			fmt.Println("unmount", upath, "error:", err)
		}
	}
}

func rmCgroupSubDirs(path string) {
	items, err := os.ReadDir(path)
	if err != nil {
		fmt.Println("read cgroup's dir:", path, "err:", err)
		return
	}
	for _, item := range items {
		if !item.IsDir() {
			continue
		}
		tpath := fmt.Sprintf("%s/%s", path, item.Name())
		rmCgroupSubDirs(tpath)
	}
	err = os.Remove(path)
	if err != nil {
		time.Sleep(time.Second * 2)
		err = os.Remove(path)
		if err != nil {
			fmt.Println("remove", path, "err:", err)
		}
	}
}
