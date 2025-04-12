package repository

import (
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
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
	const op = "repository.reception.Create"

	var rec api.Reception
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	err := psql.Insert(receptionsTable).
		Columns("pvz_id").
		Values(pvzID).
		Suffix("RETURNING id, date, pvz_id, status").
		RunWith(r.db).
		QueryRow().Scan(&rec.Id, &rec.DateTime, &rec.PvzId, &rec.Status)
	if err != nil {
		return api.Reception{}, fmt.Errorf("%s: %w", op, err)
	}
	return rec, nil
}
func (r *ReceptionPostgres) AddProduct(recID uuid.UUID, prodType api.ProductType) (api.Product, error) {
	const op = "repository.reception.AddProduct"

	var prod api.Product
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	err := psql.Insert(productsTable).
		Columns("reception_id", "type").
		Values(recID, prodType).
		Suffix("RETURNING id, date, reception_id, type").
		RunWith(r.db).
		QueryRow().Scan(&prod.Id, &prod.DateTime, &prod.ReceptionId, &prod.Type)
	if err != nil {
		return api.Product{}, fmt.Errorf("%s: %w", op, err)
	}
	return prod, nil
}
