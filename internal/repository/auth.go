package repository

import (
	"database/sql"

	"github.com/ST359/pvz-service/internal/api"
)

type UserPostgres struct {
	db *sql.DB
}

func NewUserPostgres(db *sql.DB) *UserPostgres {
	return &UserPostgres{db: db}
}

func (a *UserPostgres) Create(user api.PostRegisterJSONBody) (api.User, error) {
	panic("Not implemented")
}
func (a *UserPostgres) Login(credentials api.PostLoginJSONBody) (api.Token, error) {
	panic("Not implemented")
}
