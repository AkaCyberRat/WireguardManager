package network

import (
	"WireguardManager/internal/core"

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
	prKey, err := wgtypes.ParseKey(privateKey)
	if err != nil {
		return "", err
	}

	return prKey.PublicKey().String(), nil
}

func (t *Tool) GeneratePrivateKey() (string, error) {
	pubKey, err := wgtypes.GeneratePrivateKey()
	return pubKey.String(), err
}
