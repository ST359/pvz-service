package handler

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ST359/pvz-service/internal/api"
	errs "github.com/ST359/pvz-service/internal/app_errors"
	"github.com/ST359/pvz-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockReceptionService is a mock implementation of service.Reception
type MockReceptionService struct {
	mock.Mock
}

func (m *MockReceptionService) Create(pvzID uuid.UUID) (api.Reception, error) {
	args := m.Called(pvzID)
	return args.Get(0).(api.Reception), args.Error(1)
}

func (m *MockReceptionService) AddProduct(pvzID uuid.UUID, product api.ProductType) (api.Product, error) {
	args := m.Called(pvzID, product)
	return args.Get(0).(api.Product), args.Error(1)
}

func (m *MockReceptionService) GetReceptionInProgress(pvzID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(pvzID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockReceptionService) DeleteLastProduct(pvzID uuid.UUID) error {
	args := m.Called(pvzID)
	return args.Error(0)
}

func (m *MockReceptionService) CloseLastReception(pvzID uuid.UUID) (api.Reception, error) {
	args := m.Called(pvzID)
	return args.Get(0).(api.Reception), args.Error(1)
}

func setupReceptionRouter(h *Handler) *gin.Engine {
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set(userRole, api.UserRoleEmployee) // Default to employee role for tests
	})
	router.POST("/receptions", h.CreateReception)
	router.POST("/products", h.AddProduct)
	router.DELETE("/receptions/:pvzId/products/last", h.DeleteLastProduct)
	router.PUT("/receptions/:pvzId/close", h.CloseLastReception)
	return router
}

func TestCreateReception_Success(t *testing.T) {
	mockReception := new(MockReceptionService)
	pvzID := uuid.New()
	reqBody := api.PostReceptionsJSONBody{PvzId: pvzID}
	reception := api.Reception{
		Id:     &pvzID,
		PvzId:  pvzID,
		Status: api.InProgress,
	}

	mockReception.On("Create", pvzID).Return(reception, nil)

	h := &Handler{
		Services: &service.Service{Reception: mockReception},
		Logger:   slog.Default(),
	}
	router := setupReceptionRouter(h)

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/receptions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response api.Reception
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, reception, response)
	mockReception.AssertExpectations(t)
}

func TestCreateReception_NotEmployee(t *testing.T) {
	h := &Handler{
		Logger: slog.Default(),
	}
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set(userRole, api.UserRoleModerator) // Non-employee role
	})
	router.POST("/receptions", h.CreateReception)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/receptions", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateReception_ReceptionNotClosed(t *testing.T) {
	mockReception := new(MockReceptionService)
	pvzID := uuid.New()
	reqBody := api.PostReceptionsJSONBody{PvzId: pvzID}

	mockReception.On("Create", pvzID).Return(api.Reception{}, errs.ErrReceptionNotClosed)

	h := &Handler{
		Services: &service.Service{Reception: mockReception},
		Logger:   slog.Default(),
	}
	router := setupReceptionRouter(h)

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/receptions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockReception.AssertExpectations(t)
}

func TestCloseLastReception_Success(t *testing.T) {
	mockReception := new(MockReceptionService)
	pvzID := uuid.New()
	reception := api.Reception{
		Id:     &pvzID,
		PvzId:  pvzID,
		Status: api.Close,
	}

	mockReception.On("CloseLastReception", pvzID).Return(reception, nil)

	h := &Handler{
		Services: &service.Service{Reception: mockReception},
		Logger:   slog.Default(),
	}
	router := setupReceptionRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/receptions/"+pvzID.String()+"/close", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response api.Reception
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, reception, response)
	mockReception.AssertExpectations(t)
}

func TestCloseLastReception_NoReceptions(t *testing.T) {
	mockReception := new(MockReceptionService)
	pvzID := uuid.New()

	mockReception.On("CloseLastReception", pvzID).Return(api.Reception{}, errs.ErrNoReceptionsInProgress)

	h := &Handler{
		Services: &service.Service{Reception: mockReception},
		Logger:   slog.Default(),
	}
	router := setupReceptionRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/receptions/"+pvzID.String()+"/close", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockReception.AssertExpectations(t)
}

func TestDeleteLastProduct_Success(t *testing.T) {
	mockReception := new(MockReceptionService)
	pvzID := uuid.New()

	mockReception.On("DeleteLastProduct", pvzID).Return(nil)

	h := &Handler{
		Services: &service.Service{Reception: mockReception},
		Logger:   slog.Default(),
	}
	router := setupReceptionRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/receptions/"+pvzID.String()+"/products/last", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockReception.AssertExpectations(t)
}

func TestDeleteLastProduct_NoProducts(t *testing.T) {
	mockReception := new(MockReceptionService)
	pvzID := uuid.New()

	mockReception.On("DeleteLastProduct", pvzID).Return(errs.ErrNoProductsInReception)

	h := &Handler{
		Services: &service.Service{Reception: mockReception},
		Logger:   slog.Default(),
	}
	router := setupReceptionRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/receptions/"+pvzID.String()+"/products/last", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockReception.AssertExpectations(t)
}

func TestAddProduct_Success(t *testing.T) {
	mockReception := new(MockReceptionService)
	pvzID := uuid.New()
	reqBody := api.PostProductsJSONBody{
		PvzId: pvzID,
		Type:  api.PostProductsJSONBodyTypeShoes,
	}
	product := api.Product{
		Id:          &pvzID,
		ReceptionId: pvzID,
		Type:        api.ProductTypeShoes,
	}

	mockReception.On("AddProduct", pvzID, api.ProductTypeShoes).Return(product, nil)

	h := &Handler{
		Services: &service.Service{Reception: mockReception},
		Logger:   slog.Default(),
	}
	router := setupReceptionRouter(h)

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response api.Product
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, product, response)
	mockReception.AssertExpectations(t)
}

func TestAddProduct_NoReception(t *testing.T) {
	mockReception := new(MockReceptionService)
	pvzID := uuid.New()
	reqBody := api.PostProductsJSONBody{
		PvzId: pvzID,
		Type:  api.PostProductsJSONBodyTypeShoes,
	}

	mockReception.On("AddProduct", pvzID, api.ProductTypeShoes).Return(api.Product{}, errs.ErrNoReceptionsInProgress)

	h := &Handler{
		Services: &service.Service{Reception: mockReception},
		Logger:   slog.Default(),
	}
	router := setupReceptionRouter(h)

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockReception.AssertExpectations(t)
}

func TestAddProduct_NotEmployee(t *testing.T) {
	h := &Handler{
		Logger: slog.Default(),
	}
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set(userRole, api.UserRoleModerator) // Non-employee role
	})
	router.POST("/products", h.AddProduct)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/products", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
