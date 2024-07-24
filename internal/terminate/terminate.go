package terminate

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/vishvananda/netlink"
)

func init() {
	if len(os.Args) != 2 || os.Args[1] != "terminate" {
		return
	}
	err := stopContainer()
	if err != nil {
		fmt.Println("stop openwrt error:", err)
		os.Exit(-1)
	}
	releaseNetworkDev()
	umount()
	os.Exit(0)
}

func stopContainer() error {
	pidStr := os.Getenv("NSPID")
	if pidStr == "" {
		return fmt.Errorf("get env NSPID error")
	}
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("covert %s to int error", string(pidStr))
	}
	if checkPidName(1, "/sbin/procd") {
		syscall.Kill(pid, syscall.SIGTERM)
		flag := false
		for range 120 {
			if !flag {
				fmt.Printf("waiting pid: %d exit.", pid)
				flag = true
			} else {
				fmt.Print(".")
			}
			time.Sleep(1 * time.Second)
			if checkPidExist("1") {
				continue
			}
			fmt.Println()
			break
		}
	}
	return nil
}

func checkPidName(pid int, name string) bool {
	pidPath := fmt.Sprintf("/proc/%d/cmdline", pid)
	cmdlines, err := os.ReadFile(pidPath)
	if err != nil {
		fmt.Println("get cmd line error:", err)
		return false
	}
	lines := bytes.Split(cmdlines, []byte{0})
	if len(lines) < 1 {
		fmt.Printf("get cmd name error, /proc/%d/cmdline did not have enough data", pid)
	}
	return string(lines[0]) == name
}
func checkPidExist(pid string) bool {
	pidDir := fmt.Sprintf("/proc/%s", pid)
	if _, err := os.Stat(pidDir); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func releaseNetworkDev() {
	links, err := netlink.LinkList()
	if err != nil {
		fmt.Println("get links error:", err)
	}
	for _, link := range links {
		linkType := strings.ToLower(link.Type())
		if linkType == "ipvlan" || linkType == "macvlan" {
			if err := netlink.LinkSetDown(link); err != nil {
				fmt.Println("set link:", link.Attrs().Name, "down error:", err)
				continue
			}
			if err := netlink.LinkDel(link); err != nil {
				fmt.Println("delete link:", link.Attrs().Name, "err:", err)
				continue
			}
		}

	}
}

func umount() {
	syscall.Unmount("/proc", 0)
	syscall.Unmount("/sys", 0)
}
