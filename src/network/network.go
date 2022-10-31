package network

import (
	"github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl"
)

func Configure() {
	client, err := wgctrl.New()
	if err != nil {
		logrus.Fatalf("Failed to create wgctrl client: %v", err.Error())
	}

	err = CreateWgInf()
	if err != nil {
		logrus.Fatalf("Failed to create wg interface: %v", err.Error())
	}

	err = ConfigureWgInf(client)
	if err != nil {
		logrus.Fatal("Failed to configure wg interface: ", err.Error())
	}

	// err = network.SetWgPeers(client, GeneratePeers(client), false)
	// if err != nil {
	// 	logrus.Fatal("Cant set wg peers: ", err.Error())
	// }

	err = UpWgInf()
	if err != nil {
		logrus.Fatal("Failed to up wg interface: ", err.Error())
	}

	err = TcConfigure()
	if err != nil {
		logrus.Fatal("Failed to configure tc: ", err.Error())
	}
}
