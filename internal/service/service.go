package service

import "github.com/ST359/pvz-service/internal/api"

type User interface {
	CreateUser(usr api.PostRegisterJSONBody) (api.User, error)
	Login(creds api.PostLoginJSONBody) (string, error)
	ParseToken(tok string) (string, error)
}
