package main

import (
	"fmt"
	"io"
	"net"

	"github.com/urfave/cli/v2"
)

var serveCmd = cli.Command{
	Name: "serve",
	Usage: "start an unix socket server",
	Action: startServer,

}

func startServer(ctx *cli.Context) error {
	l, err := net.Listen("unix", "/serve_path")
	if err != nil {
		return fmt.Errorf("start server on path: serve_path error: %v", err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go func(c *net.Conn) {
			defer (*c).Close()
			ctrl := make(chan bool)
			cli, err := net.Dial("tcp", "127.0.0.1:80")
			if err != nil {
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
			for i:=0; i< 2; i++ {
				<- ctrl
			}
		}(&conn)
	}
}