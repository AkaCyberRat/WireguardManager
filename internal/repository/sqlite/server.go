package sqlite

import (
	"WireguardManager/internal/core"
	"WireguardManager/internal/repository/sqlite/models"

	"gorm.io/gorm"
)

type ServerRepository struct {
	db *gorm.DB
}

func NewSqliteServerRepository(db *gorm.DB) *ServerRepository {
	return &ServerRepository{db: db}
}

func (r *ServerRepository) Get() (*core.Server, error) {
	var server models.Server

	err := r.db.First(&server).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, core.ErrServerNotFound
		}

		return nil, err
	}

	model := server.ToCore()
	return model, err
}
func (r *ServerRepository) Save(model *core.Server) (*core.Server, error) {

	var server models.Server

	server.FromCore(model)
	server.Id = 1

	err := r.db.Save(&server).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, core.ErrServerNotFound
		}

		return nil, err
	}

	return model, err
}
