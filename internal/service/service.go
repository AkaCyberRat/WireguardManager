package service

import (
	"WireguardManager/internal/repository"
	"WireguardManager/internal/utility/network"
)

type Deps struct {
	NetTool          network.NetworkTool
	PeerRepository   repository.PeerRepository
	ServerRepository repository.ServerRepository
}

type Services struct {
	PeerService    PeerService
	ServerService  ServerService
	RecoverService RecoverService
	SyncService    SyncService
}

func NewServices(deps Deps) Services {

	syncService := NewSyncService()
	recoverService := NewRecoverService(deps.PeerRepository, deps.ServerRepository, deps.NetTool)
	peerService := NewPeerService(deps.ServerRepository, deps.PeerRepository, deps.NetTool, syncService)
	serverService := NewServerService(deps.ServerRepository, deps.NetTool, syncService, recoverService)

	services := Services{
		PeerService:    peerService,
		ServerService:  serverService,
		RecoverService: recoverService,
		SyncService:    syncService,
	}

	return services
}
