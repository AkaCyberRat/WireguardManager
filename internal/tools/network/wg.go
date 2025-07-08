package network

import (
	"errors"
	"net"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	WgIpNet = "10.0.0.0/8"
	WgIp    = "10.0.0.0"
)

func (t *Tool) wgServerUp(privateKey string) error {
	var err error

	// Create wg interface
	if isWgExists(t.interfaceName) {
		return errors.New("wg interface already exists")
	}

	linkAtrrs := netlink.NewLinkAttrs()
	linkAtrrs.Name = t.interfaceName
	linkAtrrs.MTU = 1420
	linkAtrrs.TxQLen = 1000

	wg_link := wgLink{}
	wg_link.LinkAttrs = &linkAtrrs
	wg_link.LinkType = "wireguard"

	handle, err := netlink.NewHandle()
	if err != nil {
		return err
	}

	if err = handle.LinkAdd(netlink.Link(wg_link)); err != nil {
		return err
	}

	// Configure wg interface
	pk, err := wgtypes.ParseKey(privateKey)
	if err != nil {
		return err
	}

	config := wgtypes.Config{
		PrivateKey:   &pk,
		ListenPort:   &t.port,
		ReplacePeers: false,
		Peers:        make([]wgtypes.PeerConfig, 0),
	}

	link, err := handle.LinkByName(t.interfaceName)
	if err != nil {
		return err
	}

	addr, err := netlink.ParseAddr(WgIpNet)
	if err != nil {
		return err
	}

	if err = handle.AddrAdd(link, addr); err != nil {
		return err
	}

	if err = t.wgClient.ConfigureDevice(t.interfaceName, config); err != nil {
		return err
	}

	// Up wg interface
	nl_link, err := handle.LinkByName(t.interfaceName)
	if err != nil {
		return err
	}

	if err = netlink.LinkSetUp(netlink.Link(nl_link)); err != nil {
		return err
	}

	logrus.Tracef("Wireguard interface enabled. [InterfaceName=%v, IpNet=%v, Port=%v]", t.interfaceName, WgIpNet, t.port)
	return nil
}

func (t *Tool) wgServerDown() error {
	linkAtrrs := netlink.NewLinkAttrs()
	linkAtrrs.Name = t.interfaceName
	linkAtrrs.MTU = 1420
	linkAtrrs.TxQLen = 1000

	link := wgLink{}
	link.LinkAttrs = &linkAtrrs
	link.LinkType = "wireguard"

	if err := netlink.LinkDel(netlink.Link(link)); err != nil {
		return err
	}

	logrus.Tracef("Wireguard interface disabled. [InterfaceName=%v, IpNet=%v, Port=%v]", t.interfaceName, WgIpNet, t.port)
	return nil
}

func (t *Tool) wgPeerUp(ip string, publicKey string, presharedKey string) error {
	pubKey, err := wgtypes.ParseKey(publicKey)
	if err != nil {
		return err
	}
	_, ipnet, err := net.ParseCIDR(ip + "/32")
	if err != nil {
		return err
	}

	ipAddresses := []net.IPNet{*ipnet}
	// interval := time.Minute
	peer := wgtypes.PeerConfig{
		PublicKey:  pubKey,
		AllowedIPs: ipAddresses,
		// PersistentKeepaliveInterval: &interval,
	}

	if strings.TrimSpace(presharedKey) != "" {
		preKey, err := wgtypes.ParseKey(presharedKey)
		if err != nil {
			return err
		}
		peer.PresharedKey = &preKey
	}

	peers := []wgtypes.PeerConfig{peer}
	if err = t.wgClient.ConfigureDevice(t.interfaceName, wgtypes.Config{Peers: peers}); err != nil {
		return err
	}

	logrus.Tracef("Wireguard peer enabled. [Ip=%v, PubKey=%v]", ip, publicKey)
	return nil
}

func (t *Tool) wgPeerDown(ip string, publicKey string) error {
	pubKey, err := wgtypes.ParseKey(publicKey)
	if err != nil {
		return err
	}
	_, ipnet, err := net.ParseCIDR(ip + "/32")
	if err != nil {
		return err
	}

	ipAddresses := []net.IPNet{*ipnet}

	peer := wgtypes.PeerConfig{
		PublicKey:  pubKey,
		AllowedIPs: ipAddresses,
		Remove:     true,
	}

	peers := []wgtypes.PeerConfig{peer}
	if err = t.wgClient.ConfigureDevice(t.interfaceName, wgtypes.Config{Peers: peers}); err != nil {
		return err
	}

	logrus.Tracef("Wireguard peer disabled. [Ip=%v, PubKey=%v]", ip, publicKey)
	return nil
}

//
// Helpers
//

func isWgExists(interfaceName string) bool {

	_, err := netlink.LinkByName(interfaceName)

	return err == nil
}

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
