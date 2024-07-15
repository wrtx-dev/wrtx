package init

import (
	"fmt"
	"wrtx/package/libinit"
	_ "wrtx/package/nsenter"
	"os"
	"runtime"
)

func init() {
	if len(os.Args) < 2 || os.Args[1] != "init" {
		return
	}
	runtime.GOMAXPROCS(1)
	runtime.LockOSThread()
	fmt.Println("start init...")
	libinit.Init()
	os.Exit(0)
}
