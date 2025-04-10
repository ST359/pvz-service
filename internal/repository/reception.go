package repository

import (
	"database/sql"

	"github.com/ST359/pvz-service/internal/api"
	"github.com/google/uuid"
)

type ReceptionPostgres struct {
	db *sql.DB
}

func NewReceptionPostgres(db *sql.DB) *ReceptionPostgres {
	return &ReceptionPostgres{db: db}
}

func (r *ReceptionPostgres) Create(pvzID uuid.UUID) (api.Reception, error) {
	panic("not implemented")
}
func (r *ReceptionPostgres) AddProduct(pvzId uuid.UUID, product api.ProductType) (api.Product, error) {
	panic("not implemented")
}
