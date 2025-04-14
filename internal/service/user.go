package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/ST359/pvz-service/internal/api"
	errs "github.com/ST359/pvz-service/internal/app_errors"
	"github.com/ST359/pvz-service/internal/repository"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

const (
	secretKey = "secretKey"
	tokenTTL  = 24 * time.Hour
)

type tokenClaims struct {
	jwt.StandardClaims
	UserRole string `json:"role"`
}
type UserService struct {
	repo repository.User
}

func NewUserService(repo repository.User) *UserService {
	return &UserService{repo: repo}
}

// CreateUser return an api.User on success
func (u *UserService) CreateUser(usr api.PostRegisterJSONBody) (api.User, error) {
	const op = "service.user.CreateUser"

	if len(usr.Password) >= 72 {
		return api.User{}, errs.ErrPasswordTooLong
	}
	exists, err := u.repo.EmailExists(string(usr.Email))
	if err != nil {
		return api.User{}, fmt.Errorf("%s: %w", op, err)
	}
	if exists {
		return api.User{}, errs.ErrEmailExists
	}

	var createdUser api.User
	passHash := generatePasswordHash(usr.Password)

	id, err := u.repo.Create(string(usr.Email), passHash, string(usr.Role))
	if err != nil {
		return api.User{}, fmt.Errorf("%s: %w", op, err)
	}
	createdUser.Email = usr.Email
	createdUser.Id = &id
	createdUser.Role = api.UserRole(usr.Role)
	return createdUser, nil
}

// Login returns a JWT token on success
func (u *UserService) Login(creds api.PostLoginJSONBody) (string, error) {
	const op = "service.user.Login"

	passHash, role, err := u.repo.Login(string(creds.Email))
	if err != nil {
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passHash), []byte(creds.Password)); err != nil {
		return "", errs.ErrWrongCreds
	}
	tok, err := u.GenerateToken(role)
	if err != nil {
		return "", fmt.Errorf("%s: error generating jwt: %w", op, err)
	}
	return tok, nil
}

// ParseToken returns a role of a user on success
func (u *UserService) ParseToken(tok string) (string, error) {

	token, err := jwt.ParseWithClaims(tok, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(secretKey), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return "", errs.ErrWrongCreds
	}

	return claims.UserRole, nil
}

func (u *UserService) GenerateToken(role string) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		role,
	})

	return token.SignedString([]byte(secretKey))
}

// generatePasswordHash return a hash, password must be less than 72 bytes long
func generatePasswordHash(password string) string {
	//ignoring err because password recieved should be <72 bytes(check made in CreateUser)
	bPas, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bPas)
}
