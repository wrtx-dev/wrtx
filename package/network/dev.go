package network

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const (
	typePid  = 1
	typeName = 2
)

func NewMacvlanDev(name, parent string) (netlink.Link, error) {
	if _, err := netlink.LinkByName(name); err == nil {
		return nil, fmt.Errorf("dev %s existed", name)
	}
	dev, err := netlink.LinkByName(parent)
	if err != nil {
		return nil, errors.WithMessagef(err, "get dev: %v error", parent)
	}

	la := netlink.NewLinkAttrs()
	la.Name = name
	la.ParentIndex = dev.Attrs().Index

	macvlan := &netlink.Macvlan{
		LinkAttrs: la,
		Mode:      netlink.MACVLAN_MODE_BRIDGE,
	}
	err = netlink.LinkAdd(macvlan)
	return macvlan, errors.WithMessagef(err, "add dev %s error", name)
}

func NewIPvlanDev(name, parent string) (netlink.Link, error) {
	if _, err := netlink.LinkByName(name); err == nil {
		return nil, fmt.Errorf("dev %s existed", name)
	}
	dev, err := netlink.LinkByName(parent)

	if err != nil {
		return nil, errors.WithMessagef(err, "get dev: %s error", parent)
	}
	la := netlink.NewLinkAttrs()
	la.Name = name
	la.ParentIndex = dev.Attrs().Index
	ipVlan := &netlink.IPVlan{
		LinkAttrs: la,
		Mode:      netlink.IPVLAN_MODE_L2,
		Flag:      netlink.IPVLAN_FLAG_BRIDGE,
	}
	return ipVlan, netlink.LinkAdd(ipVlan)
}

func DelDevByName(name string) error {
	if dev, err := netlink.LinkByName(name); err == nil {
		err = netlink.LinkDel(dev)
		return errors.Wrapf(err, "del net dev %s error", name)
	} else {
		return errors.WithMessagef(err, "get net dev %s error", name)
	}
}

func AddDevToNamespaceByPID(dev, name string, pid int) error {
	ndev, err := netlink.LinkByName(dev)
	if err != nil {
		return err
	}
	_fn, err := enterNamespaceWithOutter(&ndev, strconv.Itoa(pid), typePid)
	if err != nil {
		return err
	}
	defer _fn()
	if name != "" {
		err = netlink.LinkSetName(ndev, name)
		if err != nil {
			return err
		}
	}
	return netlink.LinkSetUp(ndev)

}

func AddDevToNamespaceByName(dev, name string) error {
	ndev, err := netlink.LinkByName(dev)
	if err != nil {
		return err
	}
	_fn, err := enterNamespaceWithOutter(&ndev, name, typeName)
	if err != nil {
		return err
	}
	defer _fn()
	return nil
}

func AddLinkToNamespaceByPID(dev *netlink.Link, pid int) error {
	_fn, err := enterNamespaceWithOutter(dev, strconv.Itoa(pid), typePid)
	if err != nil {
		return err
	}
	defer _fn()

	return nil
}

func AddLinkToNamespaceByName(dev *netlink.Link, name string) error {
	_fn, err := enterNamespaceWithOutter(dev, name, typeName)
	if err != nil {
		return err
	}
	defer _fn()
	return nil
}

func enterNamespaceWithOutter(dev *netlink.Link, name string, spaceType int) (fn func(), err error) {
	var handler netns.NsHandle
	switch spaceType {
	case typePid:
		ipid, err := strconv.Atoi(name)
		if err != nil {
			return nil, errors.WithMessagef(err, "get namespace from %s error", name)
		}
		handler, err = netns.GetFromPid(ipid)
		if err != nil {
			return nil, errors.WithMessagef(err, "get namespace from %d error", ipid)
		}
	case typeName:
		handler, err = netns.GetFromName(name)
		if err != nil {
			return nil, errors.Wrapf(err, "get namespace from name: %s error", name)
		}
	default:
		return nil, errors.New("unknown type")

	}
	originNs, err := netns.Get()
	if err != nil {
		return nil, err
	}
	runtime.LockOSThread()
	if err := netlink.LinkSetNsFd(*dev, int(handler)); err != nil {
		return nil, err
	}
	fn = func() {
		netns.Set(originNs)
		runtime.UnlockOSThread()
	}
	err = netns.Set(handler)
	if err != nil {
		return nil, err
	}
	// netlink.LinkSetName(*dev, "eth0")
	return fn, nil

}

func CheckDevPromisc(name string) error {
	if dev, err := netlink.LinkByName(name); err != nil {
		fmt.Println("get dev ", name, " error:", err)
		return err
	} else {
		fmt.Println("Promisc: ", dev.Attrs().Promisc)
		return nil
	}
}

func SetNetDevPromisc(name string) error {
	dev, err := netlink.LinkByName(name)
	if dev.Attrs().Promisc == 1 {
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "get dev %s error", name)
	}
	return netlink.SetPromiscOn(dev)

}

func GetCurrentNetns() (netns.NsHandle, error) {
	return netns.GetFromPid(os.Getpid())
}

func BackToCurrentNetns() (func(), error) {
	h, err := GetCurrentNetns()
	if err != nil {
		return nil, err
	}
	return func() {
		netns.Set(h)

	}, err
}

func SetNetDevNameInNetnsByPid(old, name string, pid int) error {
	originNs, err := netns.Get()
	if err != nil {
		return fmt.Errorf("get origin net ns error %v", err)
	}
	defer func() {
		netns.Set(originNs)
		runtime.UnlockOSThread()
	}()

	dstNs, err := netns.GetFromPid(pid)
	if err != nil {
		return fmt.Errorf("get new net ns from pid: %d ,error %v", pid, err)
	}
	runtime.LockOSThread()
	err = netns.Set(dstNs)
	if err != nil {
		return fmt.Errorf("set new net ns error: %v", err)
	}
	link, err := netlink.LinkByName(old)
	if err != nil {
		return fmt.Errorf("get dev: %s error: %v", old, err)
	}
	return netlink.LinkSetName(link, name)
}

func SetNetDevsUp(devs []string) error {
	for _, name := range devs {
		dev, err := netlink.LinkByName(name)
		if err != nil {
			return err
		}
		err = netlink.LinkSetUp(dev)
		if err != nil {
			return err
		}
	}
	return nil
}
