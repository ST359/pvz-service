package service

import (
	"errors"
	"testing"
	"time"

	"github.com/ST359/pvz-service/internal/api"
	errs "github.com/ST359/pvz-service/internal/app_errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockReceptionRepository struct {
	mock.Mock
}

func (m *MockReceptionRepository) Create(pvzID uuid.UUID) (api.Reception, error) {
	args := m.Called(pvzID)
	return args.Get(0).(api.Reception), args.Error(1)
}

func (m *MockReceptionRepository) AddProduct(receptionID uuid.UUID, product api.ProductType) (api.Product, error) {
	args := m.Called(receptionID, product)
	return args.Get(0).(api.Product), args.Error(1)
}

func (m *MockReceptionRepository) GetReceptionInProgress(pvzID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(pvzID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockReceptionRepository) DeleteLastProduct(receptionID uuid.UUID) error {
	args := m.Called(receptionID)
	return args.Error(0)
}

func (m *MockReceptionRepository) CloseLastReception(receptionID uuid.UUID) (api.Reception, error) {
	args := m.Called(receptionID)
	return args.Get(0).(api.Reception), args.Error(1)
}

func TestReceptionService_Create(t *testing.T) {
	pvzID := uuid.New()
	receptionID := uuid.New()

	tests := []struct {
		name        string
		pvzID       uuid.UUID
		mockSetup   func(*MockReceptionRepository)
		expected    api.Reception
		expectedErr string
	}{
		{
			name:  "successful creation",
			pvzID: pvzID,
			mockSetup: func(m *MockReceptionRepository) {
				m.On("GetReceptionInProgress", pvzID).Return(uuid.Nil, errs.ErrNoReceptionsInProgress)
				m.On("Create", pvzID).Return(api.Reception{
					Id:     &receptionID,
					PvzId:  pvzID,
					Status: api.InProgress,
				}, nil)
			},
			expected: api.Reception{
				Id:     &receptionID,
				PvzId:  pvzID,
				Status: api.InProgress,
			},
			expectedErr: "",
		},
		{
			name:  "existing reception in progress",
			pvzID: pvzID,
			mockSetup: func(m *MockReceptionRepository) {
				m.On("GetReceptionInProgress", pvzID).Return(receptionID, nil)
			},
			expected:    api.Reception{},
			expectedErr: errs.ErrReceptionNotClosed.Error(),
		},
		{
			name:  "repository error on check",
			pvzID: pvzID,
			mockSetup: func(m *MockReceptionRepository) {
				m.On("GetReceptionInProgress", pvzID).Return(uuid.Nil, errors.New("db error"))
			},
			expected:    api.Reception{},
			expectedErr: "service.reception.Create:service.reception.GetReceptionInProgress:db error",
		},
		{
			name:  "repository error on create",
			pvzID: pvzID,
			mockSetup: func(m *MockReceptionRepository) {
				m.On("GetReceptionInProgress", pvzID).Return(uuid.Nil, errs.ErrNoReceptionsInProgress)
				m.On("Create", pvzID).Return(api.Reception{}, errors.New("db error"))
			},
			expected:    api.Reception{},
			expectedErr: "service.reception.Create:db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockReceptionRepository)
			tt.mockSetup(mockRepo)

			service := NewReceptionService(mockRepo)
			result, err := service.Create(tt.pvzID)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestReceptionService_AddProduct(t *testing.T) {
	pvzID := uuid.New()
	receptionID := uuid.New()
	productType := api.ProductTypeElectronics

	tests := []struct {
		name        string
		pvzID       uuid.UUID
		product     api.ProductType
		mockSetup   func(*MockReceptionRepository)
		expected    api.Product
		expectedErr error
	}{
		{
			name:    "successful add product",
			pvzID:   pvzID,
			product: productType,
			mockSetup: func(m *MockReceptionRepository) {
				m.On("GetReceptionInProgress", pvzID).Return(receptionID, nil)
				m.On("AddProduct", receptionID, productType).Return(api.Product{
					Id:          &receptionID,
					ReceptionId: receptionID,
					Type:        productType,
				}, nil)
			},
			expected: api.Product{
				Id:          &receptionID,
				ReceptionId: receptionID,
				Type:        productType,
			},
			expectedErr: nil,
		},
		{
			name:    "no reception in progress",
			pvzID:   pvzID,
			product: productType,
			mockSetup: func(m *MockReceptionRepository) {
				m.On("GetReceptionInProgress", pvzID).Return(uuid.Nil, errs.ErrNoReceptionsInProgress)
			},
			expected:    api.Product{},
			expectedErr: errs.ErrNoReceptionsInProgress,
		},
		{
			name:    "repository error on add",
			pvzID:   pvzID,
			product: productType,
			mockSetup: func(m *MockReceptionRepository) {
				m.On("GetReceptionInProgress", pvzID).Return(receptionID, nil)
				m.On("AddProduct", receptionID, productType).Return(api.Product{}, errors.New("db error"))
			},
			expected:    api.Product{},
			expectedErr: errors.New("service.reception.AddProduct:db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockReceptionRepository)
			tt.mockSetup(mockRepo)

			service := NewReceptionService(mockRepo)
			result, err := service.AddProduct(tt.pvzID, tt.product)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestReceptionService_DeleteLastProduct(t *testing.T) {
	pvzID := uuid.New()
	receptionID := uuid.New()

	tests := []struct {
		name        string
		pvzID       uuid.UUID
		mockSetup   func(*MockReceptionRepository)
		expectedErr error
	}{
		{
			name:  "successful delete",
			pvzID: pvzID,
			mockSetup: func(m *MockReceptionRepository) {
				m.On("GetReceptionInProgress", pvzID).Return(receptionID, nil)
				m.On("DeleteLastProduct", receptionID).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:  "no reception in progress",
			pvzID: pvzID,
			mockSetup: func(m *MockReceptionRepository) {
				m.On("GetReceptionInProgress", pvzID).Return(uuid.Nil, errs.ErrNoReceptionsInProgress)
			},
			expectedErr: errs.ErrNoReceptionsInProgress,
		},
		{
			name:  "no products in reception",
			pvzID: pvzID,
			mockSetup: func(m *MockReceptionRepository) {
				m.On("GetReceptionInProgress", pvzID).Return(receptionID, nil)
				m.On("DeleteLastProduct", receptionID).Return(errs.ErrNoProductsInReception)
			},
			expectedErr: errs.ErrNoProductsInReception,
		},
		{
			name:  "repository error",
			pvzID: pvzID,
			mockSetup: func(m *MockReceptionRepository) {
				m.On("GetReceptionInProgress", pvzID).Return(receptionID, nil)
				m.On("DeleteLastProduct", receptionID).Return(errors.New("db error"))
			},
			expectedErr: errors.New("service.reception.DeleteLastProduct:db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockReceptionRepository)
			tt.mockSetup(mockRepo)

			service := NewReceptionService(mockRepo)
			err := service.DeleteLastProduct(tt.pvzID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestReceptionService_CloseLastReception(t *testing.T) {
	pvzID := uuid.New()
	receptionID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		pvzID       uuid.UUID
		mockSetup   func(*MockReceptionRepository)
		expected    api.Reception
		expectedErr error
	}{
		{
			name:  "successful close",
			pvzID: pvzID,
			mockSetup: func(m *MockReceptionRepository) {
				m.On("GetReceptionInProgress", pvzID).Return(receptionID, nil)
				m.On("CloseLastReception", receptionID).Return(api.Reception{
					Id:       &receptionID,
					PvzId:    pvzID,
					Status:   api.Close,
					DateTime: now,
				}, nil)
			},
			expected: api.Reception{
				Id:       &receptionID,
				PvzId:    pvzID,
				Status:   api.Close,
				DateTime: now,
			},
			expectedErr: nil,
		},
		{
			name:  "no reception in progress",
			pvzID: pvzID,
			mockSetup: func(m *MockReceptionRepository) {
				m.On("GetReceptionInProgress", pvzID).Return(uuid.Nil, errs.ErrNoReceptionsInProgress)
			},
			expected:    api.Reception{},
			expectedErr: errs.ErrNoReceptionsInProgress,
		},
		{
			name:  "repository error on close",
			pvzID: pvzID,
			mockSetup: func(m *MockReceptionRepository) {
				m.On("GetReceptionInProgress", pvzID).Return(receptionID, nil)
				m.On("CloseLastReception", receptionID).Return(api.Reception{}, errors.New("db error"))
			},
			expected:    api.Reception{},
			expectedErr: errors.New("service.reception.AddProduct:db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockReceptionRepository)
			tt.mockSetup(mockRepo)

			service := NewReceptionService(mockRepo)
			result, err := service.CloseLastReception(tt.pvzID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
