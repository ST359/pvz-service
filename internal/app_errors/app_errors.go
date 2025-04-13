package app_errors

import "errors"

var (
	ErrPasswordTooLong = errors.New("Password is too long, should be less than 72 bytes")
	ErrEmailExists     = errors.New("User with this email already exists")
	ErrWrongCreds      = errors.New("Wrong email or password")
)
