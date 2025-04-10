package repository

import (
	"database/sql"

	"github.com/ST359/pvz-service/internal/api"
	"github.com/google/uuid"
)

type PVZPostgres struct {
	db *sql.DB
}

func NewPVZPostgres(db *sql.DB) *PVZPostgres {
	return &PVZPostgres{db: db}
}
func (p *PVZPostgres) Create(pvz api.PVZ) (api.PVZ, error) {
	panic("not implemented")
}
func (p *PVZPostgres) GetByDate(params api.GetPvzParams) ([]api.PVZInfo, error) {
	panic("not implemented")
}
func (p *PVZPostgres) CloseLastReception(pvzID uuid.UUID) (api.Reception, error) {
	panic("not implemented")
}
func (p *PVZPostgres) DeleteLastProduct(pvzID uuid.UUID) error {
	panic("not implemented")
}
