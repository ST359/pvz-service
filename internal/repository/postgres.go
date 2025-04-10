package repository

import (
	"database/sql"
	"fmt"

	"github.com/ST359/pvz-service/internal/config"
	_ "github.com/lib/pq"
)

const (
	usersTable    = "users"
	pvzTable      = "pvzs"
	receiptsTable = "receipts"
	productsTable = "products"
)

func NewPostgresDB(cfg *config.Config) (*sql.DB, error) {
	const op = "storage.postgres.New"

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbPassword, cfg.DbName)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return db, nil
}
