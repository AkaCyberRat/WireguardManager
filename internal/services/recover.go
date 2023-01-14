package services

import (
	"WireguardManager/internal/core"
	"WireguardManager/internal/repositories"
	"WireguardManager/internal/tools/network"
)

type RecoverService interface {
	RecoverServer() error
	RecoverPeers() error
}

type Recover struct {
	peerRepos   repositories.PeerRepository
	serverRepos repositories.ServerRepository
	netTool     network.NetworkTool
}

func NewRecoverService(peerRepos repositories.PeerRepository, serverRepos repositories.ServerRepository, netTool network.NetworkTool) RecoverService {
	return &Recover{
		peerRepos:   peerRepos,
		serverRepos: serverRepos,
		netTool:     netTool,
	}
}

func (r *Recover) RecoverServer() error {
	server, err := r.serverRepos.Get()
	if err != nil {
		return err
	}

	if server.Enabled {
		if err = r.netTool.EnableServer(server); err != nil {
			return err
		}
	}

	return nil
}

func (r *Recover) RecoverPeers() error {
	peers, err := r.peerRepos.GetAll()
	if err != nil {
		return err
	}

	for _, peer := range peers {
		if peer.Status == core.Enabled {
			if err = r.netTool.EnablePeer(peer); err != nil {
				return err
			}
		}
	}

	return nil
}
