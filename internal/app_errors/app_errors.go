package app_errors

import "errors"

var (
	ErrPasswordTooLong = errors.New("password is too long, should be less than 72 bytes")
	ErrEmailExists     = errors.New("user with this email already exists")
	ErrWrongCreds      = errors.New("wrong email or password")

	ErrNoReceptionsInProgress = errors.New("no receptions in progress")
	ErrNoProductsInReception  = errors.New("no products in this reception")
	ErrReceptionNotClosed     = errors.New("there is reception in progress")
)
