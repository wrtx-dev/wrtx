package init

import (
	"os"
	"runtime"
	"wrtx/package/libinit"
	_ "wrtx/package/nsenter"

	"github.com/sirupsen/logrus"
)

func init() {
	if len(os.Args) < 2 || os.Args[1] != "init" {
		return
	}
	runtime.GOMAXPROCS(1)
	runtime.LockOSThread()
	logrus.Debug("start init...")
	libinit.Init()
	os.Exit(0)
}
