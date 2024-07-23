package rmnet

import (
	"fmt"
	"os"
	"strings"

	"github.com/vishvananda/netlink"
)

func init() {
	if len(os.Args) != 2 || os.Args[1] != "rmnet" {
		return
	}
	links, err := netlink.LinkList()
	if err != nil {
		fmt.Println("get links error:", err)
	}
	for _, link := range links {
		linkType := strings.ToLower(link.Type())
		if linkType == "ipvlan" || linkType == "macvlan" {
			if err := netlink.LinkSetDown(link); err != nil {
				fmt.Println("set link:", link.Attrs().Name, "down error:", err)
				continue
			}
			if err := netlink.LinkDel(link); err != nil {
				fmt.Println("delete link:", link.Attrs().Name, "err:", err)
				continue
			}
		}

	}
	os.Exit(0)
}
