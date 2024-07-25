package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
	"wrtx/internal/config"
	"wrtx/internal/instances"
	_ "wrtx/internal/terminate"

	"github.com/sirupsen/logrus"
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
	globalConfLoaded := false

	var instanceName string
	if len(ctx.Args().Slice()) == 0 {
		instanceName = "openwrt"
	} else {
		instanceName = ctx.Args().First()
	}
	if globalConfPath == "" {
		globalConfPath = config.DefaultConfPath
	}
	globalConfig := config.NewGlobalConf()
	if err := globalConfig.Load(globalConfPath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("load global config error: %v", err)
		}
	} else {
		globalConfLoaded = true
	}
	if !globalConfLoaded {
		return fmt.Errorf("global config not found: %s", globalConfPath)
	}

	conf := config.NewConf()
	instanceConfig := fmt.Sprintf("%s/%s/config.json", globalConfig.InstancesPath, instanceName)
	if err := conf.Load(instanceConfig); err != nil {
		return fmt.Errorf("load instance config error: %v", err)
	}
	if err := conf.Load(instanceConfig); err != nil {
		return fmt.Errorf("load instance config error: %v", err)
	}
	status := instances.NewStatus()
	if err := status.Load(conf.StatusFile); err != nil {
		return fmt.Errorf("load status error: %v", err)
	}

	pid := status.PID
	err := terminateCmd(pid)
	if err != nil {
		return fmt.Errorf("termiate openwrt error:%v", err)
	}

	nspath := fmt.Sprintf("%s/ns", conf.Instances)
	releaseNamespace(nspath)
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

	// if conf.CgroupPath != "" {
	// 	rmCgroupSubDirs(conf.CgroupPath)
	// }
	if conf.ResLimit {
		rmCgroupSubDirs(status.CgroupPath)
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
