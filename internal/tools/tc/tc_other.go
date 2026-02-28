//go:build !linux

package tc

import (
	"errors"
	"net"
)

var ErrUnsupportedPlatform = errors.New("unsupported platform")

func SetupTcBase(wgInf, ifbInf string) error {
	return ErrUnsupportedPlatform
}

func CheckTcBaseRules(wgInf, ifbInf string) error {
	return ErrUnsupportedPlatform
}

func ApplyTcForPeer(wgInf, ifbInf string, peerIp net.IP, serverNetworkMask net.IPMask, downloadSpeedMb, uploadSpeedMb uint) error {
	return ErrUnsupportedPlatform
}

func CheckTcPeerRules(wgInf, ifbInf string, peerIp net.IP, serverNetworkMask net.IPMask, downloadSpeedMb, uploadSpeedMb uint) error {
	return ErrUnsupportedPlatform
}

func DiscardTcForPeer(wgInf string, peerIp net.IP, serverNetworkMask net.IPMask) error {
	return ErrUnsupportedPlatform
}
