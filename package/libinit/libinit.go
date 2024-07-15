package libinit

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"golang.org/x/sys/unix"
)

func Init() {
	rootfs := os.Getenv("_ROOTDIR")
	ctrlFdstr := os.Getenv("_CTRL_PIPE")
	if ctrlFdstr == "" {
		os.Exit(-1)
	}
	if fd, err := strconv.Atoi(ctrlFdstr); err == nil {
		ctrlPipe := os.NewFile(uintptr(fd), "ctrlpipe")
		msg := make([]byte, 4096)
		n, err := ctrlPipe.Read(msg)
		if err != nil {
			os.Exit(-1)
		}
		if string(msg[:n]) != "continue" {
			os.Exit(-1)
		}
	}
	err := prepareRoot(rootfs)
	if err != nil {
		os.Exit(-1)
	}
	if err := unix.Chdir(rootfs); err != nil {
		os.Exit(-1)
	}
	if err := pivotRoot(rootfs); err != nil {
		os.Exit(-1)
	}
	initBin := "/sbin/init"
	path, err := exec.LookPath(initBin)
	if err != nil {
		fmt.Println("find path err:", err)
		os.Exit(-1)
	}
	fmt.Println("start run /sbin/init")
	err = syscall.Exec(path, nil, nil)
	if err != nil {
		fmt.Println("run err:", err)
	}
}
