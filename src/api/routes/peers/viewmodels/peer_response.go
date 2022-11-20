package viewmodels

import (
	"WireguardManager/src/db/models"
	"fmt"
)

type PeerResponse struct {
	Id            string
	IpAddress     string
	PublicKey     string
	PresharedKey  string
	DownloadSpeed int
	UploadSpeed   int
	TrafficAmount int
	Enabled       bool
}

func (model *PeerResponse) BindModel(peer models.Peer) *PeerResponse {
	var enabled bool
	if peer.Status == models.Enabled {
		enabled = true
	} else {
		enabled = false
	}
	model.Id = fmt.Sprint(peer.ID)
	model.IpAddress = peer.IpAddress
	model.PublicKey = peer.PublicKey
	model.PresharedKey = peer.PresharedKey
	model.DownloadSpeed = peer.DownloadSpeed
	model.UploadSpeed = peer.UploadSpeed
	model.TrafficAmount = peer.TrafficAmount
	model.Enabled = enabled

	return model
}
