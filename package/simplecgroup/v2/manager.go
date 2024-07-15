package cgroupv2

import (
	"fmt"
	"os"
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
func (c *CgroupV2) setCpuMax(percent int) error {
	if c.Path == "" {
		return fmt.Errorf("control cgroup root path wasn't init")
	}
	cPath := fmt.Sprintf("%s/%s", c.Path, "cpu.max")
	cpuMax := 100000 / 100 * percent
	cFD, err := os.OpenFile(cPath, os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	infoLine := fmt.Sprintf("%d %d", cpuMax, 100000)
	_, err = cFD.Write([]byte(infoLine))
	return err
}

func (c *CgroupV2) setMemMax(maxMem int) error {
	if c.Path == "" {
		return fmt.Errorf("control cgroup root path wasn't init")
	}
	memPath := fmt.Sprintf("%s/%s", c.Path, "memory.max")
	mFD, err := os.OpenFile(memPath, os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	infoLine := fmt.Sprintf("%d", maxMem)
	_, err = mFD.Write([]byte(infoLine))
	return err
}

func (c *CgroupV2) SetCPUMemLimit(cpu int, mem int) error {
	err := c.setCpuMax(cpu)
	if err != nil {
		return err
	}
	err = c.setMemMax(mem)
	return err
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
	for _, pid := range pids {
		pLine := fmt.Sprintf("%d\n", pid)
		_, err := pFD.Write([]byte(pLine))
		if err != nil {
			return fmt.Errorf("write pid: %d to %s err: %v", pid, pPath, err)
		}
	}
	return nil
}
