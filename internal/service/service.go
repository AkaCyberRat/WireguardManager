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

type InitDeps struct {
	PeerInitDeps PeerInitDeps
}

func (s *Services) Init(deps InitDeps) error {
	err := s.PeerService.Init(deps.PeerInitDeps)

	if err != nil {
		return err
	}
}
