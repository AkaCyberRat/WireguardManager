package repositories

import "WireguardManager/internal/core"

type ServerRepository interface {
	Get() (*core.Server, error)
	Save(model *core.Server) (*core.Server, error)
}
