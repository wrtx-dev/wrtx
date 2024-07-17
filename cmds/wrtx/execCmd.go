package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"wrtx/internal/config"

	"github.com/urfave/cli/v2"
)

var execCmd = cli.Command{
	Name:   "exec",
	Usage:  "execute a command in openwrt",
	Action: execAction,
}

func init() {
	if os.Getenv("NSLIST") != "" {
		runtime.GOMAXPROCS(1)
		runtime.LockOSThread()
	}
	// fmt.Println("os.args:", os.Args)
}
func execAction(ctx *cli.Context) error {
	// runtime.GOMAXPROCS(1)
	// runtime.LockOSThread()
	args := ctx.Args().Slice()
	// fmt.Println("args:", args)
	if os.Getenv("NSLIST") == "" {
		pidfile := config.DefaultWrtxRunPidFile
		buf, err := os.ReadFile(pidfile)
		if err != nil {
			return fmt.Errorf("read file %s error: %v", pidfile, err)
		}
		pid, err := strconv.Atoi(string(buf))
		if err != nil {
			return fmt.Errorf("covert %s to int error", string(buf))
		}

		fp, err := os.Open(fmt.Sprintf("/proc/%d/environ", pid))
		if err != nil {
			return fmt.Errorf("open environ file: %s error: %v", fmt.Sprintf("/proc/%d/environ"), err)
		}
		defer fp.Close()

		cmd := exec.Command("/proc/self/exe", os.Args[1:]...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.ExtraFiles = append(cmd.ExtraFiles, fp)
		cmd.Env = []string{fmt.Sprintf("NSPID=%d", pid)}
		cmd.Env = append(cmd.Env, fmt.Sprintf("NSLIST=%d", syscall.CLONE_NEWIPC|syscall.CLONE_NEWNET|syscall.CLONE_NEWNS|syscall.CLONE_NEWPID|syscall.CLONE_NEWUTS|syscall.CLONE_NEWCGROUP))

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("start new exec process error: %v", err)
		}

		cmd.Wait()
	} else {
		var cmd *exec.Cmd
		envFile := os.NewFile(uintptr(3), "environ")
		envBuf, err := io.ReadAll(envFile)
		if err != nil {
			return fmt.Errorf("read environ file error: %v", err)
		}
		envsBuf := bytes.Split(envBuf, []byte{0})
		os.Clearenv()
		for _, envBytes := range envsBuf {
			envKV := strings.Split(string(envBytes), "=")
			if len(envKV) != 2 {
				continue
			}
			if err = os.Setenv(envKV[0], envKV[1]); err != nil {
				return fmt.Errorf("set env variable %s value: %s error: %v", envKV[0], envKV[1], err)
			}
		}
		cmdPath, err := exec.LookPath(args[0])
		if err != nil {
			return fmt.Errorf("looking path for %s error: %v", args[0], err)
		}
		if len(args) > 1 {
			cmd = exec.Command(cmdPath, args[1:]...)
		} else {
			cmd = exec.Command(cmdPath)
		}
		cmd.Env = make([]string, 0)
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, "NSLIST") || strings.HasPrefix(env, "NSPID") {
				continue
			}
			cmd.Env = append(cmd.Env, env)
		}
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("start run cmd error: %v", err)
		}
		cmd.Wait()
	}
	return nil
}
