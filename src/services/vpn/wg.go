package vpn

import (
	"net"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var wg_client *wgctrl.Client

func wgPeerUp(ip string, publicKey string, presharedKey string) error {
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
	err = wg_client.ConfigureDevice("wg0", wgtypes.Config{Peers: peers})

	return err
}
func wgPeerDown(ip string, publicKey string) error {
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
	err = wg_client.ConfigureDevice("wg0", wgtypes.Config{Peers: peers})

	return err
}

func wgConfigure() error {

	client, err := wgctrl.New()
	if err != nil {
		return err
	}
	wg_client = client

	err = wgCreate()
	if err != nil {
		return err
	}

	err = wgSetConf("10.0.0.1/8", 51830, "8EUc3jnJUrWoZVDFA4KL6Rs++a34sfMmukYWxnAGDmA=")
	if err != nil {
		return err
	}

	err = wgUp()
	return err
}

func wgCreate() error {
	isExists := isWgInfExists()
	if isExists {
		return nil
	}

	// handle, err := netlink.NewHandle()
	// if err != nil {
	// 	return err
	// }

	h, _ := netlink.NewHandle()

	linkAtrrs := netlink.NewLinkAttrs()
	linkAtrrs.Name = "wg0"
	linkAtrrs.MTU = 1420
	linkAtrrs.TxQLen = 1000

	link := wgLink{}
	link.LinkAttrs = &linkAtrrs
	link.LinkType = "wireguard"

	err := h.LinkAdd(netlink.Link(link))
	return err
}

func wgUp() error {
	h, _ := netlink.NewHandle()

	link, err := h.LinkByName("wg0")
	if err != nil {
		return err
	}

	err = netlink.LinkSetUp(netlink.Link(link))
	return err
}

func wgSetConf(ipv4 string, port int, privateKey string) error {

	wg_client, err := wgctrl.New()
	if err != nil {
		return err
	}

	pk, err := wgtypes.ParseKey(privateKey) //wgtypes.GeneratePrivateKey()
	if err != nil {
		return err
	}

	config := wgtypes.Config{
		PrivateKey:   &pk,
		ListenPort:   &port,
		ReplacePeers: false,
		Peers:        make([]wgtypes.PeerConfig, 0),
	}

	logrus.Infof("Server ip: %v", ipv4)
	logrus.Infof("Server port: 51830")
	logrus.Infof("Server pk: %v", pk.String())
	logrus.Infof("Server pk: %v", pk.PublicKey().String())

	h, _ := netlink.NewHandle()

	link, _ := h.LinkByName("wg0")
	addr, _ := netlink.ParseAddr(ipv4)
	err = h.AddrAdd(link, addr)
	if err != nil {
		return err
	}

	err = wg_client.ConfigureDevice("wg0", config)
	return err
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
