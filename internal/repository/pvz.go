package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/ST359/pvz-service/internal/api"
	"github.com/google/uuid"
)

const (
	defaultLimit  = 10
	defaultOffset = 0
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

	// Prepare parameters
	limit := defaultLimit
	if params.Limit != nil {
		limit = *params.Limit
	}

	offset := defaultOffset
	if params.Page != nil {
		offset = (*params.Page - 1) * limit
	}

	// Prepare date parameters
	var startDate, endDate interface{}
	if params.StartDate != nil {
		startDate = *params.StartDate
	} else {
		startDate = nil
	}
	if params.EndDate != nil {
		endDate = *params.EndDate
	} else {
		endDate = nil
	}

	rows, err := p.db.Query(
		"SELECT * FROM get_pvz_with_receptions_paginated($1, $2, $3, $4)",
		startDate,
		endDate,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var result []api.PVZInfo

	for rows.Next() {
		var (
			pvzID          uuid.UUID
			city           string
			regDate        time.Time
			receptionsJSON []byte
		)

		if err := rows.Scan(&pvzID, &city, &regDate, &receptionsJSON); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		var receptionInfos []api.ReceptionInfo
		if err := json.Unmarshal(receptionsJSON, &receptionInfos); err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal receptions: %w", op, err)
		}

		pvzInfo := api.PVZInfo{
			Pvz: &api.PVZ{
				Id:               (*uuid.UUID)(&pvzID),
				City:             api.PVZCity(city),
				RegistrationDate: &regDate,
			},
			Receptions: &receptionInfos,
		}

		result = append(result, pvzInfo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return result, nil
}
