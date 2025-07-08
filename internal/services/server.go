package services

import (
	"context"

	"WireguardManager/internal/config"
	"WireguardManager/internal/core"
	"WireguardManager/internal/repositories"
	"WireguardManager/internal/tools/network"

	"github.com/sirupsen/logrus"
)

type ServerService interface {
	Get(ctx context.Context) (*core.ResponseServer, error)
	Update(ctx context.Context, model *core.UpdateServer) (*core.ResponseServer, error)
}

type Server struct {
	syncService    SyncService
	recoverService RecoverService
	serverRepos    repositories.ServerRepository
	netTool        network.NetworkTool
	config         config.Configuration
}

type ServerDeps struct {
	ServerRepository repositories.ServerRepository
	SyncService      SyncService
	RecoverService   RecoverService
	NetTool          network.NetworkTool
	Config           config.Configuration
}

func NewServerService(deps ServerDeps) *Server {
	return &Server{
		serverRepos:    deps.ServerRepository,
		syncService:    deps.SyncService,
		recoverService: deps.RecoverService,
		netTool:        deps.NetTool,
		config:         deps.Config,
	}
}

func (s *Server) Get(ctx context.Context) (*core.ResponseServer, error) {
	var server *core.Server

	err := s.syncService.InServerUseContext(func() error {
		var err error

		server, err = s.serverRepos.Get()
		return err
	})

	response := core.ResponseServer{
		HostIp:    s.config.Host.Ip,
		DnsIp:     network.WgIp,
		PublicKey: server.PublicKey,
		Port:      s.config.Wireguard.Port,
		Enabled:   server.Enabled,
	}

	logrus.Infof("Server service get server.")
	return &response, err
}

func (s *Server) Update(ctx context.Context, model *core.UpdateServer) (*core.ResponseServer, error) {
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

	response := core.ResponseServer{
		HostIp:    s.config.Host.Ip,
		DnsIp:     network.WgIp,
		PublicKey: server.PublicKey,
		Port:      s.config.Wireguard.Port,
		Enabled:   server.Enabled,
	}

	logrus.Infof("Server service update server.")
	return &response, nil
}
