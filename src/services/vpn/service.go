package vpn

import (
	"WireguardManager/src/config"
	"WireguardManager/src/db"
	"WireguardManager/src/db/models"
)

func GetPeersByStatus(statuses []models.Status, offset int, limit int) []models.Peer {
	peers := []models.Peer{}

	db.Instance.Limit(limit).Offset(offset).Where("Status IN ?", statuses).Find(&peers)

	return peers
}

func GetPeersCountByStatus(statuses []models.Status) int {
	val := int64(0)

	db.Instance.Model(&models.Peer{}).Where("Status IN ?", statuses).Count(&val)

	return int(val)
}

func GetUsedPeersCount() int {
	return GetPeersCountByStatus([]models.Status{models.Enabled, models.Disabled})
}

func PeersLimit() int {
	server := models.Server{}

	db.Instance.First(&server)

	return int(server.PeerLimit)
}

func GetPeerById(id int) (models.Peer, error) {
	peer := models.Peer{}

	err := db.Instance.First(&peer, id).Error
	if err != nil {
		return models.Peer{}, err
	}

	return peer, nil
}

func UpdatePeer(model models.Peer) error {
	peer, err := GetPeerById(int(model.ID))
	if err != nil {
		return err
	}

	if peer.Status != model.Status && model.Status == models.Enabled {
		err = enablePeer(model)
	} else if peer.Status != model.Status && peer.Status == models.Enabled {
		err = disablePeer(model)
	}

	if err != nil {
		panic(err)
	}

	peer.PublicKey = model.PublicKey
	peer.PresharedKey = model.PresharedKey
	peer.Status = model.Status
	peer.DownloadSpeed = model.DownloadSpeed
	peer.UploadSpeed = model.UploadSpeed
	peer.TrafficAmount = model.TrafficAmount

	db.Instance.Save(&peer)

	return nil
}

func DropPeerById(id int) error {
	peer, err := GetPeerById(id)
	if err != nil {
		return err
	}

	if peer.Status == models.Enabled {
		disablePeer(peer)
	}

	peer.PublicKey = ""
	peer.PresharedKey = ""
	peer.Status = models.Unused
	peer.DownloadSpeed = 0
	peer.UploadSpeed = 0
	peer.TrafficAmount = 0

	db.Instance.Save(&peer)

	return nil
}

func enablePeer(peer models.Peer) error {
	config := config.Get()

	err := wgPeerUp(peer.IpAddress, peer.PublicKey, peer.PresharedKey)
	if err != nil {
		return err
	}

	if !config.UseTC {
		return nil
	}

	err = tcRulesUp(peer.IpAddress, peer.DownloadSpeed, peer.UploadSpeed)
	return err
}

func disablePeer(peer models.Peer) error {
	err := wgPeerDown(peer.IpAddress, peer.PublicKey)
	if err != nil {
		return err
	}

	err = tcRulesDown(peer.IpAddress)
	return err
}
