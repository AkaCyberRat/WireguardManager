package models

import (
	"fmt"
	"strconv"

	"WireguardManager/internal/core"
)

type Peer struct {
	Id            uint `gorm:"primarykey"`
	Ip            string
	PublicKey     string
	PresharedKey  string
	DownloadSpeed uint
	UploadSpeed   uint
	Status        uint
}

func (p *Peer) ToCore() *core.Peer {
	var peer core.Peer

	peer.Id = fmt.Sprint(p.Id)
	peer.Ip = p.Ip
	peer.PublicKey = p.PublicKey
	peer.PresharedKey = p.PresharedKey
	peer.DownloadSpeed = int(p.DownloadSpeed)
	peer.UploadSpeed = int(p.UploadSpeed)

	switch p.Status {
	case 0:
		peer.Status = core.Unused
	case 1:
		peer.Status = core.Enabled
	case 2:
		peer.Status = core.Disabled
	default:
		panic("unexpected status value")
	}

	return &peer
}

func (p *Peer) FromCore(peer *core.Peer) {
	if peer == nil {
		panic("nil argument exception")
	}

	id, _ := strconv.Atoi(peer.Id)

	p.Id = uint(id)
	p.Ip = peer.Ip
	p.PublicKey = peer.PublicKey
	p.PresharedKey = peer.PresharedKey
	p.DownloadSpeed = uint(peer.DownloadSpeed)
	p.UploadSpeed = uint(peer.UploadSpeed)

	switch peer.Status {
	case core.Unused:
		p.Status = 0
	case core.Enabled:
		p.Status = 1
	case core.Disabled:
		p.Status = 2
	default:
		panic("unexpected status value")
	}
}
