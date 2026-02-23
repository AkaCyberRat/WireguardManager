package network

import (
	"WireguardManager/internal/core"
	"WireguardManager/pkg/shell"
	"errors"
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type NetworkTool interface {
	EnableServer(peer *core.Server) error
	DisableServer() error

	EnablePeer(peer *core.Peer) error
	DisablePeer(peer *core.Peer) error

	GeneratePublicKey(privateKey string) (string, error)
	GeneratePrivateKey() (string, error)
}

type Tool struct {
	interfaceName string
	port          int
	wgClient      *wgctrl.Client
}

func NewNetworkTool(port int) *Tool {
	wgClient, err := wgctrl.New()
	if err != nil {
		panic("can't create wg client")
	}

	tool := Tool{interfaceName: "wg0", port: port, wgClient: wgClient}

	return &tool
}

func (t *Tool) EnableServer(server *core.Server) error {

	if err := t.wgServerUp(server.PrivateKey); err != nil {
		return err
	}

	if err := t.tcServerUp(); err != nil {
		return err
	}

	logrus.Debugf("Server enabled. [IpNet=%v, InterfaceName=%v, Port=%v]", WgIpNet, t.interfaceName, t.port)
	return nil
}

func (t *Tool) DisableServer() error {

	if err := t.tcServerDown(); err != nil {
		return err
	}

	if err := t.wgServerDown(); err != nil {
		return err
	}

	logrus.Debugf("Server disabled. [IpNet=%v, InterfaceName=%v, Port=%v]", WgIpNet, t.interfaceName, t.port)
	return nil
}

func (t *Tool) EnablePeer(peer *core.Peer) error {

	if err := t.wgPeerUp(peer.Ip, peer.PublicKey, peer.PresharedKey); err != nil {
		return err
	}

	if err := t.tcPeerUp(peer.Ip, peer.DownloadSpeed, peer.UploadSpeed); err != nil {
		return err
	}

	logrus.Debugf("Peer enabled. [Id=%v, Ip=%v, PubKey=%v]", peer.Id, peer.Ip, peer.PublicKey)
	return nil
}

func (t *Tool) DisablePeer(peer *core.Peer) error {

	if err := t.wgPeerDown(peer.Ip, peer.PublicKey); err != nil {
		return err
	}

	if err := t.tcPeerDown(peer.Ip); err != nil {
		return err
	}

	logrus.Debugf("Peer disabled. [Id=%v, Ip=%v, PubKey=%v]", peer.Id, peer.Ip, peer.PublicKey)
	return nil
}

func (t *Tool) GeneratePublicKey(privateKey string) (string, error) {
	return GeneratePublicKey(privateKey)
}

func (t *Tool) GeneratePrivateKey() (string, error) {
	return GeneratePrivateKey()
}

func GeneratePublicKey(privateKey string) (string, error) {
	prKey, err := wgtypes.ParseKey(privateKey)
	if err != nil {
		return "", err
	}

	return prKey.PublicKey().String(), nil
}

func GeneratePrivateKey() (string, error) {
	prKey, err := wgtypes.GeneratePrivateKey()
	return prKey.String(), err
}

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

func (t *Tool) tcServerUp() error {
	//
	// Add tc base rule to limit client download bandwidth (server upload)
	//
	_, err := shell.RunExecWithTimeout(fmt.Sprintf("tc qdisc add dev %s root handle 1: htb", t.interfaceName))
	if err != nil {
		return err
	}

	//
	// Add tc base rule to limit client upload bandwidth (server download)
	//
	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc qdisc add dev %s ingress", t.interfaceName))
	if err != nil {
		return err
	}

	logrus.Tracef("Traffic control rules for server enabled.")
	return nil
}

func (t *Tool) tcServerDown() error {
	//
	// Add tc base rule to limit client download bandwidth (server upload)
	//
	_, err := shell.RunExecWithTimeout(fmt.Sprintf("tc qdisc del dev %s root handle 1: htb", t.interfaceName))
	if err != nil {
		return err
	}

	//
	// Add tc base rule to limit client upload bandwidth (server download)
	//
	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc qdisc del dev %s ingress", t.interfaceName))
	if err != nil {
		return err
	}

	logrus.Tracef("Traffic control rules for server disabled.")
	return nil
}

func (t *Tool) tcPeerUp(ip string, downloadSpeed int, uploadSpeed int) error {
	ind := getIpIndex(ip)

	//
	// Limit download bandwidth
	//
	_, err := shell.RunExecWithTimeout(fmt.Sprintf("tc class add dev %s parent 1: classid 1:%v htb rate %vmbit ceil %vmbit", t.interfaceName, ind, downloadSpeed, downloadSpeed))
	if err != nil {
		return err
	}

	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc filter add dev %s protocol ip parent 1: prio %v u32 match ip src %v flowid 1:%v", t.interfaceName, ind, ip, ind))
	if err != nil {
		return err
	}

	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc filter add dev %s protocol ip parent 1: prio %v u32 match ip dst %v flowid 1:%v", t.interfaceName, ind, ip, ind))
	if err != nil {
		return err
	}

	//
	// Limit upload bandwidth
	//
	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc filter add dev %s protocol ip ingress prio %v u32 match ip src %v action police rate %vmbit burst 5mbit", t.interfaceName, ind, ip, uploadSpeed))
	if err != nil {
		return err
	}

	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc filter add dev %s protocol ip ingress prio %v u32 match ip dst %v action police rate %vmbit burst 5mbit", t.interfaceName, ind, ip, uploadSpeed))
	if err != nil {
		return err
	}

	logrus.Tracef("Traffic control rules for peer enabled. [Ip=%v]", ip)
	return nil
}

func (t *Tool) tcPeerDown(ip string) error {
	ind := getIpIndex(ip)

	commands := []string{
		fmt.Sprintf("tc filter del dev %s parent 1: prio %v", t.interfaceName, ind),
		fmt.Sprintf("tc filter del dev %s ingress prio %v", t.interfaceName, ind),
		fmt.Sprintf("tc class del dev %s parent 1: classid 1:%v", t.interfaceName, ind),
	}

	for _, command := range commands {
		_, err := shell.RunExecWithTimeout(command)
		if err != nil {
			return err
		}
	}

	logrus.Tracef("Traffic control rules for peer disabled. [Ip=%v]", ip)
	return nil
}
