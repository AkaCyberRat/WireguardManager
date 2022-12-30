package service

import (
	"WireguardManager/internal/repository"
	"WireguardManager/internal/utility/network"
)

type Deps struct {
	NetTool            network.NetworkTool
	PeerRepos          repository.PeerRepository
	TransactionManager repository.TransactionManager
}

type Services struct {
	PeerService PeerService
}

func NewServices(deps Deps) Services {

	services := Services{
		PeerService: NewPeerService(deps.PeerRepos, deps.NetTool, deps.TransactionManager),
	}

	return services
}
