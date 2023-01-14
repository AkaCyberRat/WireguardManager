package service

import (
	"context"

	"WireguardManager/internal/core"
	"WireguardManager/internal/repository"
	"WireguardManager/internal/utility/network"
)

type ServerService interface {
	Get(ctx context.Context) (*core.Server, error)
	Update(ctx context.Context, model *core.UpdateServer) (*core.Server, error)
}

type Server struct {
	syncService    SyncService
	serverRepos    repository.ServerRepository
	netTool        network.NetworkTool
	recoverService RecoverService
}

func NewServerService(serverRepos repository.ServerRepository, netTool network.NetworkTool, syncService SyncService, recoverService RecoverService) *Server {
	return &Server{serverRepos: serverRepos, netTool: netTool, syncService: syncService, recoverService: recoverService}
}

func (s *Server) Get(ctx context.Context) (*core.Server, error) {
	var server *core.Server

	err := s.syncService.InServerUseContext(func() error {
		var err error

		server, err = s.serverRepos.Get()
		return err
	})

	return server, err
}

func (s *Server) Update(ctx context.Context, model *core.UpdateServer) (*core.Server, error) {
	var server *core.Server

	if !model.Validate() {
		return nil, core.ErrModelValidation
	}

	err := s.syncService.InServerEditContext(func() error {
		var wasEnabled bool
		var err error

		server, err = s.serverRepos.Get()
		if err != nil {
			return err
		}

		if model.PrivateKey != nil {
			publicKey, err := s.netTool.GeneratePublicKey(*model.PrivateKey)
			if err != nil {
				return err
			}

			server.PrivateKey = *model.PrivateKey
			server.PublicKey = publicKey
		}

		wasEnabled = server.Enabled
		if model.Enabled != nil {
			server.Enabled = *model.Enabled
		}

		if wasEnabled {
			if err = s.netTool.DisableServer(); err != nil {
				return err
			}
		}

		if server.Enabled {
			if err = s.netTool.EnableServer(server); err != nil {
				return err
			}

			if err = s.recoverService.RecoverPeers(); err != nil {
				return err
			}
		}

		_, err = s.serverRepos.Save(server)
		return err
	})
	if err != nil {
		return nil, err
	}

	return server, nil
}
