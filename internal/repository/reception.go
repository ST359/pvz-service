package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/ST359/pvz-service/internal/api"
	errs "github.com/ST359/pvz-service/internal/app_errors"
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
func (r *ReceptionPostgres) GetReceptionInProgress(pvzID uuid.UUID) (uuid.UUID, error) {
	const op = "repository.pvz.ReceptionInProgress"

	var id uuid.UUID
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	err := psql.Select("id").
		From(receptionsTable).
		Where(squirrel.And{squirrel.Eq{"pvz_id": pvzID}, squirrel.Eq{"status": "in_progress"}}).
		RunWith(r.db).
		QueryRow().Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, errs.ErrNoReceptionsInProgress
		}
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

// DeleteLastProduct can return ErrNoProductsInReception
func (r *ReceptionPostgres) DeleteLastProduct(recID uuid.UUID) error {
	const op = "repository.pvz.DeleteLastProduct"

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	var lastProductID uuid.UUID
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	err = psql.Select("id").
		From("products").
		Where(squirrel.Eq{"reception_id": recID}).
		OrderBy("date DESC").
		Limit(1).
		RunWith(tx).
		QueryRow().
		Scan(&lastProductID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrNoProductsInReception
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if lastProductID != uuid.Nil {
		_, err = psql.Delete("products").
			Where(squirrel.Eq{"id": lastProductID}).
			RunWith(tx).
			Exec()

		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
func (r *ReceptionPostgres) CloseLastReception(recID uuid.UUID) (api.Reception, error) {
	const op = "repository.pvz.CloseLastReception"

	var rec api.Reception
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	err := psql.Update(receptionsTable).
		Set("status", "close").
		Where(squirrel.Eq{"id": recID}).
		Suffix("RETURNING id, date, pvz_id, status").
		RunWith(r.db).
		QueryRow().Scan(&rec.Id, &rec.DateTime, &rec.PvzId, &rec.Status)
	if err != nil {
		return api.Reception{}, fmt.Errorf("%s: %w", op, err)
	}
	return rec, nil
}
