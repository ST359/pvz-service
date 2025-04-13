package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	errs "github.com/ST359/pvz-service/internal/app_errors"
	"github.com/google/uuid"
)

type UserPostgres struct {
	db *sql.DB
}

func NewUserPostgres(db *sql.DB) *UserPostgres {
	return &UserPostgres{db: db}
}

// Create creates user and returns new user's ID(UUID)
func (u *UserPostgres) Create(email string, password_hash string, role string) (uuid.UUID, error) {
	const op = "repository.user.Create"

	var id uuid.UUID
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	err := psql.Insert(usersTable).
		Columns("email", "password_hash", "role").
		Values(email, password_hash, role).
		Suffix("RETURNING id").
		RunWith(u.db).
		QueryRow().Scan(&id)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

// Login returns password hash and role of a user with given email
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
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", errs.ErrWrongCreds
		}
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	return passHash, role, nil
}
func (u *UserPostgres) EmailExists(email string) (bool, error) {
	const op = "repository.user.EmailExists"

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	var exists bool
	err := psql.Select("COUNT(*)>0").
		From(usersTable).
		Where(squirrel.Eq{"email": email}).
		RunWith(u.db).
		QueryRow().Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return exists, nil
}
