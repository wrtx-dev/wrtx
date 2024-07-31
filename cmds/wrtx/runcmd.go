package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"wrtx/internal/agent"
	"wrtx/internal/config"
	"wrtx/internal/instances"
	"wrtx/internal/netconf"
	"wrtx/package/network"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var runcmd = cli.Command{
	Name:      "run",
	Usage:     "run openwrt img",
	ArgsUsage: " image_name",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "phy",
			Usage: "phy net dev name",
		},
		&cli.StringFlag{
			Name:  "veth",
			Usage: "virtual net dev name",
		},
		&cli.StringFlag{
			Name:  "conf",
			Usage: "conf file's path, default: " + config.DefaultConfPath,
		},
		&cli.StringFlag{
			Name:  "vtype",
			Usage: "virtual nic dev's type, Only Support: ipvlan macvlan, default macvla, eg.: ipvlan",
		},
		&cli.StringFlag{
			Name:  "image",
			Usage: "image name which will run",
		},
		&cli.StringFlag{
			Name:  "period",
			Usage: "set cpu period use percentage,eg.: 50",
		},
		&cli.StringFlag{
			Name:  "cpus",
			Usage: "how many cpu cores can openwrt use,eg.: 1",
		},
		&cli.StringFlag{
			Name:  "mem",
			Usage: "how many MB can openwrt use,eg.: 256",
		},
		&cli.StringFlag{
			Name:  "ip",
			Usage: "set ip address, eg.: 192.168.64.6",
		},
		&cli.StringFlag{
			Name:  "mask",
			Usage: "set netmask, eg.: 255.255.255.0",
		},
		&cli.StringFlag{
			Name:  "gateway",
			Usage: "set gateway, eg.: 192.168.64.1",
		},
		&cli.StringFlag{
			Name:  "dns",
			Usage: "set dns, eg.: 8.8.8.8",
		},
		&cli.BoolFlag{
			Name:  "always",
			Usage: "if need always restart instance",
			Value: false,
		},
	},
	Action: runWrt,
}

func runWrt(ctx *cli.Context) error {
	globalConfPath := ctx.String("conf")
	globalConfLoaded := false
	imgName := ctx.String("image")
	nictype := ctx.String("vtype")
	cpuCores := ctx.String("cpus")
	cpuPeriod := ctx.String("period")
	mem := ctx.String("mem")
	phy := ctx.String("phy")
	netDevName := ctx.String("veth")

	always := ctx.Bool("always")

	ip := ctx.String("ip")
	mask := ctx.String("mask")
	gateway := ctx.String("gateway")
	dns := ctx.String("dns")

	configNetwork, err := checkNetConfig(ip, mask, gateway, dns)
	if err != nil {
		return err
	}

	name := ctx.Args().First()
	if name == "" {
		name = "openwrt"
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
		globalConfig = config.DefaultGlobalConf()
	}
	globalConfig.Dumps(globalConfPath)
	instancePath := fmt.Sprintf("%s/%s", globalConfig.InstancesPath, name)
	if _, err := os.Stat(instancePath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("instance path %s already exist, stat error: %v", instancePath, err)
		}
	} else {
		return fmt.Errorf("instance path %s already exist", instancePath)
	}
	confPath := fmt.Sprintf("%s/config.json", instancePath)
	conf := config.NewConf()
	conf.Instances = instancePath
	conf.AlwaysRestart = always
	period := 0
	memory := 0
	cpus := 0
	rLimit := false
	if cpuCores != "" || cpuPeriod != "" || mem != "" {
		rLimit = true
	}

	if cpuCores != "" {
		var err error
		cpus, err = strconv.Atoi(cpuCores)
		if err != nil {
			return fmt.Errorf("parse cpus error: %v", err)
		}
		if cpus < 1 {
			return fmt.Errorf("invaild vaule for cpus: %d", cpus)
		}
		if cpus > runtime.NumCPU() {
			cpus = runtime.NumCPU()
		}
	}
	if cpuPeriod != "" {
		var err error
		period, err = strconv.Atoi(cpuPeriod)
		if err != nil {
			return fmt.Errorf("parse cpu period error: %v", err)
		}
		if period < 0 {
			return fmt.Errorf("invaild cpu period: %d", period)
		}
		if mem != "" {
			var err error
			if memory, err = strconv.Atoi(mem); err != nil {
				return fmt.Errorf("parse mem error: %v", err)
			}
			if memory < 0 {
				return fmt.Errorf("invail mem: %d", memory)
			}
		}

		if nictype == "" {
			nictype = "macvlan"
		}
	}

	if mem != "" {
		var err error
		if memory, err = strconv.Atoi(mem); err != nil {
			return fmt.Errorf("parse mem error: %v", err)
		}
		if memory < 0 {
			return fmt.Errorf("invail mem: %d", memory)
		}
	}
	if imgName == "" {
		imgName = config.DefaultImageName
	}

	existed, err := instances.CheckImages(globalConfig, imgName)
	if err != nil {
		return fmt.Errorf("check images error: %v", err)
	}
	if !existed {
		return fmt.Errorf("image %s not exist", imgName)
	}
	if nictype == "" {
		nictype = "macvlan"
	}
	conf.ResLimit = rLimit
	conf.Cpus = cpus
	conf.Period = period
	conf.Mem = memory

	conf.WorkDir = conf.Instances + "/work"
	conf.UpperDir = conf.Instances + "/upper"
	conf.MergeDir = conf.Instances + "/merge"
	conf.NicType = nictype
	conf.ImgPath = fmt.Sprintf("%s/%s", globalConfig.ImagePath, imgName)
	if err := createInstanceDir(conf); err != nil {
		return err
	}
	if phy == "" {
		phy = "eth0"
	}
	if netDevName == "" {
		netDevName = "wrtx_veth"
	}
	conf.NetDevName = netDevName
	conf.PhyDevName = phy
	conf.HardwareAddr = network.GenRandMac()
	if configNetwork {
		netconfPath := fmt.Sprintf("%s/network", conf.Instances)
		if _, err := os.Stat(netconfPath); err != nil {
			if !os.IsNotExist(err) {

				return fmt.Errorf("config network file %s stat error: %v", netconfPath, err)
			}
		}
		fp, err := os.OpenFile(netconfPath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("open network file error: %v", err)
		}
		defer fp.Close()
		netconfs := netconf.NewWrtxNetConfig(netDevName, "192.168.64.6", "255.255.255.0", "192.168.64.1", "8.8.8.8")

		if err := netconf.GenerateNetConfig(netconfs, fp); err != nil {
			fmt.Printf("config network err:\n%v\n\n\n", err)
			return fmt.Errorf("config network error: %v", err)
		}

		conf.NetConfigFile = netconfPath
	}
	conf.ConfigNetwork = configNetwork
	conf.StatusFile = fmt.Sprintf("%s/status.json", conf.Instances)
	if err := conf.Dumps(confPath); err != nil {
		return fmt.Errorf("create new instance error: %v", err)
	}
	return agent.StartWrtxInstance(confPath)
	// return nil
}

func createInstanceDir(conf *config.WrtxConfig) error {
	paths := []string{conf.Instances, conf.WorkDir, conf.UpperDir, conf.MergeDir}
	for _, path := range paths {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "create dir: %s error", path)
		}
	}
	return nil
}

func checkNetConfig(ip, mask, gateway, dns string) (bool, error) {
	netconfs := [...]string{ip, mask, gateway, dns}
	idx := [...]string{"ip", "mask", "gateway", "dns"}
	needconfig := false
	for _, conf := range netconfs {
		if conf != "" {
			needconfig = true
			break
		}
	}
	if needconfig {
		for i, conf := range netconfs {
			if conf == "" {
				return false, fmt.Errorf("config network error: %s is empty", idx[i])
			}
		}
	}
	return needconfig, nil
}
