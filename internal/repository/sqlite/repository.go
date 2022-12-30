package sqlite

import (
	"WireguardManager/internal/core"
	"WireguardManager/internal/repository/sqlite/models"
	"fmt"

	"gorm.io/gorm"
)

type InitDeps struct {
	WireguardInterfacePublicKey string
	WireguardInterfaceIp        string

	PeerCount int
}

type Repositories struct {
	PeerRepos          *PeerRepository
	TransactionManager *TransactionManager
}

func NewRepositories(db *gorm.DB) Repositories {
	db.AutoMigrate(models.Peer{})

	repositories := Repositories{
		PeerRepos:          NewSqlitePeerRepository(db),
		TransactionManager: NewSqliteTransactionManager(db),
	}

	return repositories
}

func (r Repositories) Init(deps InitDeps) (Repositories, error) {

	if r.PeerRepos.GetPeersCount() == 0 && r.PeerRepos.GetPeersLimit() == 0 {
		err := InitPeers(deps.PeerCount, r.PeerRepos)

		return r, err
	}

	return r, nil
}

func InitPeers(count int, r *PeerRepository) error {

	for i := 1; i <= count; i++ {
		oct3 := 255 & (i >> 8)
		oct4 := 255 & i

		peer := core.Peer{
			Id:     fmt.Sprintf("%d", i),
			Ip:     fmt.Sprintf("10.0.%d.%d", oct3, oct4),
			Status: core.Unused,
		}

		_, err := r.Add(&peer)
		if err != nil {
			return err
		}
	}

	return nil
}
