package handler_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ST359/pvz-service/internal/api"
	errs "github.com/ST359/pvz-service/internal/app_errors"
	"github.com/ST359/pvz-service/internal/handler"
	"github.com/ST359/pvz-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of service.User
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(creds api.PostRegisterJSONBody) (api.User, error) {
	args := m.Called(creds)
	return args.Get(0).(api.User), args.Error(1)
}

func (m *MockUserService) Login(creds api.PostLoginJSONBody) (string, error) {
	args := m.Called(creds)
	return args.String(0), args.Error(1)
}

func (m *MockUserService) ParseToken(tok string) (api.UserRole, error) {
	args := m.Called(tok)
	return args.Get(0).(api.UserRole), args.Error(1)
}

func (m *MockUserService) GenerateToken(role string) (string, error) {
	args := m.Called(role)
	return args.String(0), args.Error(1)
}

func setupRouter(h *handler.Handler) *gin.Engine {
	router := gin.Default()
	router.POST("/dummy-login", h.DummyLogin)
	router.POST("/login", h.Login)
	router.POST("/register", h.Register)
	return router
}

func TestDummyLogin_Success(t *testing.T) {
	mockUserService := new(MockUserService)
	mockUserService.On("GenerateToken", string(api.PostDummyLoginJSONBodyRoleEmployee)).Return("test-token", nil)

	h := &handler.Handler{
		Services: &service.Service{User: mockUserService},
		Logger:   slog.Default(),
	}
	router := setupRouter(h)

	reqBody := api.PostDummyLoginJSONBody{Role: api.PostDummyLoginJSONBodyRoleEmployee}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/dummy-login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response api.Token
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-token", string(response))
	mockUserService.AssertExpectations(t)
}

func TestDummyLogin_InvalidRole(t *testing.T) {
	h := &handler.Handler{
		Logger: slog.Default(),
	}
	router := setupRouter(h)

	reqBody := map[string]string{"role": "invalid"}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/dummy-login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response api.Error
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Bad request", response.Message)
}

func TestLogin_Success(t *testing.T) {
	mockUserService := new(MockUserService)
	email := openapi_types.Email("test@example.com")
	creds := api.PostLoginJSONBody{
		Email:    email,
		Password: "password",
	}
	mockUserService.On("Login", creds).Return("test-token", nil)

	h := &handler.Handler{
		Services: &service.Service{User: mockUserService},
		Logger:   slog.Default(),
	}
	router := setupRouter(h)

	body, _ := json.Marshal(creds)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response api.Token
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-token", string(response))
	mockUserService.AssertExpectations(t)
}

func TestRegister_Success(t *testing.T) {
	mockUserService := new(MockUserService)
	email := openapi_types.Email("new@example.com")
	creds := api.PostRegisterJSONBody{
		Email:    email,
		Password: "password",
		Role:     api.Employee,
	}
	userID := uuid.New()
	user := api.User{
		Email: email,
		Id:    &userID,
		Role:  api.UserRoleEmployee,
	}
	mockUserService.On("CreateUser", creds).Return(user, nil)

	h := &handler.Handler{
		Services: &service.Service{User: mockUserService},
		Logger:   slog.Default(),
	}
	router := setupRouter(h)

	body, _ := json.Marshal(creds)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response api.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, user.Email, response.Email)
	assert.Equal(t, *user.Id, *response.Id)
	assert.Equal(t, user.Role, response.Role)
	mockUserService.AssertExpectations(t)
}

func TestRegister_EmailExists(t *testing.T) {
	mockUserService := new(MockUserService)
	email := openapi_types.Email("exists@example.com")
	creds := api.PostRegisterJSONBody{
		Email:    email,
		Password: "password",
		Role:     api.Employee,
	}
	mockUserService.On("CreateUser", creds).Return(api.User{}, errs.ErrEmailExists)

	h := &handler.Handler{
		Services: &service.Service{User: mockUserService},
		Logger:   slog.Default(),
	}
	router := setupRouter(h)

	body, _ := json.Marshal(creds)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response api.Error
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Bad request", response.Message)
	mockUserService.AssertExpectations(t)
}

func TestRegister_PasswordTooLong(t *testing.T) {
	mockUserService := new(MockUserService)
	email := openapi_types.Email("test@example.com")
	creds := api.PostRegisterJSONBody{
		Email:    email,
		Password: "thispasswordiswaytoolongandshouldberejectedbytheservice",
		Role:     api.Employee,
	}
	mockUserService.On("CreateUser", creds).Return(api.User{}, errs.ErrPasswordTooLong)

	h := &handler.Handler{
		Services: &service.Service{User: mockUserService},
		Logger:   slog.Default(),
	}
	router := setupRouter(h)

	body, _ := json.Marshal(creds)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response api.Error
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Bad request", response.Message)
	mockUserService.AssertExpectations(t)
}
