package services

import (
	"WireguardManager/internal/core"
	"WireguardManager/internal/repositories"
	"WireguardManager/internal/tools/network"

	"github.com/sirupsen/logrus"
)

type RecoverService interface {
	RecoverServer() error
	RecoverPeers() error
	RecoverAll() error
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

	logrus.Infof("Recover service recover server complete.")
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

	logrus.Infof("Recover service recover peers complete.")
	return nil
}

func (r *Recover) RecoverAll() error {
	if err := r.RecoverServer(); err != nil {
		return err
	}

	if err := r.RecoverPeers(); err != nil {
		return err
	}

	logrus.Infof("Recover service full recovering complete")
	return nil
}
