package repository

import (
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
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
	const op = "repository.pvz.Create"

	var resPVZ api.PVZ
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	err := psql.Insert(pvzTable).
		Columns("pvz_id").
		Values(pvz.City).
		Suffix("RETURNING id, registration_date, city").
		RunWith(p.db).
		QueryRow().Scan(&resPVZ.Id, &resPVZ.RegistrationDate, &resPVZ.City)
	if err != nil {
		return api.PVZ{}, fmt.Errorf("%s: %w", op, err)
	}
	return resPVZ, nil
}
func (p *PVZPostgres) GetByDate(params api.GetPvzParams) ([]api.PVZInfo, error) {
	const op = "repository.pvz.GetByDate"

	panic("not implemented")
}
func (p *PVZPostgres) CloseLastReception(recID uuid.UUID) (api.Reception, error) {
	const op = "repository.pvz.CloseLastReception"

	var rec api.Reception
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	err := psql.Update(receptionsTable).
		Set("status", "close").
		Where(squirrel.Eq{"id": recID}).
		Suffix("RETURNING id, date, pvz_id, status").
		RunWith(p.db).
		QueryRow().Scan(&rec.Id, &rec.DateTime, &rec.PvzId, &rec.Status)
	if err != nil {
		return api.Reception{}, fmt.Errorf("%s: %w", op, err)
	}
	return rec, nil
}

func (p *PVZPostgres) DeleteLastProduct(recID uuid.UUID) error {
	const op = "repository.pvz.DeleteLastProduct"

	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	var lastProductID uuid.UUID
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	err = psql.Select("id").
		From("products").
		Where("reception_id = ?", recID).
		OrderBy("date DESC").
		Limit(1).
		RunWith(tx).
		QueryRow().
		Scan(&lastProductID)

	if err != nil {
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

func (p *PVZPostgres) GetReceptionInProgress(pvzID uuid.UUID) (uuid.UUID, error) {
	const op = "repository.pvz.ReceptionInProgress"

	var id uuid.UUID
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	err := psql.Select("id").
		From(receptionsTable).
		Where(squirrel.And{squirrel.Eq{"pvz_id": pvzID}, squirrel.Eq{"status": "in_progress"}}).
		RunWith(p.db).
		QueryRow().Scan(&id)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}
