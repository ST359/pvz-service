package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"log/slog"

	"github.com/ST359/pvz-service/internal/api"
	"github.com/ST359/pvz-service/internal/handler"
	"github.com/ST359/pvz-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPVZService is a mock implementation of PVZ service
type MockPVZService struct {
	mock.Mock
}

func (m *MockPVZService) Create(pvz api.PVZ) (api.PVZ, error) {
	args := m.Called(pvz)
	return args.Get(0).(api.PVZ), args.Error(1)
}

func (m *MockPVZService) GetByDate(params api.GetPvzParams) ([]api.PVZInfo, error) {
	args := m.Called(params)
	return args.Get(0).([]api.PVZInfo), args.Error(1)
}

func TestHandler_CreatePVZ(t *testing.T) {
	// Common test data
	testID := uuid.New()
	testTime := time.Now().UTC()
	testCity := api.Moscow

	tests := []struct {
		name           string
		role           interface{}
		requestBody    interface{}
		mockSetup      func(*MockPVZService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "successful creation",
			role: api.UserRoleModerator,
			requestBody: api.PVZ{
				City: testCity,
			},
			mockSetup: func(m *MockPVZService) {
				m.On("Create", mock.AnythingOfType("api.PVZ")).Return(api.PVZ{
					Id:               &testID,
					City:             testCity,
					RegistrationDate: &testTime,
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: api.PVZ{
				Id:               &testID,
				City:             testCity,
				RegistrationDate: &testTime,
			},
		},
		{
			name:           "access denied for non-moderator",
			role:           api.UserRoleEmployee,
			requestBody:    api.PVZ{},
			mockSetup:      func(m *MockPVZService) {},
			expectedStatus: http.StatusForbidden,
			expectedBody:   api.Error{Message: "Access denied"},
		},
		{
			name:           "bad request - invalid body",
			role:           api.UserRoleModerator,
			requestBody:    "invalid",
			mockSetup:      func(m *MockPVZService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   api.Error{Message: "Bad request"},
		},
		{
			name: "service error",
			role: api.UserRoleModerator,
			requestBody: api.PVZ{
				City: testCity,
			},
			mockSetup: func(m *MockPVZService) {
				m.On("Create", mock.AnythingOfType("api.PVZ")).Return(api.PVZ{}, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   api.Error{Message: "Internal server error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock PVZ service
			mockPVZ := new(MockPVZService)
			tt.mockSetup(mockPVZ)

			// Create handler with mock services
			h := &handler.Handler{
				Services: &service.Service{
					PVZ: mockPVZ,
				},
				Logger: slog.Default(),
			}

			// Setup Gin test context
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Set("userRole", tt.role)

			// Create request
			jsonBody, _ := json.Marshal(tt.requestBody)
			ctx.Request = httptest.NewRequest("POST", "/pvz", bytes.NewBuffer(jsonBody))
			ctx.Request.Header.Set("Content-Type", "application/json")

			// Call handler
			h.CreatePVZ(ctx)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Assert response body
			if tt.expectedBody != nil {
				var responseBody interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &responseBody); err != nil {
					t.Fatalf("Failed to unmarshal response body: %v", err)
				}

				expectedJSON, _ := json.Marshal(tt.expectedBody)
				var expectedBody interface{}
				json.Unmarshal(expectedJSON, &expectedBody)

				assert.Equal(t, expectedBody, responseBody)
			}

			mockPVZ.AssertExpectations(t)
		})
	}
}

func TestHandler_GetPVZ(t *testing.T) {
	// Common test data
	testTime := time.Now().UTC()
	testID := uuid.New()
	testCity := api.Kazan

	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*MockPVZService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:        "successful get with date range",
			queryParams: "startDate=" + testTime.Add(-24*time.Hour).Format(time.RFC3339) + "&endDate=" + testTime.Format(time.RFC3339),
			mockSetup: func(m *MockPVZService) {
				m.On("GetByDate", mock.AnythingOfType("api.GetPvzParams")).Return([]api.PVZInfo{
					{
						Pvz: &api.PVZ{
							Id:               &testID,
							City:             testCity,
							RegistrationDate: &testTime,
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []api.PVZInfo{
				{
					Pvz: &api.PVZ{
						Id:               &testID,
						City:             testCity,
						RegistrationDate: &testTime,
					},
				},
			},
		},
		{
			name:        "successful get with pagination",
			queryParams: "page=1&limit=10",
			mockSetup: func(m *MockPVZService) {
				m.On("GetByDate", mock.AnythingOfType("api.GetPvzParams")).Return([]api.PVZInfo{
					{
						Pvz: &api.PVZ{
							Id:               &testID,
							City:             testCity,
							RegistrationDate: &testTime,
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []api.PVZInfo{
				{
					Pvz: &api.PVZ{
						Id:               &testID,
						City:             testCity,
						RegistrationDate: &testTime,
					},
				},
			},
		},
		{
			name:           "bad request - invalid date format",
			queryParams:    "startDate=invalid",
			mockSetup:      func(m *MockPVZService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   api.Error{Message: "Bad request"},
		},
		{
			name:        "service error",
			queryParams: "startDate=" + testTime.Format(time.RFC3339),
			mockSetup: func(m *MockPVZService) {
				m.On("GetByDate", mock.AnythingOfType("api.GetPvzParams")).Return([]api.PVZInfo{}, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   api.Error{Message: "Internal server error"},
		},
		{
			name:        "empty result",
			queryParams: "startDate=" + testTime.Add(24*time.Hour).Format(time.RFC3339),
			mockSetup: func(m *MockPVZService) {
				m.On("GetByDate", mock.AnythingOfType("api.GetPvzParams")).Return([]api.PVZInfo{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []api.PVZInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock PVZ service
			mockPVZ := new(MockPVZService)
			tt.mockSetup(mockPVZ)

			// Create handler with mock services
			h := &handler.Handler{
				Services: &service.Service{
					PVZ: mockPVZ,
				},
				Logger: slog.Default(),
			}

			// Setup Gin test context
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			// Create request with query params
			ctx.Request = httptest.NewRequest("GET", "/pvz?"+tt.queryParams, nil)

			// Call handler
			h.GetPVZ(ctx)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Assert response body
			if tt.expectedBody != nil {
				var responseBody interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &responseBody); err != nil {
					t.Fatalf("Failed to unmarshal response body: %v", err)
				}

				expectedJSON, _ := json.Marshal(tt.expectedBody)
				var expectedBody interface{}
				json.Unmarshal(expectedJSON, &expectedBody)

				assert.Equal(t, expectedBody, responseBody)
			}

			mockPVZ.AssertExpectations(t)
		})
	}
}
