package mount

import (
	"fmt"
	"syscall"
)

const (
	DefaultMountOptions = syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
)

func MountBind(dst, src string) error {
	return syscall.Mount(src, dst, "", syscall.MS_BIND|syscall.MS_REC, "")
}

func MountOverlayFs(workdir, upper, lower, merged string) error {
	return syscall.Mount("overlay", merged, "overlay", 0, fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lower, upper, workdir))
}

func RemountRootPath() error {
	return syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
}

func MountProc(root string) error {
	return syscall.Mount("proc", fmt.Sprintf("%s/proc", root), "proc", uintptr(DefaultMountOptions), "")
}

func MountSys(root string) error {
	return syscall.Mount("sys", fmt.Sprintf("%s/sys", root), "sysfs", uintptr(DefaultMountOptions), "")
}

func MountDev(root string) error {
	err := syscall.Mount("devtmpfs", fmt.Sprintf("%s/dev", root), "devtmpfs", uintptr(DefaultMountOptions), "mode=755")
	if err != nil {
		return err
	}
	// err = syscall.Mount("devpts", "/dev/pts", "devpts", syscall.MS_NOSUID|syscall.MS_NOEXEC, "newinstance,ptmxmode=0666,mode=0620")
	// if err != nil {
	// 	return fmt.Errorf("mount /dev/pts error: %s", err)
	// }
	return nil
}

func MountTMP() error {
	return syscall.Mount("tmpfs", "/tmp", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=777")

}

func Unmount(target string) error {
	return syscall.Unmount(target, 0)
}
