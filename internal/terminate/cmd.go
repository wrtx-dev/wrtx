package terminate

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func TerminateCmd(pid int) error {
	cmd := exec.Command("/proc/self/exe", "terminate")
	cmd.Env = []string{fmt.Sprintf("NSPID=%d", pid)}
	cmd.Env = append(cmd.Env, fmt.Sprintf("NSLIST=%d", syscall.CLONE_NEWNS|syscall.CLONE_NEWNET|syscall.CLONE_NEWCGROUP))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
