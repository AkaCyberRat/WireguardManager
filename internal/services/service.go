package services

import (
	"WireguardManager/internal/config"
	"WireguardManager/internal/repositories"
	"WireguardManager/internal/tools/network"
)

type Deps struct {
	NetTool          network.NetworkTool
	PeerRepository   repositories.PeerRepository
	ServerRepository repositories.ServerRepository
	Config           config.Configuration
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
	serverService := NewServerService(ServerDeps{
		ServerRepository: deps.ServerRepository,
		SyncService:      syncService,
		RecoverService:   recoverService,
		NetTool:          deps.NetTool,
		Config:           deps.Config,
	})

	services := Services{
		PeerService:    peerService,
		ServerService:  serverService,
		RecoverService: recoverService,
		SyncService:    syncService,
	}

	return services
}
