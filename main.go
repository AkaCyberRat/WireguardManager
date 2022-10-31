package main

import (
	"net"

	"WireguardManager/src/logging"
	"WireguardManager/src/network"

	"github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func main() {

	logging.Configure()
	network.Configure()

	err := network.TcConfigure()
	if err != nil {
		logrus.Fatal("Cant configure tc tool: ", err.Error())
	}

	for i := 0; i < 1; i++ {
		err = network.TcAddRules("10.1.1.2/32", 10, 10, 1)
		if err != nil {
			logrus.Fatal("Cant configure tc tool: ", err.Error())
		}
	}

	select {}
}

func GeneratePeers(client *wgctrl.Client) *[]wgtypes.PeerConfig {

	prKey, _ := wgtypes.ParseKey("uKIkAl5agqGLoodeDAdtgZHh91vXck5z/mmxETx2dWs=")
	pubKey := prKey.PublicKey()
	ipAddress := "10.1.1.2/32"

	logrus.Infof("Client pk: %v", prKey.String())
	logrus.Infof("Client pubk: %v", prKey.PublicKey().String())

	var ipAddresses []net.IPNet
	_, ipnet, err := net.ParseCIDR(ipAddress)
	if err != nil {
		logrus.Fatal("Cant parse ip net address for peer")
	}
	ipAddresses = append(ipAddresses, *ipnet)

	peer := wgtypes.PeerConfig{
		PublicKey:  pubKey,
		AllowedIPs: ipAddresses,
	}

	var peers []wgtypes.PeerConfig
	peers = append(peers, peer)

	return &peers
}
