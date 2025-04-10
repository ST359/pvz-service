package repository

import (
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
)

type UserPostgres struct {
	db *sql.DB
}

func NewUserPostgres(db *sql.DB) *UserPostgres {
	return &UserPostgres{db: db}
}

func (u *UserPostgres) Create(email string, password_hash string, role string) (string, error) {
	const op = "repository.user.Create"

	var id string
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	err := psql.Insert(usersTable).
		Columns("email", "password_hash", "role").
		Values(email, password_hash, role).
		Suffix("RETURNING \"id\"").
		RunWith(u.db).
		QueryRow().Scan(&id)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return id, nil

}
func (u *UserPostgres) Login(email string) (string, string, error) {
	const op = "repository.user.Login"

	var passHash, role string
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	err := psql.Select("password_hash", "role").
		From(usersTable).
		Where(squirrel.Eq{"email": email}).
		RunWith(u.db).
		QueryRow().Scan(&passHash, &role)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	return passHash, role, nil
}
