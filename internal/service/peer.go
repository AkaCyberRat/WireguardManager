package service

import (
	"context"
	"fmt"

	"WireguardManager/internal/core"
	"WireguardManager/internal/repository"
	"WireguardManager/internal/utility/network"
)

type PeerService interface {
	Get(ctx context.Context, model *core.GetPeer) (*core.Peer, error)
	Create(ctx context.Context, model *core.CreatePeer) (*core.Peer, error)
	Update(ctx context.Context, model *core.UpdatePeer) (*core.Peer, error)
	Delete(ctx context.Context, model *core.DeletePeer) (*core.Peer, error)
}

// PeerService interface implementation
type Peer struct {
	peerRepos repository.PeerRepository
	txManager repository.TransactionManager
	netTool   network.NetworkTool
}

func NewPeerService(peerRep repository.PeerRepository, netTool network.NetworkTool, tm repository.TransactionManager) *Peer {
	return &Peer{peerRepos: peerRep, netTool: netTool, txManager: tm}
}

func (s *Peer) Get(ctx context.Context, model *core.GetPeer) (*core.Peer, error) {
	peer, err := s.peerRepos.GetById(model.Id)
	if err != nil {
		if err == core.ErrPeerNotFound {
			return nil, err
		}

		return nil, fmt.Errorf("Peer repository error: %w", err)
	}

	if peer.Status == core.Unused {
		return nil, core.ErrPeerNotFound
	}

	return peer, nil
}

func (s *Peer) Create(ctx context.Context, model *core.CreatePeer) (*core.Peer, error) {
	var peer *core.Peer
	var err error

	if !model.Validate() {
		return nil, core.ErrModelValidation
	}

	_, err = s.txManager.Transaction(func(tm repository.TransactionManager) (out interface{}, err error) {
		pRepos := tm.GetPeerRepository()

		count := pRepos.GetPeersCount()
		limit := pRepos.GetPeersLimit()
		if count >= limit {
			err = core.ErrPeerLimitReached
			return
		}

		peer, err = pRepos.GetUnused()
		if err != nil {
			return
		}

		peer.PublicKey = model.PublicKey
		peer.PresharedKey = model.PresharedKey
		peer.DownloadSpeed = model.DownloadSpeed
		peer.UploadSpeed = model.UploadSpeed
		if *model.Enabled {
			peer.Status = core.Enabled
		} else {
			peer.Status = core.Disabled
		}

		if peer.Status == core.Enabled {
			err = s.netTool.Enable(peer)
			if err != nil {
				return
			}
		}

		_, err = pRepos.Update(peer)

		return
	})

	return peer, err
}

func (s *Peer) Update(ctx context.Context, model *core.UpdatePeer) (*core.Peer, error) {
	var peer *core.Peer
	var err error

	if !model.Validate() {
		return nil, core.ErrModelValidation
	}

	peer, err = s.peerRepos.GetById(model.Id)
	if err != nil {
		if err == core.ErrPeerNotFound {
			return nil, err
		}

		return nil, fmt.Errorf("Peer repository error: %w", err)
	}

	if peer.Status == core.Unused {
		return nil, core.ErrPeerNotFound
	}

	lastStatus := peer.Status

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
	if model.Enabled != nil {
		if *model.Enabled {
			peer.Status = core.Enabled
		} else {
			peer.Status = core.Disabled
		}
	}

	if lastStatus == core.Enabled {
		err = s.netTool.Disable(peer)
		if err != nil {
			return nil, err
		}
	}
	if peer.Status == core.Enabled {
		err = s.netTool.Enable(peer)
		if err != nil {
			return nil, err
		}
	}

	_, err = s.peerRepos.Update(peer)

	return peer, err
}

func (s *Peer) Delete(ctx context.Context, model *core.DeletePeer) (*core.Peer, error) {
	var peer *core.Peer
	var err error

	if !model.Validate() {
		return nil, core.ErrModelValidation
	}

	peer, err = s.peerRepos.GetById(model.Id)
	if err != nil {
		if err == core.ErrPeerNotFound {
			return nil, err
		}

		return nil, fmt.Errorf("Peer repository error: %w", err)
	}

	if peer.Status == core.Unused {
		return nil, core.ErrPeerNotFound
	}

	if peer.Status == core.Enabled {
		err = s.netTool.Disable(peer)
		if err != nil {
			return nil, err
		}
	}

	rewritedPeer := core.Peer{}
	rewritedPeer.Id = peer.Id
	rewritedPeer.Ip = peer.Ip
	rewritedPeer.PublicKey = ""
	rewritedPeer.PresharedKey = ""
	rewritedPeer.DownloadSpeed = 0
	rewritedPeer.UploadSpeed = 0
	rewritedPeer.Status = core.Unused

	_, err = s.peerRepos.Update(&rewritedPeer)
	if err != nil {
		return nil, err
	}

	return peer, nil
}
