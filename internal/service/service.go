package service

import (
	"github.com/ST359/pvz-service/internal/api"
	"github.com/google/uuid"
)

type User interface {
	CreateUser(usr api.PostRegisterJSONBody) (api.User, error)
	Login(creds api.PostLoginJSONBody) (string, error)
	ParseToken(tok string) (string, error)
}
type Reception interface {
	Create(pvzID uuid.UUID) (api.Reception, error)
	AddProduct(pvzID uuid.UUID, product api.ProductType) (api.Product, error)
	GetReceptionInProgress(pvzID uuid.UUID) (uuid.UUID, error)
	DeleteLastProduct(recID uuid.UUID) error
}
