package wg

import (
	"errors"
	"fmt"
	"net"
	"os"

	"WireguardManager/internal/tools"

	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Tool struct {
	client *wgctrl.Client
}

func NewWgTool() (Tool, error) {
	client, err := wgctrl.New()
	if err != nil {
		return Tool{}, err
	}

	return Tool{client: client}, nil
}

type ServerParams struct {
	// InfName is the name of wg interface (e.g. 'wg0')
	InfName string
	// Ip is the IP address of wg interface (e.g. '11.0.0.1')
	Ip net.IP
	// Mask is the subnet mask of wg interface (e.g. '24')
	Mask net.IPMask
	// PrivateKey is the private key of wg server
	PrivateKey string
	// Port is the port number on which wg server listens (e.g. 51820)
	Port int
}

// AddWgInterface creates and configures a wireguard interface with the given parameters.
func (t *Tool) AddWgInterface(params ServerParams) error {
	const WgLinkType = "wireguard"
	const WgLinkMTU = 1420
	const WgLinkTxQLen = 1000

	handle, err := netlink.NewHandle()
	if err != nil {
		return err
	}
	defer handle.Close()

	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = params.InfName
	linkAttrs.MTU = WgLinkMTU
	linkAttrs.TxQLen = WgLinkTxQLen

	wireguardLink := WgLink{}
	wireguardLink.LinkType = WgLinkType
	wireguardLink.LinkAttrs = &linkAttrs

	if err = handle.LinkAdd(netlink.Link(wireguardLink)); err != nil {
		return err
	}

	pk, err := wgtypes.ParseKey(params.PrivateKey)
	if err != nil {
		return err
	}

	config := wgtypes.Config{
		PrivateKey:   &pk,
		ListenPort:   &params.Port,
		ReplacePeers: true,
		Peers:        make([]wgtypes.PeerConfig, 0),
	}

	link, err := handle.LinkByName(params.InfName)
	if err != nil {
		return err
	}

	if err = handle.AddrAdd(link, &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   params.Ip,
			Mask: params.Mask,
		},
	}); err != nil {
		return err
	}

	if err = t.client.ConfigureDevice(params.InfName, config); err != nil {
		return err
	}

	if err = handle.LinkSetUp(link); err != nil {
		return err
	}

	return nil
}

// CheckWgInterface checks if a wireguard interface with the given name exists.
// Returns nil if the interface exists, returns ErrNotExist if not exist, otherwise returns an error.
func (t *Tool) CheckWgInterface(params ServerParams) error {
	_, err := t.client.Device(params.InfName)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("wireguard interface %t %w", params.InfName, tools.ErrNotExist)
	}

	return err
}

type PeerParams struct {
	// InfName is the name of wg interface (e.g. 'wg0')
	InfName string
	// Ip is the IP address of peer (e.g. '11.0.0.2')
	Ip net.IP
	// Mask is the subnet mask of peer (e.g. '32')
	Mask net.IPMask
	// PublicKey is the public key of peer
	PublicKey string
	// PreSharedKey is the pre-shared key of peer (optional)
	PreSharedKey *string
}

// AddWgPeer adds a peer to the wireguard interface with the given parameters.
func (t *Tool) AddWgPeer(params PeerParams) error {
	pubKey, err := wgtypes.ParseKey(params.PublicKey)
	if err != nil {
		return err
	}

	peerIpNet := []net.IPNet{{IP: params.Ip, Mask: params.Mask}}

	peer := wgtypes.PeerConfig{
		PublicKey:  pubKey,
		AllowedIPs: peerIpNet,
	}

	if params.PreSharedKey != nil {
		preKey, err := wgtypes.ParseKey(*params.PreSharedKey)
		if err != nil {
			return err
		}
		peer.PresharedKey = &preKey
	}

	return t.client.ConfigureDevice(params.InfName, wgtypes.Config{Peers: []wgtypes.PeerConfig{peer}})
}

type CheckPeerParams PeerParams

// CheckWgPeer check a peer existence for the wireguard interface with the given parameters.
// Returns nil if the peer exists, returns ErrNotExist if not exist, otherwise returns an error.
func (t *Tool) CheckWgPeer(params CheckPeerParams) error {
	device, err := t.client.Device(params.InfName)
	if err != nil {
		return err
	}

	targetNet := net.IPNet{
		IP:   params.Ip,
		Mask: params.Mask,
	}

	for _, peer := range device.Peers {
		for _, allowed := range peer.AllowedIPs {
			if allowed.IP.Equal(targetNet.IP) &&
				net.IP(allowed.Mask).Equal(net.IP(targetNet.Mask)) {
				return nil
			}
		}
	}

	return fmt.Errorf("peer %w ( Interface: '%s', Ip: '%v', Mask: '%v' )", tools.ErrNotExist, params.InfName, params.Ip, params.Mask)
}

type RemovePeerParams struct {
	// InfName is the name of wg interface (e.g. 'wg0')
	InfName string
	// Ip is the IP address of peer (e.g. '11.0.0.2')
	Ip net.IP
	// Mask is the subnet mask of peer (e.g. '32')
	Mask net.IPMask
}

// RemoveWgPeer removes a peer from the wireguard interface with the given parameters.
func (t *Tool) RemoveWgPeer(params RemovePeerParams) error {
	peerIpNet := []net.IPNet{{IP: params.Ip, Mask: params.Mask}}

	peer := wgtypes.PeerConfig{
		AllowedIPs: peerIpNet,
		Remove:     true,
	}

	return t.client.ConfigureDevice(params.InfName, wgtypes.Config{Peers: []wgtypes.PeerConfig{peer}})
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
