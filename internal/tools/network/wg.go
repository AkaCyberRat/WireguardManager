package network

import (
	"fmt"
	"net"
	"net/netip"
	"os"

	"golang.zx2c4.com/wireguard/wgctrl"

	"github.com/coreos/go-iptables/iptables"
	"github.com/google/shlex"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	WgIpNet = "11.0.0.1/24"
	WgIp    = "11.0.0.1"
)

type WgService struct {
	client *wgctrl.Client
	ipt    *iptables.IPTables
}

func NewWgService() (WgService, error) {
	var empty WgService

	client, err := wgctrl.New()
	if err != nil {
		return empty, err
	}

	ipt, err := iptables.New()
	if err != nil {
		return empty, err
	}

	return WgService{client: client, ipt: ipt}, nil
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
func (s *WgService) SetupWgInterface(infName string, ip net.IP, mask net.IPMask, privateKey string, port int) error {
	const WgLinkType = "wireguard"
	const WgLinkMTU = 1420
	const WgLinkTxQLen = 1000

	handle, err := netlink.NewHandle()
	if err != nil {
		return err
	}
	defer handle.Close()

	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = infName
	linkAttrs.MTU = WgLinkMTU
	linkAttrs.TxQLen = WgLinkTxQLen

	wireguardLink := wgLink{}
	wireguardLink.LinkType = WgLinkType
	wireguardLink.LinkAttrs = &linkAttrs

	if err = handle.LinkAdd(netlink.Link(wireguardLink)); err != nil {
		return err
	}

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

	if err = s.client.ConfigureDevice(infName, config); err != nil {
		return err
	}

	if err = handle.LinkSetUp(link); err != nil {
		return err
	}

	return nil
}

// IsWgInterfaceExists checks if a wireguard interface with the given name exists.
func (s *WgService) IsWgInterfaceExists(interfaceName string) (bool, error) {
	_, err := s.client.Device(interfaceName)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
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
func (s *WgService) AddWgPeer(wgInf string, ip net.IP, mask net.IPMask, publicKey string, preSharedKey *string) error {
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

	return s.client.ConfigureDevice(wgInf, wgtypes.Config{Peers: []wgtypes.PeerConfig{peer}})
}

// IsWgPeerExists check a peer existence for the wireguard interface with the given parameters.
//
// - 'wgInf' is the name of wg interface (e.g. 'wg0')
//
// - 'ip' is the IP address of peer (e.g. '11.0.0.2')
//
// - 'mask' is the subnet mask of peer (e.g. '32')
//
// - 'publicKey' is the public key of peer
func (s *WgService) IsWgPeerExists(wgInf string, ip net.IP, mask net.IPMask, publicKey string) (bool, error) {
	pubKey, err := wgtypes.ParseKey(publicKey)
	if err != nil {
		return false, err
	}

	device, err := s.client.Device(wgInf)
	if err != nil {
		return false, err
	}

	targetNet := net.IPNet{
		IP:   ip,
		Mask: mask,
	}

	for _, peer := range device.Peers {
		if peer.PublicKey != pubKey {
			continue
		}

		for _, allowed := range peer.AllowedIPs {
			if allowed.IP.Equal(targetNet.IP) &&
				net.IP(allowed.Mask).Equal(net.IP(targetNet.Mask)) {
				return true, nil
			}
		}
	}

	return false, nil
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
func (s *WgService) RemoveWgPeer(wgInf string, ip net.IP, mask net.IPMask, publicKey string) error {
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

	return s.client.ConfigureDevice(wgInf, wgtypes.Config{Peers: []wgtypes.PeerConfig{peer}})
}

// SetupWgNAT turns on ipt (legacy) NAT for packets forwarding between interfaces.
//
// - wgInf is the name of wg interface (e.g. wg0)
//
// - gwInf is the name of gateway interface (e.g. eth0)
//
// - wgPort is the port number on which wg server listens (e.g. 51820)
//
// - wgNet is the CIDR prefix of wg network (e.g. '11.0.0.0/24')
func (s *WgService) SetupWgNAT(wgInf string, gwInf string, wgPort int, wgNet netip.Prefix) error {
	commands := wgNatCommands(wgInf, gwInf, wgPort, wgNet)

	for _, command := range commands {
		args, err := shlex.Split(command)
		if err != nil {
			return err
		}

		if err = s.ipt.AppendUnique(args[1], args[3], args[4:]...); err != nil {
			return err
		}
	}

	return nil
}

// IsWgNatExists checks ipt (legacy) NAT rules existence.
//
// - wgInf is the name of wg interface (e.g. wg0)
//
// - gwInf is the name of gateway interface (e.g. eth0)
//
// - wgPort is the port number on which wg server listens (e.g. 51820)
//
// - wgNet is the CIDR prefix of wg network (e.g. '11.0.0.0/24')
func (s *WgService) IsWgNatExists(wgInf string, gwInf string, wgPort int, wgNet netip.Prefix) (bool, error) {
	commands := wgNatCommands(wgInf, gwInf, wgPort, wgNet)

	for _, command := range commands {
		args, err := shlex.Split(command)
		if err != nil {
			return false, err
		}

		if exists, err := s.ipt.Exists(args[1], args[3], args[4:]...); !exists || err != nil {
			return false, err
		}
	}

	return true, nil
}

// DeleteWgNat delete ipt (legacy) NAT rules if exists.
//
// - wgInf is the name of wg interface (e.g. wg0)
//
// - gwInf is the name of gateway interface (e.g. eth0)
//
// - wgPort is the port number on which wg server listens (e.g. 51820)
//
// - wgNet is the CIDR prefix of wg network (e.g. '11.0.0.0/24')
func (s *WgService) DeleteWgNat(wgInf string, gwInf string, wgPort int, wgNet netip.Prefix) error {
	commands := wgNatCommands(wgInf, gwInf, wgPort, wgNet)

	for _, command := range commands {
		args, err := shlex.Split(command)
		if err != nil {
			return err
		}

		if err = s.ipt.DeleteIfExists(args[1], args[3], args[4:]...); err != nil {
			return err
		}
	}

	return nil
}

func wgNatCommands(wgInf string, gwInf string, wgPort int, wgNet netip.Prefix) []string {
	return []string{
		// NAT for wg interface
		fmt.Sprintf("-t nat -A POSTROUTING -s %v -o %v -j MASQUERADE", wgNet, gwInf),
		fmt.Sprintf("-t filter -A INPUT -p udp -m udp --dport %v -j ACCEPT", wgPort),
		fmt.Sprintf("-t filter -A FORWARD -i %v -j ACCEPT", wgInf),
		fmt.Sprintf("-t filter -A FORWARD -o %v -j ACCEPT", wgInf),

		// DNS redirect
		fmt.Sprintf("-t nat -A PREROUTING -i %v -p udp --dport 53 -j DNAT --to-destination 8.8.8.8:53", wgInf),
		fmt.Sprintf("-t nat -A PREROUTING -i %v -p tcp --dport 53 -j DNAT --to-destination 8.8.8.8:53", wgInf),
	}
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
