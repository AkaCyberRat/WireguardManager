package network

import "github.com/vishvananda/netlink"

type wgLink struct {
	LinkAttrs *netlink.LinkAttrs
	LinkType  string
}

func (l wgLink) Attrs() *netlink.LinkAttrs {
	return l.LinkAttrs
}

func (l wgLink) Type() string {
	return l.LinkType
}

func isWgInfExists() bool {

	_, err := netlink.LinkByName("wg0")

	return err == nil
}
