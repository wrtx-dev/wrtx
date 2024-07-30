package instances

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
	"wrtx/internal/config"
	fsMount "wrtx/package/mount"
	"wrtx/package/network"
	"wrtx/package/simplecgroup"
	cgroupv2 "wrtx/package/simplecgroup/v2"

	"github.com/pkg/errors"
)

const stdioCount = 3

type pidMsg struct {
	ChildPID      int `json:"childpid"`
	GrandChildPid int `json:"grandchildpid"`
}

func StartInstance(conf *config.WrtxConfig) error {

	status := NewStatus()
	alwaryRestart := false

	if err := mountRootfs(*conf); err != nil {
		fmt.Println(err.Error())
		return errors.WithMessage(err, "mount rootfs error")
	}
	if conf.ResLimit {
		if err := setResLimit(conf, status); err != nil {
			return fmt.Errorf("set res limit error: %v", err)
		}
	}
	os.Args[0] = "wrtxd"
	var err error
	for {

		err = execInstance(conf, status)
		if err == nil {
			status.Dump(conf.StatusFile)
		}
		// if !backgroud {
		// 	pid, err := daemon.Daemon()
		// 	if err != nil {
		// 		break
		// 	}

		// 	if pid != 0 {
		// 		fmt.Println("pid=", pid)
		// 		break
		// 	}
		// 	backgroud = true
		// }
		var stat syscall.WaitStatus
		for {
			fmt.Println("wait child:", status.PID, "exit")
			syscall.Wait4(status.PID, &stat, 0, nil)
			if stat.Exited() {
				fmt.Printf("child pid:%d exit, status:%d\n", status.PID, stat.ExitStatus())
				break
			}
		}
		if !alwaryRestart {
			break
		}
	}
	return err
}

func execInstance(conf *config.WrtxConfig, status *Status) error {
	cmd := exec.Command("/proc/self/exe", "init")

	rootDir := conf.MergeDir

	r, _, cw, err := NewProcess(cmd, rootDir)
	if err != nil {
		return fmt.Errorf("create new process error")
	}

	msg := make([]byte, 4096)
	rr := os.NewFile(uintptr(r[0]), "__init_pipe")
	_, err = rr.Write([]byte("continue"))
	if err != nil {
		syscall.Kill(cmd.Process.Pid, syscall.SIGKILL)
		return fmt.Errorf("send msg to child err: %v", err)
	}
	n, err := rr.Read(msg)
	if err != nil && err != io.EOF {
		return fmt.Errorf("read msg from sub process error: %v", err)
	}
	// fmt.Println("read json msg:", string(msg[:n]))
	jsMsg := &pidMsg{}
	err = json.Unmarshal(msg[:n], jsMsg)
	if err != nil {
		return fmt.Errorf("get sub pid error: %v", err)
	}
	_, err = syscall.Wait4(jsMsg.ChildPID, nil, 0, nil)
	if err != nil {
		return fmt.Errorf("wait error: %v", err)
	}
	go func() {
		syscall.Wait4(jsMsg.ChildPID, nil, 0, nil)
	}()
	// fmt.Println("child pid:", jsMsg.ChildPID, "exit, wpid:", wpid)
	if conf.NicType == "ipvlan" {
		_, err = network.NewIPvlanDev(conf.NetDevName, conf.PhyDevName)
		if err != nil {
			syscall.Kill(jsMsg.GrandChildPid, syscall.SIGKILL)
			return err
		}
	} else if conf.NicType == "macvlan" {
		_, err = network.NewMacvlanDev(conf.NetDevName, conf.PhyDevName, conf.HardwareAddr)
		if err != nil {
			syscall.Kill(jsMsg.GrandChildPid, syscall.SIGKILL)
			return err
		}
	}
	err = network.AddDevToNamespaceByPID(conf.NetDevName, conf.NetDevName, jsMsg.GrandChildPid)
	if err != nil {
		syscall.Kill(jsMsg.GrandChildPid, syscall.SIGKILL)
		return err
	}
	var nsfiles = [...]string{"ipc", "net", "pid", "uts", "cgroup"}
	savedNSPath := fmt.Sprintf("%s/ns", conf.Instances)
	if err := saveNameSpaces(jsMsg.GrandChildPid, nsfiles[:], savedNSPath); err != nil {
		syscall.Kill(jsMsg.GrandChildPid, syscall.SIGKILL)
		return fmt.Errorf("save namespace error: %v", err)
	}

	cw.Write([]byte("continue"))
	status.PID = jsMsg.GrandChildPid
	status.Status = "Running"
	if err := status.Dump(conf.StatusFile); err != nil {
		return fmt.Errorf("dump status error: %v", err)
	}
	return nil
}

func mountRootfs(conf config.WrtxConfig) error {
	var wrtxPaths = [...]string{conf.WorkDir, conf.UpperDir, conf.MergeDir}
	for _, wrtxPath := range wrtxPaths {
		if stat, err := os.Stat(wrtxPath); err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("check path: %s error: %v", wrtxPath, err)
			}
			if err = os.MkdirAll(wrtxPath, os.ModePerm); err != nil {
				return errors.WithMessagef(err, "create dir %s error", wrtxPath)
			}
		} else {
			if !stat.IsDir() {
				return fmt.Errorf("check path: %s exist, but it's not a dir", wrtxPath)
			}
		}
	}
	err := fsMount.MountOverlayFs(conf.WorkDir, conf.UpperDir, conf.ImgPath, conf.MergeDir)
	if err != nil {
		return err
	}
	if conf.ConfigNetwork {
		target := fmt.Sprintf("%s/etc/config/network", conf.MergeDir)
		if _, err := os.Stat(target); err != nil {
			if os.IsNotExist(err) {
				fp, err := os.Create(target)
				if err != nil {
					return errors.WithMessagef(err, "create file %s error", target)
				}
				fp.Close()
			} else {
				return errors.WithMessagef(err, "check file %s error", target)
			}
		}
		err = fsMount.MountBind(target, conf.NetConfigFile)
		return errors.WithMessagef(err, "mount bind %s to %s error", conf.NetConfigFile, target)
	}
	return nil
}

func NewProcess(cmd *exec.Cmd, rootDir string) ([2]int, *os.File, *os.File, error) {
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	r, err := syscall.Socketpair(syscall.AF_LOCAL, syscall.SOCK_STREAM, 0)
	if err != nil {
		return [2]int{0}, nil, nil, fmt.Errorf("create init pipe error: %v", err)
	}
	cr, cw, err := os.Pipe()
	if err != nil {
		return r, nil, nil, fmt.Errorf("create ctrl pipe error: %v", err)
	}
	cmd.ExtraFiles = append(cmd.ExtraFiles, os.NewFile(uintptr(r[1]), "init_pipe"))

	cmd.Env = []string{fmt.Sprintf("_INIT_PIPE=%d", stdioCount+len(cmd.ExtraFiles)-1)}

	cmd.ExtraFiles = append(cmd.ExtraFiles, cr)
	cmd.Env = append(cmd.Env, fmt.Sprintf("_CTRL_PIPE=%d", stdioCount+len(cmd.ExtraFiles)-1))

	cmd.Env = append(cmd.Env, fmt.Sprintf("_ROOTDIR=%s", rootDir))
	// fmt.Println("cmd's env: ", cmd.Env)
	if err := cmd.Start(); err != nil {
		return r, cr, cw, fmt.Errorf("start cmd err: %s", err)
	}
	return r, cr, cw, nil
}

func timeHashString() string {
	now := time.Now().Unix()
	sum := md5.Sum([]byte(fmt.Sprintf("%d", now)))
	return fmt.Sprintf("%x", sum)[:8]
}

func saveNameSpaces(pid int, nsList []string, savedNSPath string) error {
	nsPath := fmt.Sprintf("/proc/%d/ns/", pid)
	if stat, err := os.Stat(savedNSPath); err == nil {
		if !stat.IsDir() {
			return fmt.Errorf("%s isn't a dir", savedNSPath)
		}
	} else {
		if !os.IsNotExist(err) {
			return fmt.Errorf("%s is exist but check it error: %v", savedNSPath, err)
		}
		err = os.Mkdir(savedNSPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("create dir %s error: %v", savedNSPath, err)
		}
	}
	for _, nsfile := range nsList {
		nsfilePath := fmt.Sprintf("%s/%s", nsPath, nsfile)
		savedNSFilePath := fmt.Sprintf("%s/%s", savedNSPath, nsfile)
		if _, err := os.Stat(savedNSFilePath); err != nil {
			if os.IsNotExist(err) {
				fmt.Println("create file:", savedNSFilePath)
				fp, err := os.Create(savedNSFilePath)
				if err != nil {
					fmt.Println("create file", savedNSFilePath, " err: ", err)
					continue
				}
				fp.Close()
			}
		}
		err := fsMount.MountBind(savedNSFilePath, nsfilePath)
		if err != nil {
			fmt.Printf("save namespace: %s to %s error: %v\n", nsfilePath, savedNSFilePath, err)
		}

	}
	return nil
}

func setResLimit(conf *config.WrtxConfig, status *Status) error {
	cg, err := simplecgroup.GetCgroupType()
	if err != nil {
		syscall.Kill(os.Getpid(), syscall.SIGKILL)
		return err
	}
	if cg&simplecgroup.CGTypeTwo != 0 {
		cgroupSubDir := fmt.Sprintf("wrtx_%s", timeHashString())
		// fmt.Println("cgroupSubDir:", cgroupSubDir)
		status.CgroupPath = cgroupSubDir
		mgr, err := cgroupv2.New("", cgroupSubDir)
		if err != nil {
			return err
		}
		err = mgr.SetCPUMemLimit(conf.Cpus, conf.Period, conf.Mem)
		if err != nil {
			return err
		}
		err = mgr.AddProcesssors([]int{os.Getpid()})
		if err != nil {
			return err
		}
	}
	return nil
}
