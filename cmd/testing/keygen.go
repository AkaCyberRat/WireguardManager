package main

import (
	"WireguardManager/internal/logging"
	"WireguardManager/internal/tools/wg"

	"github.com/sirupsen/logrus"
)

func main() {
	logging.SetTempConfiguration()

	privateKey, err := wg.GeneratePrivateKey()
	if err != nil {
		logrus.Fatal("Failed to generate private key:", err.Error())
	}

	publicKey, err := wg.GeneratePublicKey(privateKey)
	if err != nil {
		logrus.Fatal("Failed to generate public key:", err.Error())
	}

	logrus.Infof("PrivateKey:\t%v", privateKey)
	logrus.Infof("PublicKey:\t%v", publicKey)
}
