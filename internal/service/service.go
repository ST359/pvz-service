package service

import (
	"github.com/ST359/pvz-service/internal/api"
	"github.com/ST359/pvz-service/internal/repository"
	"github.com/google/uuid"
)

type User interface {
	CreateUser(usr api.PostRegisterJSONBody) (api.User, error)
	Login(creds api.PostLoginJSONBody) (string, error)
	ParseToken(tok string) (string, error)
	GenerateToken(role string) (string, error)
}

type Reception interface {
	Create(pvzID uuid.UUID) (api.Reception, error)
	AddProduct(pvzID uuid.UUID, product api.ProductType) (api.Product, error)
	GetReceptionInProgress(pvzID uuid.UUID) (uuid.UUID, error)
	DeleteLastProduct(recID uuid.UUID) error
	CloseLastReception(pvzID uuid.UUID) (api.Reception, error)
}

type PVZ interface {
	Create(pvz api.PVZ) (api.PVZ, error)
	GetByDate(params api.GetPvzParams) ([]api.PVZInfo, error)
}
type Service struct {
	User
	PVZ
	Reception
}

func NewService(repo *repository.Repository) *Service {
	return &Service{
		User:      NewUserService(repo.User),
		PVZ:       NewPVZService(repo.PVZ),
		Reception: NewReceptionService(repo.Reception),
	}
}
