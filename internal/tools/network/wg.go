package network

import (
	"errors"
	"golang.zx2c4.com/wireguard/wgctrl"
	"net"
	"net/netip"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"WireguardManager/pkg/shell"
)

const (
	WgIpNet = "11.0.0.1/24"
	WgIp    = "11.0.0.1"
)

func (t *Tool) wgServerUp(privateKey string) error {
	var err error

	// Create wg interface
	if isWgExists(t.interfaceName) {
		return errors.New("wg interface already exists")
	}

	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = t.interfaceName
	linkAttrs.MTU = 1420
	linkAttrs.TxQLen = 1000

	wg_link := wgLink{}
	wg_link.LinkAttrs = &linkAttrs
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

// SetupWgInterface creates and configures a wireguard interface with the given parameters.
// - 'infName' is the name of wg interface (e.g. 'wg0')
//
// - 'ip' is the IP address of wg interface (e.g. '11.0.0.1')
//
// - 'mask' is the subnet mask of wg interface (e.g. '24')
//
// - 'privateKey' is the private key of wg server
//
// - 'port' is the port number on which wg server listens (e.g. 51820)
func SetupWgInterface(infName string, ip net.IP, mask net.IPMask, privateKey string, port int) error {
	const WgLinkType = "wireguard"
	const WgLinkMTU = 1420
	const WgLinkTxQLen = 1000

	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = infName
	linkAttrs.MTU = WgLinkMTU
	linkAttrs.TxQLen = WgLinkTxQLen

	wireguardLink := wgLink{}
	wireguardLink.LinkType = WgLinkType
	wireguardLink.LinkAttrs = &linkAttrs

	handle, err := netlink.NewHandle()
	if err != nil {
		return err
	}

	if err = handle.LinkAdd(netlink.Link(wireguardLink)); err != nil {
		return err
	}

	// Configure wg interface
	pk, err := wgtypes.ParseKey(privateKey)
	if err != nil {
		return err
	}

	config := wgtypes.Config{
		PrivateKey:   &pk,
		ListenPort:   &port,
		ReplacePeers: true,
		Peers:        make([]wgtypes.PeerConfig, 0),
	}

	link, err := handle.LinkByName(infName)
	if err != nil {
		return err
	}

	if err = handle.AddrAdd(link, &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ip,
			Mask: mask,
		},
	}); err != nil {
		return err
	}

	client, err := wgctrl.New()
	if err != nil {
		return err
	}

	if err = client.ConfigureDevice(infName, config); err != nil {
		return err
	}

	link, err = handle.LinkByName(infName)
	if err != nil {
		return err
	}

	if err = netlink.LinkSetUp(link); err != nil {
		return err
	}

	return nil
}

// IsWgInterfaceExists checks if a wireguard interface with the given name exists.
func IsWgInterfaceExists(interfaceName string) bool {

	_, err := netlink.LinkByName(interfaceName)

	return err == nil
}

// AddWgPeer adds a peer to the wireguard interface with the given parameters.
//
// - 'wgInf' is the name of wg interface (e.g. 'wg0')
//
// - 'ip' is the IP address of peer (e.g. '11.0.0.2')
//
// - 'mask' is the subnet mask of peer (e.g. '32')
//
// - 'publicKey' is the public key of peer
//
// - 'preSharedKey' is the pre-shared key of peer (optional)
func AddWgPeer(wgInf string, ip net.IP, mask net.IPMask, publicKey string, preSharedKey *string) error {
	pubKey, err := wgtypes.ParseKey(publicKey)
	if err != nil {
		return err
	}

	peerIpNet := []net.IPNet{{IP: ip, Mask: mask}}

	peer := wgtypes.PeerConfig{
		PublicKey:  pubKey,
		AllowedIPs: peerIpNet,
	}

	if preSharedKey != nil {
		preKey, err := wgtypes.ParseKey(*preSharedKey)
		if err != nil {
			return err
		}
		peer.PresharedKey = &preKey
	}

	client, err := wgctrl.New()
	if err != nil {
		return err
	}

	return client.ConfigureDevice(wgInf, wgtypes.Config{Peers: []wgtypes.PeerConfig{peer}})
}

// RemoveWgPeer removes a peer from the wireguard interface with the given parameters.
//
// - 'wgInf' is the name of wg interface (e.g. 'wg0')
//
// - 'ip' is the IP address of peer (e.g. '11.0.0.2')
//
// - 'mask' is the subnet mask of peer (e.g. '32')
//
// - 'publicKey' is the public key of peer
func RemoveWgPeer(wgInf string, ip net.IP, mask net.IPMask, publicKey string) error {
	pubKey, err := wgtypes.ParseKey(publicKey)
	if err != nil {
		return err
	}

	peerIpNet := []net.IPNet{{IP: ip, Mask: mask}}

	peer := wgtypes.PeerConfig{
		PublicKey:  pubKey,
		AllowedIPs: peerIpNet,
		Remove:     true,
	}

	client, err := wgctrl.New()
	if err != nil {
		return err
	}

	return client.ConfigureDevice(wgInf, wgtypes.Config{Peers: []wgtypes.PeerConfig{peer}})
}

// SetupWgNAT turns on iptables (legacy) NAT for packets forwarding between interfaces.
//
// - wgInf is the name of wg interface (e.g. wg0)
//
// - gwInf is the name of gateway interface (e.g. eth0)
//
// - wgPort is the port number on which wg server listens (e.g. 51820)
//
// - wgNet is the CIDR prefix of wg network (e.g. '11.0.0.0/24')
func SetupWgNAT(wgInf string, gwInf string, wgPort int, wgNet netip.Prefix) error {
	if err := shell.RunBlockingExec(
		"iptables",
		"-t", "nat",
		"-A", "POSTROUTING",
		"-s", wgNet,
		"-o", gwInf,
		"-j", "MASQUERADE",
	); err != nil {
		return err
	}

	if err := shell.RunBlockingExec(
		"iptables",
		"-A", "INPUT",
		"-p", "udp",
		"-m", "udp",
		"--dport", wgPort,
		"-j", "ACCEPT",
	); err != nil {
		return err
	}

	if err := shell.RunBlockingExec(
		"iptables",
		"-A", "FORWARD",
		"-i", wgInf,
		"-j", "ACCEPT",
	); err != nil {
		return err
	}

	return shell.RunBlockingExec(
		"iptables",
		"-A", "FORWARD",
		"-o", wgInf,
		"-j", "ACCEPT",
	)
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
