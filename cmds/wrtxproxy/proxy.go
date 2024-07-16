package main

import (
	"fmt"
	"io"
	"net"

	"github.com/urfave/cli/v2"
)

var proxyCmd = cli.Command{
	Name:  "proxy",
	Usage: "start a proxy",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "unix",
			Usage: "unix socket's addr",
		},
		&cli.StringFlag{
			Name:  "port",
			Usage: "proxy listen port",
		},
	},
	Action: startProxy,
}

func startProxy(ctx *cli.Context) error {
	unixAddr := ctx.String("unix")
	port := ctx.String("port")
	if port == "" {
		port = "80"
	}
	if unixAddr == "" {
		return fmt.Errorf("param unix can't be empty")
	}
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	if err != nil {
		return fmt.Errorf("start server on path: serve_path error: %v", err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}
		go func(c *net.Conn) {
			defer (*c).Close()
			ctrl := make(chan bool)
			cli, err := net.Dial("unix", unixAddr)
			if err != nil {
				fmt.Println("connect to", unixAddr, "error:", err)
				return
			}
			go func() {
				io.Copy(*c, cli)
				ctrl <- true
			}()
			go func() {
				io.Copy(cli, *c)
				ctrl <- true
			}()
			for i := 0; i < 2; i++ {
				<-ctrl
			}
		}(&conn)
	}
}
