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
	DefalutLogDir         = baseWrtxPath + "/log"
	DefaultLogPath        = DefalutLogDir + "/wrtx.log"
)

type GlobalConfig struct {
	InstancesPath       string `json:"instances_path"`
	DefaultImageName    string `json:"default_image_name"`
	ImagePath           string `json:"default_image_path"`
	DefaultInstanceName string `json:"default_instance_name"`
	LogPath             string `json:"log_path"`
}
type WrtxConfig struct {
	InstanceName  string            `json:"instance_name"`
	ResLimit      bool              `json:"res_limit"`
	HardwareAddr  string            `json:"hardware_addr"`
	NetDevName    string            `json:"net_dev_name"`
	PhyDevName    string            `json:"phy_dev_name"`
	ImgPath       string            `json:"image_path"`
	Instances     string            `json:"instances"`
	WorkDir       string            `json:"work_dir"`
	MergeDir      string            `json:"merge_dir"`
	UpperDir      string            `json:"upper_dir"`
	NetConfigFile string            `json:"net_config_file"`
	Cpus          int               `json:"cpus"`
	Mem           int               `json:"mem"`
	Period        int               `json:"period"`
	ConfigNetwork bool              `json:"config_network"`
	NicType       string            `json:"nic_type"`
	MountMap      map[string]string `json:"mount_map"`
	StatusFile    string            `json:"status_file"`
	AlwaysRestart bool              `json:"always_restart"`
}

func GetNsFilesName() []string {
	return []string{"ipc", "net", "pid", "uts", "cgroup"}
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
		DefaultImageName:    DefaultImageName,
		ImagePath:           DefaultImagePath,
		DefaultInstanceName: "openwrt",
		LogPath:             DefaultLogPath,
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

func GetGlobalConfig(globalPath string) (*GlobalConfig, error) {
	if globalPath == "" {
		globalPath = DefaultConfPath
		if _, err := os.Stat(globalPath); os.IsNotExist(err) {
			gConf := DefaultGlobalConf()
			if err := createDefaultDirs(); err != nil {
				return nil, fmt.Errorf("create default dirs error: %v", err)
			}
			if err := gConf.Dumps(globalPath); err != nil {
				return nil, fmt.Errorf("save global config error: %v", err)
			}
		}
	}
	globalConf := NewGlobalConf()
	if err := globalConf.Load(globalPath); err != nil {
		return nil, fmt.Errorf("load global config error: %v", err)
	}
	return globalConf, nil
}

func createDefaultDirs() error {
	paths := []string{DefaultImagePath, DefaultInstancePath, DefaultRunDir, DefalutLogDir}
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				return fmt.Errorf("create dir %s error: %v", path, err)
			}
		}
	}
	return nil
}
