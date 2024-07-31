package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"wrtx/internal/config"
	"wrtx/internal/utils"

	"github.com/urfave/cli/v2"
)

const fdIndexBegin = 3

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
		&cli.StringSliceFlag{
			Name:  "map",
			Usage: "port map inside and outside namespace,format inside_port:outside_port, eg. -map 443:80443",
		},
		&cli.StringFlag{
			Name:  "conf",
			Usage: "global config file path, default is" + config.DefaultConfPath,
		},
	},
	Action: proxyAction,
}

func proxyAction(ctx *cli.Context) error {
	mapsList := ctx.StringSlice("map")
	portMaps := make(map[string]string)
	name := ""
	if len(ctx.Args().Slice()) > 0 {
		name = ctx.Args().First()
	}
	pid, err := utils.GetInstancesPid(ctx.String("conf"), name)
	if err != nil {
		return fmt.Errorf("get instance pid error: %v", err)
	}
	if len(mapsList) == 0 {
		mapsList = append(mapsList, []string{"80:80", "443:443"}...)
	}

	for _, portPeer := range mapsList {
		peers := strings.Split(portPeer, ":")
		if len(peers) != 2 {
			return fmt.Errorf("parse port map [%s] error", portPeer)
		}
		if _, ok := portMaps[peers[0]]; !ok {
			portMaps[peers[0]] = peers[1]
		} else {
			return fmt.Errorf("dup port map error")
		}
	}
	portMapsKeys := make([]string, 0, len(portMaps))
	for k := range portMaps {
		portMapsKeys = append(portMapsKeys, k)
	}
	sort.Strings(portMapsKeys)

	if os.Getenv("NSLIST") == "" {

		cmd := exec.Command("/proc/self/exe", os.Args[1:]...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		for _, dstPort := range portMapsKeys {
			srcPort, ok := portMaps[dstPort]
			if !ok {
				return fmt.Errorf("get src port for port %s error", dstPort)
			}
			addr := fmt.Sprintf("0.0.0.0:%s", srcPort)
			l, err := net.Listen("tcp", addr)
			if err != nil {
				return fmt.Errorf("listen on %s error: %v", addr, err)
			}
			tcpListener := l.(*net.TCPListener)
			if tcpFile, err := tcpListener.File(); err == nil {
				cmd.ExtraFiles = append(cmd.ExtraFiles, tcpFile)
			} else {
				return fmt.Errorf("get tcp listener's file error: %v", err)
			}
		}

		cmd.Env = []string{fmt.Sprintf("NSPID=%d", pid)}
		cmd.Env = append(cmd.Env, fmt.Sprintf("NSLIST=%d", syscall.CLONE_NEWNET))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("run proxy error: %v", err)
		}

	} else {
		listeners := make([]*net.Listener, 0)
		idx := fdIndexBegin
		for range len(portMapsKeys) {
			i := idx
			idx += 1
			lFp := os.NewFile(uintptr(i), fmt.Sprintf("tcplistener_%d", i))
			tcpListener, err := net.FileListener(lFp)
			if err != nil {
				return fmt.Errorf("recover tcp listener error: %v", err)
			}
			go proxyPort(&tcpListener, portMaps[portMapsKeys[i-fdIndexBegin]])
			listeners = append(listeners, &tcpListener)
		}
		done := make(chan bool)
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			sig := <-sigs
			fmt.Println("recived signal:", sig)
			done <- true
		}()
		<-done
		for _, l := range listeners {
			(*l).Close()
		}

	}
	return nil
}

func proxyPort(l *net.Listener, port string) {
	dstPort := port
	for {
		c, err := (*l).Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			break
			//TODO: fix exit error
		}
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%s", dstPort))
		if err != nil {
			fmt.Printf("connect to openwrt's webui error: %v\n", err)
			c.Close()
			continue

		}
		go dataCopy(c, conn)
	}
}

func dataCopy(from, to net.Conn) {
	defer from.Close()
	defer to.Close()
	ctrl := make(chan bool)
	go func() {
		io.Copy(from, to)
		ctrl <- true
	}()
	go func() {
		io.Copy(to, from)
		ctrl <- true
	}()
	for range 2 {
		<-ctrl
	}
}
