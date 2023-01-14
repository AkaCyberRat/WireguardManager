package sqlite

import (
	"WireguardManager/internal/core"
	"WireguardManager/internal/repository/sqlite/models"
	"WireguardManager/internal/utility/network"
	"fmt"

	"gorm.io/gorm"
)

type Repositories struct {
	PeerRepository   *PeerRepository
	ServerRepository *ServerRepository
}

func NewRepositories(db *gorm.DB) *Repositories {
	db.AutoMigrate(models.Peer{}, models.Server{})

	repositories := Repositories{
		PeerRepository:   NewSqlitePeerRepository(db),
		ServerRepository: NewSqliteServerRepository(db),
	}

	return &repositories
}

type InitDeps struct {
	NetTool network.NetworkTool

	WireguardPrivateKey string
	WireguardPort       int
	WireguardEnabled    bool

	PeerCount int
}

func (r *Repositories) Init(deps InitDeps) error {

	if r.PeerRepository.GetPeersCount() == 0 && r.PeerRepository.GetPeersLimit() == 0 {
		if err := r.initPeers(deps.PeerCount); err != nil {
			return err
		}

		if err := r.initServer(deps.NetTool, deps.WireguardPrivateKey, deps.WireguardPort, deps.WireguardEnabled); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repositories) initPeers(count int) error {

	for i := 1; i <= count; i++ {
		oct3 := 255 & (i >> 8)
		oct4 := 255 & i

		peer := core.Peer{
			Id:     fmt.Sprintf("%d", i),
			Ip:     fmt.Sprintf("10.0.%d.%d", oct3, oct4),
			Status: core.Unused,
		}

		if _, err := r.PeerRepository.Add(&peer); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repositories) initServer(netTool network.NetworkTool, privateKey string, port int, enabled bool) error {
	var err error

	if privateKey == "" {
		if privateKey, err = netTool.GeneratePrivateKey(); err != nil {
			return err
		}
	}

	publicKey, err := netTool.GeneratePublicKey(privateKey)
	if err != nil {
		return err
	}

	server := core.Server{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Enabled:    enabled,
	}

	_, err = r.ServerRepository.Save(&server)
	return err
}
