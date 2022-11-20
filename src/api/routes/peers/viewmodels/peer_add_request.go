package viewmodels

import "WireguardManager/src/db/models"

type PeerAddRequest struct {
	PublicKey     string `validate:"required,base64"`
	PresharedKey  string `validate:"omitempty,base64"`
	DownloadSpeed int    `validate:"required,numeric,min=0"`
	UploadSpeed   int    `validate:"required,numeric,min=0"`
	Enabled       *bool  `validate:"required"`
}

func (model *PeerAddRequest) BindPeer(peer *models.Peer) {
	var status models.Status
	if *model.Enabled {
		status = models.Enabled
	} else {
		status = models.Disabled
	}

	peer.PublicKey = model.PublicKey
	peer.PresharedKey = model.PresharedKey
	peer.DownloadSpeed = model.DownloadSpeed
	peer.UploadSpeed = model.UploadSpeed
	peer.Status = status
}
