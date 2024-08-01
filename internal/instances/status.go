package instances

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pkg/errors"
)

type Status struct {
	CgroupPath string `json:"cgroup_path"`
	PID        int    `json:"pid"`
	Status     string `json:"status"`
	AgentPid   int    `json:"agent_pid"`
}

func (s *Status) Dump(path string) error {
	fp, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(os.ModePerm))
	if err != nil {
		return fmt.Errorf("open file %s err: %v", path, err)
	}
	defer fp.Close()
	jsonStr, err := json.Marshal(s)
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

func NewStatus() *Status {
	return &Status{}
}

func (s *Status) Load(path string) error {
	fp, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "open file %s error", path)
	}
	defer fp.Close()
	jsonStr, err := io.ReadAll(fp)
	if err != nil {
		return fmt.Errorf("read file %s err: %v", path, err)
	}
	if err = json.Unmarshal(jsonStr, s); err != nil {
		return fmt.Errorf("unmarshal json str error: %v", err)
	}
	return nil
}

func (s *Status) Pid() int {
	return s.PID
}

func (s *Status) Cgroup() string {
	return s.CgroupPath
}

func RemoveCgroupSubDirs(path string) {
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
		RemoveCgroupSubDirs(tpath)
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

func GetInstanceStatus(path string) (*Status, error) {
	status := NewStatus()
	if err := status.Load(path); err != nil {
		return nil, err
	}
	return status, nil
}
