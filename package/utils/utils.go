package utils

import (
	"os"
	"runtime"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

var (
	haveProcThreadSelf     bool
	haveProcThreadSelfOnce sync.Once
)

type ProcThreadSelfCloser func()

func ProcThreadSelfFd(fd uintptr) (string, ProcThreadSelfCloser) {
	return ProcThreadSelf("fd/" + strconv.FormatUint(uint64(fd), 10))
}
func ProcThreadSelf(subpath string) (string, ProcThreadSelfCloser) {
	haveProcThreadSelfOnce.Do(func() {
		if _, err := os.Stat("/proc/thread-self/"); err == nil {
			haveProcThreadSelf = true
		} else {
			logrus.Debugf("cannot stat /proc/thread-self (%v), falling back to /proc/self/task/<tid>", err)
		}
	})

	runtime.LockOSThread()

	threadSelf := "/proc/thread-self/"
	if !haveProcThreadSelf {
		threadSelf = "/proc/self/task/" + strconv.Itoa(unix.Gettid()) + "/"
		if _, err := os.Stat(threadSelf); err != nil {
			if os.Getpid() == 1 {
				logrus.Debugf("/proc/thread-self (tid=%d) cannot be emulated inside the initial container setup -- using /proc/self instead: %v", unix.Gettid(), err)
			} else {
				// This should never happen, but the fallback should work in most cases...
				logrus.Warnf("/proc/thread-self could not be emulated for pid=%d (tid=%d) -- using more buggy /proc/self fallback instead: %v", os.Getpid(), unix.Gettid(), err)
			}
			threadSelf = "/proc/self/"
		}
	}
	return threadSelf + subpath, runtime.UnlockOSThread
}
