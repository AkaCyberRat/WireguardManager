package sqlite

import (
	"strconv"

	"WireguardManager/internal/core"
	"WireguardManager/internal/repository/sqlite/models"

	"gorm.io/gorm"
)

type PeerRepository struct {
	db *gorm.DB
}

func NewSqlitePeerRepository(db *gorm.DB) *PeerRepository {
	return &PeerRepository{db: db}
}

func (r PeerRepository) GetPeersLimit() int {
	var count int64

	err := r.db.Model(&models.Peer{}).Count(&count).Error
	if err != nil {
		panic(err)
	}

	return int(count)
}

func (r PeerRepository) GetPeersCount() int {
	var count int64

	err := r.db.Model(&models.Peer{}).Where("Status IN ?", []int{1, 2}).Count(&count).Error
	if err != nil {
		panic(err)
	}

	return int(count)
}

func (r PeerRepository) Add(model *core.Peer) (*core.Peer, error) {
	var peer models.Peer

	model.Id = "0"
	peer.FromCore(model)

	err := r.db.Create(&peer).Error

	model = peer.ToCore()
	return model, err
}

func (r PeerRepository) Update(model *core.Peer) (*core.Peer, error) {
	if !validId(model.Id) {
		return nil, core.ErrPeerNotFound
	}

	var peer models.Peer
	peer.FromCore(model)

	err := r.db.Save(&peer).Error
	if err != nil {
		return nil, err
	}

	model = peer.ToCore()
	return model, err
}

func (r PeerRepository) GetAll() ([]*core.Peer, error) {
	var peers []models.Peer

	err := r.db.Model(&models.Peer{}).Find(&peers).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, core.ErrPeerNotFound
		}

		return nil, err
	}

	models := []*core.Peer{}
	for _, peer := range peers {
		model := peer.ToCore()
		models = append(models, model)
	}

	return models, err
}

func (r PeerRepository) GetById(id string) (*core.Peer, error) {
	if !validId(id) {
		return nil, core.ErrPeerNotFound
	}

	intId, _ := strconv.Atoi(id)

	// peer := models.Peer{Id: uint(intId)}
	peer := models.Peer{}
	err := r.db.Model(&models.Peer{}).Where("id = ?", intId).First(&peer).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, core.ErrPeerNotFound
		}

		return nil, err
	}

	model := peer.ToCore()
	return model, err
}

func (r PeerRepository) GetUnused() (*core.Peer, error) {
	var peer models.Peer

	err := r.db.Model(&models.Peer{}).Where("Status IN ?", []int{0}).First(&peer).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, core.ErrPeerNotFound
		}

		return nil, err
	}

	model := peer.ToCore()
	return model, err
}

func validId(id string) bool {
	_, err := strconv.Atoi(id)

	return err == nil
}

// func (r PeerRepository) Remove(id string) (*core.Peer, error) {
// 	if !validId(id) {
// 		return nil, core.ErrPeerNotFound
// 	}

// 	intId, _ := strconv.Atoi(id)

// 	peer := models.Peer{Id: uint(intId)}
// 	err := r.db.Delete(&peer).Error
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, core.ErrPeerNotFound
// 		}

// 		return nil, err
// 	}

// 	model := peer.ToCore()
// 	return model, err
// }
