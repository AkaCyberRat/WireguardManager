//go:build !linux

package tc

import (
	"errors"
	"net"

	"github.com/vishvananda/netlink"
)

type TcTool struct {
	handle *netlink.Handle
}

func NewTcTool() (TcTool, error) {
	return TcTool{}, ErrUnsupportedPlatform
}

func (t *TcTool) Close() {
	t.handle.Close()
}

func (t *TcTool) SetupTcBase(wgInf, ifbInf string) error {
	return ErrUnsupportedPlatform
}

func (t *TcTool) CheckTcBaseRules(wgInf, ifbInf string) error {
	return ErrUnsupportedPlatform
}

type TcPeerParams struct {
	WgInf             string
	IfbInf            string
	PeerIp            net.IP
	ServerNetworkMask net.IPMask
	DownloadSpeedMb   uint
	UploadSpeedMb     uint
}

func (t *TcTool) ApplyTcForPeer(params TcPeerParams) error {
	return ErrUnsupportedPlatform
}

func (t *TcTool) CheckTcPeerRules(params TcPeerParams) error {
	return ErrUnsupportedPlatform
}

func (t *TcTool) DiscardTcForPeer(params TcPeerParams) error {
	return ErrUnsupportedPlatform
}

var ErrUnsupportedPlatform = errors.New("unsupported platform")
