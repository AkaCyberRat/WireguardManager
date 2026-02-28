package wg

import (
	"errors"
	"fmt"
	"net"
	"os"

	"golang.zx2c4.com/wireguard/wgctrl"

	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	WgIpNet = "11.0.0.1/24"
	WgIp    = "11.0.0.1"
)

type WgService struct {
	client *wgctrl.Client
}

func NewWgService() (WgService, error) {
	var empty WgService

	client, err := wgctrl.New()
	if err != nil {
		return empty, err
	}

	return WgService{client: client}, nil
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

	wireguardLink := WgLink{}
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

// CheckWgInterface checks if a wireguard interface with the given name exists.
// Returns nil if the interface exists, returns ErrNotExist if not exist, otherwise returns an error.
func (s *WgService) CheckWgInterface(interfaceName string) error {
	_, err := s.client.Device(interfaceName)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("wireguard interface %s %w", interfaceName, ErrNotExist)
	}

	return err
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

// CheckWgPeer check a peer existence for the wireguard interface with the given parameters.
//
// - 'wgInf' is the name of wg interface (e.g. 'wg0')
//
// - 'ip' is the IP address of peer (e.g. '11.0.0.2')
//
// - 'mask' is the subnet mask of peer (e.g. '32')
//
// - 'publicKey' is the public key of peer
//
// Returns nil if the peer exists, returns ErrNotExist if not exist, otherwise returns an error.
func (s *WgService) CheckWgPeer(wgInf string, ip net.IP, mask net.IPMask) error {
	device, err := s.client.Device(wgInf)
	if err != nil {
		return err
	}

	targetNet := net.IPNet{
		IP:   ip,
		Mask: mask,
	}

	for _, peer := range device.Peers {
		for _, allowed := range peer.AllowedIPs {
			if allowed.IP.Equal(targetNet.IP) &&
				net.IP(allowed.Mask).Equal(net.IP(targetNet.Mask)) {
				return nil
			}
		}
	}

	return fmt.Errorf("peer %w ( Interface: '%s', Ip: '%v', Mask: '%v' )", ErrNotExist, wgInf, ip, mask)
}

var ErrNotExist = fmt.Errorf("not exist")

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

func GeneratePrivateKey() (string, error) {
	prKey, err := wgtypes.GeneratePrivateKey()
	return prKey.String(), err
}

func GeneratePublicKey(privateKey string) (string, error) {
	prKey, err := wgtypes.ParseKey(privateKey)
	if err != nil {
		return "", err
	}

	return prKey.PublicKey().String(), nil
}

//
// Helpers
//

type WgLink struct {
	LinkAttrs *netlink.LinkAttrs
	LinkType  string
}

func (l WgLink) Attrs() *netlink.LinkAttrs {
	return l.LinkAttrs
}

func (l WgLink) Type() string {
	return l.LinkType
}
