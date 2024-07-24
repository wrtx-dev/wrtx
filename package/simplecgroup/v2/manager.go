package cgroupv2

import (
	"fmt"
	"os"
	"runtime"
)

const (
	DefaultCgroupV2Path = "/sys/fs/cgroup/"
)

type CgroupV2 struct {
	Path string
}

func New(rootPath, groupName string) (*CgroupV2, error) {
	if rootPath == "" {
		rootPath = DefaultCgroupV2Path
	}
	dirPath := fmt.Sprintf("%s/%s", rootPath, groupName)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return nil, err
	}
	return &CgroupV2{
		Path: dirPath,
	}, nil
}

func (c *CgroupV2) setCpuNum(cpuNum int) error {
	if c.Path == "" {
		return fmt.Errorf("control cgroup root path wasn't init")
	}
	cPath := fmt.Sprintf("%s/cpuset.cpus", c.Path)
	num := cpuNum
	if num > runtime.NumCPU() {
		num = runtime.NumCPU()
	}

	cFD, err := os.OpenFile(cPath, os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer cFD.Close()
	var infoLine string
	if num > 1 {
		infoLine = fmt.Sprintf("0-%d", num-1)
	} else {
		infoLine = fmt.Sprintf("%d", num-1)
	}
	_, err = cFD.Write([]byte(infoLine))
	return err
}

func (c *CgroupV2) setCpuMax(percent int) error {
	if c.Path == "" {
		return fmt.Errorf("control cgroup root path wasn't init")
	}
	cPath := fmt.Sprintf("%s/%s", c.Path, "cpu.max")
	if percent > 100 {
		percent = 100
	}
	cpuMax := 1000 * percent
	cFD, err := os.OpenFile(cPath, os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer cFD.Close()
	infoLine := fmt.Sprintf("%d %d", cpuMax, 100000)
	// fmt.Println("infoLine:", infoLine)
	_, err = cFD.Write([]byte(infoLine))
	return err
}

func (c *CgroupV2) setMemMax(maxMem int) error {
	memLimit := maxMem * 1024 * 1024
	if c.Path == "" {
		return fmt.Errorf("control cgroup root path wasn't init")
	}
	memPath := fmt.Sprintf("%s/%s", c.Path, "memory.max")
	mFD, err := os.OpenFile(memPath, os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer mFD.Close()
	infoLine := fmt.Sprintf("%d", memLimit)
	_, err = mFD.Write([]byte(infoLine))
	return err
}

func (c *CgroupV2) SetCPUMemLimit(cpus, period, mem int) error {
	if period > 0 {
		err := c.setCpuMax(period)
		if err != nil {
			return err
		}
	}
	if mem > 0 {
		err := c.setMemMax(mem)
		if err != nil {
			return err
		}
	}
	if cpus > 0 {
		err := c.setCpuNum(cpus)
		return err
	}
	return nil
}

func (c *CgroupV2) AddProcesssors(pids []int) error {
	if c.Path == "" {
		return fmt.Errorf("control cgroup root path wasn't init")
	}
	pPath := fmt.Sprintf("%s/cgroup.procs", c.Path)
	pFD, err := os.OpenFile(pPath, os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer pFD.Close()
	for _, pid := range pids {
		pLine := fmt.Sprintf("%d\n", pid)
		_, err := pFD.Write([]byte(pLine))
		if err != nil {
			return fmt.Errorf("write pid: %d to %s err: %v", pid, pPath, err)
		}
	}
	return nil
}
