package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/errors"
)

const (
	baseWrtxPath          = "/usr/local/wrtx"
	DefaultImagePath      = baseWrtxPath + "/images"
	DefaultConfPath       = baseWrtxPath + "/conf/conf.json"
	DefaultInstancePath   = baseWrtxPath + "/instances"
	DefaultRootPath       = DefaultImagePath + "/openwrt"
	DefaultImageName      = "openwrt"
	DefaultRunDir         = baseWrtxPath + "/run"
	DefaultWrtxRunPidFile = DefaultRunDir + "/pid"
)

type GlobalConfig struct {
	InstancesPath       string `json:"instances_path"`
	ImageName           string `json:"default_image_name"`
	ImagePath           string `json:"default_image_path"`
	DefaultInstanceName string `json:"default_instance_name"`
}
type WrtxConfig struct {
	ResLimit       bool              `json:"res_limit"`
	NetDevName     string            `json:"net_dev_name"`
	PhyDevName     string            `json:"phy_dev_name"`
	ImgPath        string            `json:"image_path"`
	Instances      string            `json:"instances"`
	WorkDir        string            `json:"work_dir"`
	MergeDir       string            `json:"merge_dir"`
	UpperDir       string            `json:"upper_dir"`
	VirtualNicType string            `json:"virtual_nic_type"`
	CgroupPath     string            `json:"cgroup_path"`
	NetConfigFile  string            `json:"net_config_file"`
	Cpus           int               `json:"cpus"`
	Mem            int               `json:"mem"`
	Period         int               `json:"period"`
	ConfigNetwork  bool              `json:"config_network"`
	NicType        string            `json:"nic_type"`
	MountMap       map[string]string `json:"mount_map"`
}

func (gc *GlobalConfig) Load(path string) error {
	jsonStr, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return errors.Wrapf(json.Unmarshal(jsonStr, gc), "load global config from %s error", path)
}

func (gc *GlobalConfig) Dumps(path string) error {
	fp, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(os.ModePerm))
	if err != nil {
		return fmt.Errorf("open file %s err: %v", path, err)
	}
	jsonStr, err := json.Marshal(gc)
	if err != nil {
		return fmt.Errorf("marshal conf json str error: %v", err)
	}
	var formatStr bytes.Buffer
	if err = json.Indent(&formatStr, jsonStr, "", "    "); err != nil {
		return fmt.Errorf("format json str error: %v", err)
	}
	_, err = fp.Write(formatStr.Bytes())
	return errors.Wrapf(err, "write json str to file: %s error", path)
}

func NewGlobalConf() *GlobalConfig {
	return &GlobalConfig{}
}

func DefaultGlobalConf() *GlobalConfig {
	return &GlobalConfig{
		InstancesPath:       DefaultInstancePath,
		ImageName:           DefaultImageName,
		ImagePath:           DefaultImagePath,
		DefaultInstanceName: "openwrt",
	}
}

func (wc *WrtxConfig) Dumps(dst string) error {
	fp, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(os.ModePerm))
	if err != nil {
		return fmt.Errorf("open file %s err: %v", dst, err)
	}
	jsonStr, err := json.Marshal(wc)
	if err != nil {
		return fmt.Errorf("marshal conf json str error: %v", err)
	}
	var formatStr bytes.Buffer
	if err = json.Indent(&formatStr, jsonStr, "", "    "); err != nil {
		return fmt.Errorf("format json str error: %v", err)
	}
	_, err = fp.Write(formatStr.Bytes())
	return errors.Wrapf(err, "write json str to file: %s error", dst)
}

func (wc *WrtxConfig) Load(dst string) error {
	jsonStr, err := os.ReadFile(dst)
	if err != nil {
		return err
	}
	return errors.Wrapf(json.Unmarshal(jsonStr, wc), "load config from %s error", dst)
}

func NewConf() *WrtxConfig {
	return &WrtxConfig{}
}
