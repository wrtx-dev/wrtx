package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"syscall"
	"wrtx/internal/config"

	"github.com/urfave/cli/v2"
)

func init() {
	if os.Getenv("NSLIST") != "" {
		runtime.GOMAXPROCS(1)
		runtime.LockOSThread()
	}
	// fmt.Println("os.args:", os.Args)
}

var proxyCmd = cli.Command{
	Name:  "proxy",
	Usage: "create a locl proxy service to use openwrt's webui",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "p",
			Usage: "local listen port, default: 80",
		},
	},
	Action: proxyAction,
}

func proxyAction(ctx *cli.Context) error {
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

		port := ctx.String("p")
		if port == "" {
			port = "80"
		}
		addr := fmt.Sprintf("0.0.0.0:%s", port)
		l, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("listen on %s error: %v", addr, err)
		}
		fmt.Println(os.Args[1:])
		cmd := exec.Command("/proc/self/exe", os.Args[1:]...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		tcpL := l.(*net.TCPListener)
		if tcpFile, err := tcpL.File(); err == nil {
			cmd.ExtraFiles = append(cmd.ExtraFiles, tcpFile)
		} else {
			return fmt.Errorf("get tcp listener's file error: %v", err)
		}
		cmd.Env = []string{fmt.Sprintf("NSPID=%d", pid)}
		cmd.Env = append(cmd.Env, fmt.Sprintf("NSLIST=%d", syscall.CLONE_NEWNET))
		fmt.Println("run cmd")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("run proxy error: %v", err)
		}

	} else {
		fmt.Println("into tcp")
		tcpFP := os.NewFile(uintptr(3), "tcplistener")
		l, err := net.FileListener(tcpFP)
		if err != nil {
			return fmt.Errorf("recover tcp listener error: %v", err)
		}

		for {
			tcpListener := l.(*net.TCPListener)
			fmt.Println("waiting for connect")
			conn, err := tcpListener.Accept()
			if err != nil {
				fmt.Println("accept error:", err)
			}
			fmt.Println("get conn, addr:", conn.RemoteAddr())
			go func(c *net.Conn) {
				ctrl := make(chan bool)
				defer close(ctrl)
				defer (*c).Close()
				cc, err := net.Dial("tcp", "127.0.0.1:80")
				if err != nil {
					fmt.Printf("connect to openwrt's webui error: %v\n", err)
				}
				defer cc.Close()
				go func() {
					io.Copy(conn, cc)
					ctrl <- true
				}()
				go func() {
					io.Copy(cc, conn)
					ctrl <- true
				}()
				for range 2 {
					<-ctrl
				}

			}(&conn)
		}
	}
	return nil
}
