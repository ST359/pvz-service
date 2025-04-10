package repository

import (
	"github.com/ST359/pvz-service/internal/api"
	"github.com/google/uuid"
)

type Authorization interface {
	Create(user api.PostRegisterJSONBody) (api.User, error)
	Login(credentials api.PostLoginJSONBody) (api.Token, error)
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
