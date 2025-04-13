package repository

import (
	"database/sql"

	"github.com/ST359/pvz-service/internal/api"
	"github.com/google/uuid"
)

type User interface {
	//Create creates a user and returns an id of the user
	Create(email string, password_hash string, role string) (uuid.UUID, error)
	//Login returns password hash and role of a user with given email
	Login(email string) (string, string, error)
	EmailExists(email string) (bool, error)
}
type PVZ interface {
	Create(pvz api.PVZ) (api.PVZ, error)
	GetByDate(params api.GetPvzParams) ([]api.PVZInfo, error)
	CloseLastReception(pvzID uuid.UUID) (api.Reception, error)
}
type Reception interface {
	Create(pvzID uuid.UUID) (api.Reception, error)
	AddProduct(pvzID uuid.UUID, product api.ProductType) (api.Product, error)
	GetReceptionInProgress(pvzID uuid.UUID) (uuid.UUID, error)
	DeleteLastProduct(recID uuid.UUID) error
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
