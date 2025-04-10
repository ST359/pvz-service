package repository

import (
	"database/sql"

	"github.com/ST359/pvz-service/internal/api"
	"github.com/google/uuid"
)

type User interface {
	Create(email string, password_hash string, role string) (string, error)
	//Login returns password hash and role of a user with given email
	Login(email string) (string, string, error)
}
type PVZ interface {
	Create(pvz api.PVZ) (api.PVZ, error)
	GetByDate(params api.GetPvzParams) ([]api.PVZInfo, error)
	CloseLastReception(pvzID uuid.UUID) (api.Reception, error)
	DeleteLastProduct(pvzID uuid.UUID) error
}
type Reception interface {
	Create(pvzID uuid.UUID) (api.Reception, error)
	AddProduct(pvzId uuid.UUID, product api.ProductType) (api.Product, error)
}
type Repository struct {
	User
	PVZ
	Reception
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		User:      NewUserPostgres(db),
		PVZ:       NewPVZPostgres(db),
		Reception: NewReceptionPostgres(db),
	}
}
