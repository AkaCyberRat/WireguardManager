package network

import "WireguardManager/internal/core"

type NetworkTool interface {
	EnablePeer(peer *core.Peer) error
	DisablePeer(peer *core.Peer) error
}

type Deps struct {
	WireguardInterface string
	WireguardIpNet     string
	PrivateKey         string
	Port               int
	UseTC              bool
}

type Tool struct {
	deps Deps
}

func NewNetworkTool(deps Deps) (*Tool, error) {
	tool := Tool{deps}

	err := tool.wgConfigure()
	if err != nil {
		return nil, err
	}

	err = tool.tcConfigure()
	if err != nil {
		return nil, err
	}

	return &tool, nil
}

func (t *Tool) EnablePeer(peer *core.Peer) error {

	err := t.wgPeerUp(peer.Ip, peer.PublicKey, peer.PresharedKey)
	if err != nil {
		return err
	}

	if !t.deps.UseTC {
		return nil
	}

	err = t.tcPeerUp(peer.Ip, peer.DownloadSpeed, peer.UploadSpeed)
	return err
}

func (t *Tool) DisablePeer(peer *core.Peer) error {

	err := t.wgPeerDown(peer.Ip, peer.PublicKey)
	if err != nil {
		return err
	}

	if !t.deps.UseTC {
		return nil
	}

	err = t.tcPeerDown(peer.Ip)
	return err
}
