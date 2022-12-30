package repository

import "WireguardManager/internal/core"

type PeerRepository interface {
	GetPeersLimit() int
	GetPeersCount() int

	Add(model *core.Peer) (*core.Peer, error)
	Update(model *core.Peer) (*core.Peer, error)
	GetById(id string) (*core.Peer, error)
	GetUnused() (*core.Peer, error)
}
