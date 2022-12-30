package network

import (
	"errors"
	"net"
	"strings"

	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var wg_client *wgctrl.Client

func (t *Tool) wgConfigure() error {
	var err error
	wg_client, err = wgctrl.New()
	if err != nil {
		return err
	}

	// Create wg interface
	if isWgExists(t.deps.WireguardInterface) {
		return errors.New("wg interface already exists")
	}

	linkAtrrs := netlink.NewLinkAttrs()
	linkAtrrs.Name = t.deps.WireguardInterface
	linkAtrrs.MTU = 1420
	linkAtrrs.TxQLen = 1000

	wg_link := wgLink{}
	wg_link.LinkAttrs = &linkAtrrs
	wg_link.LinkType = "wireguard"

	handle, err := netlink.NewHandle()
	if err != nil {
		return err
	}

	err = handle.LinkAdd(netlink.Link(wg_link))
	if err != nil {
		return err
	}

	// Configure wg interface
	pk, err := wgtypes.ParseKey(t.deps.PrivateKey)
	if err != nil {
		return err
	}

	config := wgtypes.Config{
		PrivateKey:   &pk,
		ListenPort:   &t.deps.Port,
		ReplacePeers: false,
		Peers:        make([]wgtypes.PeerConfig, 0),
	}

	link, err := handle.LinkByName(t.deps.WireguardInterface)
	if err != nil {
		return err
	}

	addr, err := netlink.ParseAddr(t.deps.WireguardIpNet)
	if err != nil {
		return err
	}

	err = handle.AddrAdd(link, addr)
	if err != nil {
		return err
	}

	err = wg_client.ConfigureDevice(t.deps.WireguardInterface, config)
	if err != nil {
		return err
	}

	// Up wg interface
	nl_link, err := handle.LinkByName(t.deps.WireguardInterface)
	if err != nil {
		return err
	}

	err = netlink.LinkSetUp(netlink.Link(nl_link))
	return err
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
	err = wg_client.ConfigureDevice(t.deps.WireguardInterface, wgtypes.Config{Peers: peers})

	return err
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
	err = wg_client.ConfigureDevice(t.deps.WireguardInterface, wgtypes.Config{Peers: peers})

	return err
}

func isWgExists(inf string) bool {

	_, err := netlink.LinkByName(inf)

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

// func wgDelete() error {
// 	linkAtrrs := netlink.NewLinkAttrs()
// 	linkAtrrs.Name = "wg0"
// 	linkAtrrs.MTU = 1420
// 	linkAtrrs.TxQLen = 1000

// 	link := wgLink{}
// 	link.LinkAttrs = &linkAtrrs
// 	link.LinkType = "wireguard"

// 	err := netlink.LinkDel(netlink.Link(link))
// 	return err
// }

// func wgDown() error {
// 	link, err := netlink.LinkByName("wg0")
// 	if err != nil {
// 		return err
// 	}

// 	err = netlink.LinkSetDown(netlink.Link(link))
// 	return err
// }
