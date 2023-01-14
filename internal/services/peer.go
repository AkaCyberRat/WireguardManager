package services

import (
	"context"
	"fmt"

	"WireguardManager/internal/core"
	"WireguardManager/internal/repositories"
	"WireguardManager/internal/tools/network"
)

type PeerService interface {
	Get(ctx context.Context, model *core.GetPeer) (*core.Peer, error)
	Create(ctx context.Context, model *core.CreatePeer) (*core.Peer, error)
	Update(ctx context.Context, model *core.UpdatePeer) (*core.Peer, error)
	Delete(ctx context.Context, model *core.DeletePeer) (*core.Peer, error)
}

// PeerService interface implementation
type Peer struct {
	syncService SyncService
	serverRepos repositories.ServerRepository
	peerRepos   repositories.PeerRepository
	netTool     network.NetworkTool
}

func NewPeerService(serverRepository repositories.ServerRepository, peerRep repositories.PeerRepository, netTool network.NetworkTool, syncService SyncService) *Peer {
	return &Peer{
		serverRepos: serverRepository,
		syncService: syncService,
		peerRepos:   peerRep,
		netTool:     netTool,
	}
}

func (s *Peer) Get(ctx context.Context, model *core.GetPeer) (*core.Peer, error) {
	if !model.Validate() {
		return nil, core.ErrModelValidation
	}

	var peer *core.Peer
	err := s.syncService.InPeerUseContext(model.Id, func() error {
		var err error

		peer, err = s.peerRepos.GetById(model.Id)
		if err != nil {
			if err == core.ErrPeerNotFound {
				return err
			}

			return fmt.Errorf("Peer repository error: %w", err)
		}

		if peer.Status == core.Unused {
			return core.ErrPeerNotFound
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return peer, nil
}

func (s *Peer) Create(ctx context.Context, model *core.CreatePeer) (*core.Peer, error) {
	var peer *core.Peer

	if !model.Validate() {
		return nil, core.ErrModelValidation
	}

	err := s.syncService.InServerUseContext(func() error {
		return s.syncService.InPeerCreateContext(func() error {
			var err error

			count := s.peerRepos.GetPeersCount()
			limit := s.peerRepos.GetPeersLimit()
			if count >= limit {
				return core.ErrPeerLimitReached
			}

			peer, err = s.peerRepos.GetUnused()
			if err != nil {
				return err
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

			server, err := s.serverRepos.Get()
			if err != nil {
				return err
			}

			if peer.Status == core.Enabled && server.Enabled {
				err = s.netTool.EnablePeer(peer)
				if err != nil {
					return err
				}
			}

			_, err = s.peerRepos.Update(peer)
			return err
		})
	})

	if err != nil {
		return nil, err
	}

	return peer, nil
}

func (s *Peer) Update(ctx context.Context, model *core.UpdatePeer) (*core.Peer, error) {
	var peer *core.Peer

	if !model.Validate() {
		return nil, core.ErrModelValidation
	}

	err := s.syncService.InServerUseContext(func() error {
		return s.syncService.InPeerEditContext(model.Id, func() error {
			var err error

			peer, err = s.peerRepos.GetById(model.Id)
			if err != nil {
				if err == core.ErrPeerNotFound {
					return err
				}

				return fmt.Errorf("Peer repository error: %w", err)
			}

			if peer.Status == core.Unused {
				return core.ErrPeerNotFound
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

			server, err := s.serverRepos.Get()
			if err != nil {
				return err
			}

			if server.Enabled {
				if lastStatus == core.Enabled {
					err = s.netTool.DisablePeer(peer)
					if err != nil {
						return err
					}
				}

				if peer.Status == core.Enabled {
					err = s.netTool.EnablePeer(peer)
					if err != nil {
						return err
					}
				}
			}

			_, err = s.peerRepos.Update(peer)
			return err
		})
	})

	if err != nil {
		return nil, err
	}

	return peer, nil
}

func (s *Peer) Delete(ctx context.Context, model *core.DeletePeer) (*core.Peer, error) {
	var peer *core.Peer

	if !model.Validate() {
		return nil, core.ErrModelValidation
	}

	err := s.syncService.InServerUseContext(func() error {
		return s.syncService.InPeerEditContext(model.Id, func() error {
			var err error

			peer, err = s.peerRepos.GetById(model.Id)
			if err != nil {
				if err == core.ErrPeerNotFound {
					return err
				}

				return fmt.Errorf("Peer repository error: %w", err)
			}

			if peer.Status == core.Unused {
				return core.ErrPeerNotFound
			}

			server, err := s.serverRepos.Get()
			if err != nil {
				return err
			}

			if peer.Status == core.Enabled && server.Enabled {
				err = s.netTool.DisablePeer(peer)
				if err != nil {
					return err
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
			return err
		})
	})

	if err != nil {
		return nil, err
	}

	return peer, nil
}
