package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"wrtx/internal/config"
	fsMount "wrtx/package/mount"
	"wrtx/package/network"
	"wrtx/package/simplecgroup"
	cgroupv2 "wrtx/package/simplecgroup/v2"

	"github.com/pkg/errors"
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
		&cli.StringFlag{
			Name:  "conf",
			Usage: "conf file's path, default: " + config.DefaultConfPath,
		},
		&cli.StringFlag{
			Name:  "nictype",
			Usage: "virtual nic dev's type, Only Support: ipvlan macvlan eg.: ipvlan",
		},
		&cli.StringFlag{
			Name:  "image",
			Usage: "image name which will run",
		},
	},
	Action: runWrt,
}

func runWrt(ctx *cli.Context) error {
	conf := config.NewConf()
	confLoaded := true
	wrtxConfPath := ctx.String("conf")
	imgName := ctx.String("image")
	if imgName == "" {
		imgName = config.DefaultImageName
	}
	if wrtxConfPath == "" {
		wrtxConfPath = config.DefaultConfPath
	}
	if err := conf.Load(wrtxConfPath); err != nil {
		confLoaded = false
	}

	workDir := conf.WorkDir
	if workDir == "" {
		workDir = config.DefaultInstancePath + "/workDir"
		conf.WorkDir = workDir
	}

	upperDir := conf.UpperDir
	if upperDir == "" {
		upperDir = config.DefaultInstancePath + "/upperDir"
		conf.UpperDir = upperDir
	}

	mergeDir := conf.MergeDir
	if mergeDir == "" {
		mergeDir = config.DefaultInstancePath + "/mergeDir"
		conf.MergeDir = mergeDir
	}
	if conf.RootDir == "" {
		if imgName != "" {
			conf.RootDir = config.DefaultImagePath + "/" + imgName
		} else {
			conf.RootDir = config.DefaultRootPath
		}
	}
	if err := mountRootfs(*conf); err != nil {
		return errors.WithMessage(err, "mount rootfs error")
	}
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	phy := ctx.String("phy")
	netDevName := ctx.String("ethname")

	rootDir := conf.MergeDir

	if confLoaded {
		if len(conf.PhyDevName) > 0 && phy == "" {
			phy = conf.PhyDevName
		}
		if len(conf.NetDevName) > 0 && netDevName == "" {
			netDevName = conf.NetDevName
		}
	}
	if phy == "" {
		phy = "eth0"
	}
	if netDevName == "" {
		netDevName = "wrtx_veth"
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
	conf.Dumps(config.DefaultConfPath)
	return nil
}

func mountRootfs(conf config.WrtxConfig) error {
	var wrtxPaths = [...]string{conf.WorkDir, conf.UpperDir, conf.MergeDir}
	for _, wrtxPath := range wrtxPaths {
		if stat, err := os.Stat(wrtxPath); err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("check path: %s error: %v", wrtxPath, err)
			}
			if err = os.MkdirAll(wrtxPath, os.ModePerm); err != nil {
				return errors.WithMessagef(err, "create dir %s error", wrtxPath)
			}
		} else {
			if !stat.IsDir() {
				return fmt.Errorf("check path: %s exist, but it's not a dir", wrtxPath)
			}
		}
	}
	return fsMount.MountOverlayFs(conf.WorkDir, conf.UpperDir, conf.RootDir, conf.MergeDir)
}
