package libinit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"golang.org/x/sys/unix"
)

const defaultMountFlags = unix.MS_NOEXEC | unix.MS_NOSUID | unix.MS_NODEV

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

	if err := mountSystemDirs(rootfs); err != nil {
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
	// fmt.Println("start run /sbin/init")
	err = syscall.Exec(path, nil, nil)
	if err != nil {
		fmt.Println("run err:", err)
	}
}

func mountSystemDirs(rootfs string) error {
	procDir := filepath.Join(rootfs, "proc")
	sysDir := filepath.Join(rootfs, "sys")
	devDir := filepath.Join(rootfs, "dev")

	for _, dir := range []string{procDir, sysDir, devDir} {
		if err := checkCreateDir(dir); err != nil {
			return err
		}
	}

	if err := mount("proc", procDir, "proc", defaultMountFlags, ""); err != nil {
		return fmt.Errorf("mount proc: %v", err)
	}
	if err := mount("sysfs", sysDir, "sysfs", unix.MS_NOSUID|unix.MS_NOEXEC, ""); err != nil {
		return fmt.Errorf("mount sysfs: %v", err)
	}
	if err := mount("tmpfs", devDir, "tmpfs", unix.MS_NOSUID|unix.MS_STRICTATIME, ""); err != nil {
		return fmt.Errorf("mount dev: %v", err)
	}
	return setupDev(rootfs)
}

func setupDev(rootfs string) error {
	devDir := filepath.Join(rootfs, "dev")
	if _, err := os.Stat(devDir); err != nil {
		return err
	}
	shmDir := filepath.Join(devDir, "shm")
	if err := checkCreateDir(shmDir); err != nil {
		return err
	}
	if err := mount("tmpfs", shmDir, "tmpfs", defaultMountFlags, "mode=1777,size=65536k"); err != nil {
		return fmt.Errorf("mount shm: %v", err)
	}

	ptsDir := filepath.Join(devDir, "pts")
	if err := checkCreateDir(ptsDir); err != nil {
		return err
	}
	if err := mount("devpts", ptsDir, "devpts", defaultMountFlags, "newinstance,ptmxmode=0666,mode=0620,gid=5"); err != nil {
		return fmt.Errorf("mount pts: %v", err)
	}
	mqueueDir := filepath.Join(devDir, "mqueue")
	if err := checkCreateDir(mqueueDir); err != nil {
		return err
	}
	if err := mount("mqueue", mqueueDir, "mqueue", defaultMountFlags, ""); err != nil {
		return fmt.Errorf("mount mqueue: %v", err)
	}

	return nil
}

func checkCreateDir(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("mkdir %s: %v", dir, err)
			}
		} else {
			return fmt.Errorf("stat %s: %v", dir, err)
		}
	}
	return nil
}
