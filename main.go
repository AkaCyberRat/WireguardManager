package main

import (
	"WireguardManager/src/api"
	"WireguardManager/src/config"
	"WireguardManager/src/db"
	"WireguardManager/src/db/models"
	"WireguardManager/src/logging"
	"WireguardManager/src/services/vpn"

	"github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func main() {
	logging.TempConfig()
	config.Load()

	logging.Configure()
	db.Configure()
	vpn.Configure()
	api.Configure()

	GeneratePeers()

	api.Run()
}

func GeneratePeers() {
	GeneratePeer(1, "uKIkAl5agqGLoodeDAdtgZHh91vXck5z/mmxETx2dWs=")

	pk, _ := wgtypes.GeneratePrivateKey()
	GeneratePeer(2, pk.String())
}

func GeneratePeer(index int, publicKey string) {

	prKey, _ := wgtypes.ParseKey(publicKey)

	peer, err := vpn.GetPeerById(index)
	if err != nil {
		logrus.Fatal("Failed to set peer: ", err)
	}

	peer.PublicKey = prKey.PublicKey().String()
	peer.PresharedKey = ""
	peer.DownloadSpeed = 10
	peer.UploadSpeed = 10
	peer.Status = models.Enabled

	logrus.Infof("Client pk: %v", prKey.String())
	logrus.Infof("Client pubk: %v", prKey.PublicKey().String())
	logrus.Infof("Client ip: %v", peer.IpAddress)
	logrus.Infof("Client status: %v", peer.Status)

	vpn.UpdatePeer(peer)
}
