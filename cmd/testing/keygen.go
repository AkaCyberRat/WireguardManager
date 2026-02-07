package main

import (
	"WireguardManager/internal/logging"
	"WireguardManager/internal/tools/network"
	"github.com/sirupsen/logrus"
)

func main() {
	logging.SetTempConfiguration()

	privateKey, err := network.GeneratePrivateKey()
	if err != nil {
		logrus.Fatal("Failed to generate private key:", err.Error())
	}

	publicKey, err := network.GeneratePublicKey(privateKey)
	if err != nil {
		logrus.Fatal("Failed to generate public key:", err.Error())
	}

	logrus.Infof("PrivateKey:\t%v", privateKey)
	logrus.Infof("PublicKey:\t%v", publicKey)
}
