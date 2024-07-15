package simplecgroup

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func readMountInfo() ([]byte, error) {
	mi, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return nil, err
	}
	buf, err := io.ReadAll(mi)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func parseMountInfo(buf []byte) (map[string][]string, error) {
	lines := strings.Split(string(buf), "\n")
	info := make(map[string][]string)
	if len(lines) == 0 {
		return info, nil
	}
	for i, line := range lines {
		if len(line) == 0 {
			continue
		}
		parts := strings.Split(line, " ")
		// fmt.Println("idx:", i, ",", line, " len:", len(line))
		if len(parts) < 10 {

			return info, fmt.Errorf("parse line: %s error,idx: %d", line, i)
		}
		if _, ok := info[parts[8]]; !ok {
			info[parts[8]] = make([]string, 0)
		}
		info[parts[8]] = append(info[parts[8]], parts[4])
	}
	return info, nil
}

const (
	CGTypeNone  = 0
	CGTypeOne   = 0x1 << 1
	CGTypeTwo   = 0x1 << 2
	CGTypeHyper = CGTypeOne | CGTypeTwo
)

func GetCgroupType() (int, error) {
	cgType := 0
	if buf, err := readMountInfo(); err == nil {
		if info, err := parseMountInfo(buf); err == nil {
			if _, OK := info["cgroup"]; OK {
				cgType |= CGTypeOne
			}
			if _, OK := info["cgroup2"]; OK {
				cgType |= CGTypeTwo
			}
		} else {
			return cgType, err
		}

	} else {
		return cgType, err
	}
	return cgType, nil
}
