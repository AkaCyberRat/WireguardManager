package vpn

import (
	"WireguardManager/src/config"
	"WireguardManager/src/db"
	"WireguardManager/src/db/models"
	"fmt"

	"github.com/sirupsen/logrus"
)

func Configure() {
	err := wgConfigure()
	if err != nil {
		logrus.Fatal("Failed to configure wg: ", err.Error())
	}

	err = tcConfigure()
	if err != nil {
		logrus.Fatal("Failed to configure tc: ", err.Error())
	}

	err = dbConfigure()
	if err != nil {
		logrus.Fatal("Failed to configure db: ", err.Error())
	}

	err = upEnabledPeers()
	if err != nil {
		logrus.Fatal("Failed to up enabled peers: ", err.Error())
	}
}

func dbConfigure() (err error) {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		err = errors.New(fmt.Sprint(r))
	// 	}
	// }()

	db.Instance.AutoMigrate(&models.Server{}, &models.Peer{})
	if err != nil {
		return err
	}

	if !isInit() {
		config := config.Get()

		// Add server model
		server := models.Server{
			IpAddress:  "10.0.0.1",
			PeerLimit:  int64(config.PeerLimit),
			PrivateKey: "",
			PublicKey:  "",
		}
		err = db.Instance.Create(&server).Error
		if err != nil {
			return err
		}

		// Add peer models
		for i := 2; i <= config.PeerLimit; i++ {

			oct3 := 0xff & (i >> 8)
			oct4 := 0xff & i

			peer := models.Peer{
				IpAddress:     fmt.Sprintf("10.0.%v.%v", oct3, oct4),
				PublicKey:     "",
				PresharedKey:  "",
				DownloadSpeed: 0,
				UploadSpeed:   0,
				Status:        models.Unused,
			}

			err = db.Instance.Create(&peer).Error
			if err != nil {
				return
			}

			logrus.Infof("Create peer in db %v", peer.IpAddress)
		}

	}

	return nil
}

func upEnabledPeers() error {
	config := config.Get()

	peers := []models.Peer{}
	db.Instance.Where(&models.Peer{Status: models.Status(models.Enabled)}).Find(&peers)

	for _, peer := range peers {
		err := wgPeerUp(peer.IpAddress, peer.PublicKey, peer.PresharedKey)
		if err != nil {
			return err
		}

		if config.UseTC {
			err = tcRulesUp(peer.IpAddress, peer.DownloadSpeed, peer.UploadSpeed)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isInit() bool {
	server := models.Server{}
	savedPeersCount := int64(0)

	db.Instance.FirstOrInit(&server, models.Server{PeerLimit: 0})
	if server.PeerLimit == 0 {
		return false
	}

	db.Instance.Model(&models.Peer{}).Count(&savedPeersCount)
	return server.PeerLimit != savedPeersCount
}
