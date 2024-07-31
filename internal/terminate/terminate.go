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
	if err := stopContainer(); err != nil {
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
		return fmt.Errorf("convert %s to int error: %v", pidStr, err)
	}
	if checkPidName(1, "/sbin/procd") {
		syscall.Kill(pid, syscall.SIGTERM)
		for i := 0; i < 120; i++ {
			fmt.Printf("\rwaiting pid: %d exit.", pid)
			time.Sleep(1 * time.Second)
			if !checkPidExist(strconv.Itoa(pid)) {
				fmt.Println("\npid exited")
				return nil
			}
		}
		fmt.Println("\ntimeout waiting for pid to exit")
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
		return false
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
		return
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
			}
		}
	}
}

func umount() {
	if err := syscall.Unmount("/proc", 0); err != nil {
		fmt.Println("unmount /proc error:", err)
	}
	if err := syscall.Unmount("/sys", 0); err != nil {
		fmt.Println("unmount /sys error:", err)
	}
}
