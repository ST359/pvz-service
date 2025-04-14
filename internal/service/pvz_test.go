package service

import (
	"errors"
	"testing"
	"time"

	"github.com/ST359/pvz-service/internal/api"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPVZRepository is a mock implementation of repository.PVZ
type MockPVZRepository struct {
	mock.Mock
}

func (m *MockPVZRepository) Create(pvz api.PVZ) (api.PVZ, error) {
	args := m.Called(pvz)
	return args.Get(0).(api.PVZ), args.Error(1)
}

func (m *MockPVZRepository) GetByDate(params api.GetPvzParams) ([]api.PVZInfo, error) {
	args := m.Called(params)
	return args.Get(0).([]api.PVZInfo), args.Error(1)
}

func TestPVZService_Create(t *testing.T) {
	now := time.Now()
	testUUID := uuid.New()

	tests := []struct {
		name        string
		input       api.PVZ
		mockSetup   func(*MockPVZRepository)
		expected    api.PVZ
		expectedErr error
	}{
		{
			name: "successful creation",
			input: api.PVZ{
				City:             api.Moscow,
				RegistrationDate: &now,
			},
			mockSetup: func(m *MockPVZRepository) {
				m.On("Create", mock.AnythingOfType("api.PVZ")).Return(api.PVZ{
					City:             api.Moscow,
					Id:               &testUUID,
					RegistrationDate: &now,
				}, nil)
			},
			expected: api.PVZ{
				City:             api.Moscow,
				Id:               &testUUID,
				RegistrationDate: &now,
			},
			expectedErr: nil,
		},
		{
			name: "repository error",
			input: api.PVZ{
				City: api.Kazan,
			},
			mockSetup: func(m *MockPVZRepository) {
				m.On("Create", mock.AnythingOfType("api.PVZ")).Return(api.PVZ{}, errors.New("db error"))
			},
			expected:    api.PVZ{},
			expectedErr: errors.New("service.pvz.Create:db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPVZRepository)
			tt.mockSetup(mockRepo)

			service := NewPVZService(mockRepo)
			result, err := service.Create(tt.input)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.City, result.City)
				assert.NotNil(t, result.Id)
				assert.Equal(t, *tt.expected.Id, *result.Id)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPVZService_GetByDate(t *testing.T) {
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()
	testUUID := uuid.New()

	tests := []struct {
		name        string
		input       api.GetPvzParams
		mockSetup   func(*MockPVZRepository)
		expected    []api.PVZInfo
		expectedErr error
	}{
		{
			name: "successful get by date",
			input: api.GetPvzParams{
				StartDate: &startDate,
				EndDate:   &endDate,
				Page:      ptrToInt(1),
				Limit:     ptrToInt(10),
			},
			mockSetup: func(m *MockPVZRepository) {
				m.On("GetByDate", mock.AnythingOfType("api.GetPvzParams")).Return([]api.PVZInfo{
					{
						Pvz: &api.PVZ{
							City:             api.Moscow,
							Id:               &testUUID,
							RegistrationDate: &endDate,
						},
					},
				}, nil)
			},
			expected: []api.PVZInfo{
				{
					Pvz: &api.PVZ{
						City:             api.Moscow,
						Id:               &testUUID,
						RegistrationDate: &endDate,
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "empty result",
			input: api.GetPvzParams{
				StartDate: &startDate,
				EndDate:   &endDate,
			},
			mockSetup: func(m *MockPVZRepository) {
				m.On("GetByDate", mock.AnythingOfType("api.GetPvzParams")).Return([]api.PVZInfo{}, nil)
			},
			expected:    []api.PVZInfo{},
			expectedErr: nil,
		},
		{
			name: "repository error",
			input: api.GetPvzParams{
				StartDate: &startDate,
			},
			mockSetup: func(m *MockPVZRepository) {
				m.On("GetByDate", mock.AnythingOfType("api.GetPvzParams")).Return([]api.PVZInfo{}, errors.New("db error"))
			},
			expected:    []api.PVZInfo{},
			expectedErr: errors.New("service.pvz.GetByDate:db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPVZRepository)
			tt.mockSetup(mockRepo)

			service := NewPVZService(mockRepo)
			result, err := service.GetByDate(tt.input)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expected), len(result))
				if len(tt.expected) > 0 {
					assert.Equal(t, tt.expected[0].Pvz.City, result[0].Pvz.City)
					assert.Equal(t, *tt.expected[0].Pvz.Id, *result[0].Pvz.Id)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Helper function
func ptrToInt(i int) *int {
	return &i
}
