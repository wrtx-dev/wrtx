package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"wrtx/internal/config"
	"wrtx/package/network"
	"wrtx/package/simplecgroup"
	cgroupv2 "wrtx/package/simplecgroup/v2"

	"github.com/urfave/cli/v2"
)

var runcmd = cli.Command{
	Name:  "run",
	Usage: "run openwrt img",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "phy",
			Usage: "parent net dev",
		},
		&cli.StringFlag{
			Name:  "ethname",
			Usage: "new eth dev name",
		},
	},
	Action: runWrt,
}

func runWrt(ctx *cli.Context) error {
	conf := config.NewConf()
	confLoaded := true
	if err := conf.Load("/usr/local/wrtx/conf/conf.json"); err != nil {
		confLoaded = false
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	phy := ctx.String("phy")
	netDevName := ctx.String("ethname")
	var rootDir string
	if conf.RootDir == "" {
		rootDir = config.DefaultImagePath
	} else {
		rootDir = conf.RootDir
	}

	if confLoaded {
		if len(conf.PhyDevName) > 0 && phy == "" {
			phy = conf.PhyDevName
		}
		if len(conf.NetDevName) > 0 && netDevName == "" {
			netDevName = conf.NetDevName
		}
	}
	if phy == "" {
		phy = "veth_wrtx"
	}
	if netDevName == "" {
		netDevName = "eth0"
	}

	r, err := syscall.Socketpair(syscall.AF_LOCAL, syscall.SOCK_STREAM, 0)
	if err != nil {
		return fmt.Errorf("create init pipe error: %v", err)
	}
	cr, cw, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("create ctrl pipe error: %v", err)
	}
	cmd.ExtraFiles = append(cmd.ExtraFiles, os.NewFile(uintptr(r[1]), "init_pipe"))

	cmd.Env = []string{fmt.Sprintf("_INIT_PIPE=%d", stdioCount+len(cmd.ExtraFiles)-1)}

	cmd.ExtraFiles = append(cmd.ExtraFiles, cr)
	cmd.Env = append(cmd.Env, fmt.Sprintf("_CTRL_PIPE=%d", stdioCount+len(cmd.ExtraFiles)-1))

	cmd.Env = append(cmd.Env, fmt.Sprintf("_ROOTDIR=%s", rootDir))
	fmt.Println("cmd's env: ", cmd.Env)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start cmd err: %s", err)
	} else {
		fmt.Println("create new proc's pid:", cmd.Process.Pid)
	}
	cg, err := simplecgroup.GetCgroupType()
	fmt.Println("cgroup type:", cg)
	if err != nil {
		syscall.Kill(cmd.Process.Pid, syscall.SIGKILL)
		return err
	}
	if cg&simplecgroup.CGTypeTwo != 0 {
		mgr, err := cgroupv2.New("", "myopenwrt")
		if err != nil {
			syscall.Kill(cmd.Process.Pid, syscall.SIGKILL)
			return err
		}
		err = mgr.SetCPUMemLimit(50, 268435456)
		if err != nil {
			syscall.Kill(cmd.Process.Pid, syscall.SIGKILL)
			return err
		}
		err = mgr.AddProcesssors([]int{cmd.Process.Pid})
		if err != nil {
			syscall.Kill(cmd.Process.Pid, syscall.SIGKILL)
			return err
		}
	}
	msg := make([]byte, 4096)
	rr := os.NewFile(uintptr(r[0]), "__init_pipe")
	_, err = rr.Write([]byte("continue"))
	if err != nil {
		syscall.Kill(cmd.Process.Pid, syscall.SIGKILL)
		return fmt.Errorf("send msg to child err: %v", err)
	}
	n, err := rr.Read(msg)
	if err != nil && err != io.EOF {
		return fmt.Errorf("read msg from sub process error: %v", err)
	}
	fmt.Println("read json msg:", string(msg[:n]))
	jsMsg := &pidMsg{}
	err = json.Unmarshal(msg[:n], jsMsg)
	if err != nil {
		return fmt.Errorf("get sub pid error: %v", err)
	}
	wpid, err := syscall.Wait4(jsMsg.ChildPID, nil, 0, nil)
	if err != nil {
		return fmt.Errorf("wait error: %v", err)
	}
	fmt.Println("child pid:", jsMsg.ChildPID, "exit, wpid:", wpid)
	_, err = network.NewIPvlanDev(netDevName, phy)
	if err != nil {
		syscall.Kill(jsMsg.GrandChildPid, syscall.SIGKILL)
		return err
	}
	err = network.AddDevToNamespaceByPID(netDevName, phy, jsMsg.GrandChildPid)
	if err != nil {
		syscall.Kill(jsMsg.GrandChildPid, syscall.SIGKILL)
		return err
	}

	cw.Write([]byte("continue"))
	return nil
}
