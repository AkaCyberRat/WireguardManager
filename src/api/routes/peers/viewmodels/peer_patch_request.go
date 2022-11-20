package viewmodels

import (
	"WireguardManager/src/db/models"
	"reflect"
)

type PeerPatchRequest struct {
	PublicKey     *string `validate:"omitempty,base64"`
	PresharedKey  *string `validate:"omitempty,base64"`
	DownloadSpeed *int    `validate:"omitempty,numeric,min=0"`
	UploadSpeed   *int    `validate:"omitempty,numeric,min=0"`
	TrafficAmount *int    `validate:"omitempty,numeric,min=0"`
	Enable        *bool   `validate:"omitempty,required_without=Disable"`
	Disable       *bool   `validate:"omitempty,required_without=Enable"`
}

func (model *PeerPatchRequest) BindPeer(peer *models.Peer) {
	if model.PublicKey != nil {
		peer.PublicKey = *model.PublicKey
	}

	if model.PresharedKey != nil {
		peer.PresharedKey = *model.PresharedKey
	}

	if model.DownloadSpeed != nil {
		peer.DownloadSpeed = *model.DownloadSpeed
	}

	if model.UploadSpeed != nil {
		peer.UploadSpeed = *model.UploadSpeed
	}

	if model.TrafficAmount != nil {
		peer.TrafficAmount = *model.TrafficAmount
	}

	if model.Enable != nil {
		peer.Status = models.Enabled
	}

	if model.TrafficAmount != nil {
		peer.Status = models.Disabled
	}
}

func HasNotNullField(p interface{}) bool {

	t := reflect.TypeOf(p)

	for i := 0; i < t.NumField(); i++ {
		if !reflect.ValueOf(t.Field(i)).IsNil() {
			return true
		}
	}

	return false
}
