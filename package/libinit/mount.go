package libinit

import (
	"wrtx/package/utils"
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

type mountSourceType string
type mountSource struct {
	Type mountSourceType `json:"type"`
	file *os.File        `json:"-"`
}

type mountError struct {
	op      string
	source  string
	srcFile *mountSource
	target  string
	dstFd   string
	flags   uintptr
	data    string
	err     error
}

func (e *mountError) Error() string {
	out := e.op + " "

	if e.source != "" {
		out += "src=" + e.source + ", "
		if e.srcFile != nil {
			out += "srcType=" + string(e.srcFile.Type) + ", "
			out += "srcFd=" + strconv.Itoa(int(e.srcFile.file.Fd())) + ", "
		}
	}
	out += "dst=" + e.target
	if e.dstFd != "" {
		out += ", dstFd=" + e.dstFd
	}

	if e.flags != uintptr(0) {
		out += ", flags=0x" + strconv.FormatUint(uint64(e.flags), 16)
	}
	if e.data != "" {
		out += ", data=" + e.data
	}

	out += ": " + e.err.Error()
	return out
}

func mountViaFds(source string, srcFile *mountSource, target, dstFd, fstype string, flags uintptr, data string) error {
	if srcFile != nil && flags&unix.MS_REMOUNT != 0 {
		srcFile = nil
	}
	dst := target
	if dstFd != "" {
		dst = dstFd
	}
	src := source
	isMoveMount := srcFile != nil && srcFile.Type == "open_tree"
	if srcFile != nil {
		srcFileFd := srcFile.file.Fd()
		if isMoveMount {
			src = "/proc/self/fd" + strconv.Itoa(int(srcFileFd))
		} else {
			var closer utils.ProcThreadSelfCloser
			src, closer = utils.ProcThreadSelfFd(srcFileFd)
			defer closer()
		}
	}
	var op string
	var err error
	if isMoveMount {
		op = "move_mount"
		err = unix.MoveMount(int(srcFile.file.Fd()), "", unix.AT_FDCWD, dstFd,
			unix.MOVE_MOUNT_F_EMPTY_PATH|unix.MOVE_MOUNT_T_SYMLINKS)
	} else {
		op = "mount"
		err = unix.Mount(src, dst, fstype, flags, data)
	}
	if err != nil {
		return &mountError{
			op:      op,
			source:  source,
			srcFile: srcFile,
			target:  target,
			dstFd:   dstFd,
			flags:   flags,
			data:    data,
			err:     err,
		}
	}
	return nil
}

func mount(source, target, fstype string, flags uintptr, data string) error {
	return mountViaFds(source, nil, target, "", fstype, flags, data)
}
