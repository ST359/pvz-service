package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ST359/pvz-service/internal/api"
	errs "github.com/ST359/pvz-service/internal/app_errors"
	"github.com/ST359/pvz-service/internal/service"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository is a mock implementation of repository.User
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) EmailExists(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) Create(email, passwordHash, role string) (uuid.UUID, error) {
	args := m.Called(email, passwordHash, role)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockUserRepository) Login(email string) (string, string, error) {
	args := m.Called(email)
	return args.String(0), args.String(1), args.Error(2)
}

func TestUserService_CreateUser(t *testing.T) {
	// Helper function to create test UUID
	newUUID := func() *openapi_types.UUID {
		id := openapi_types.UUID(uuid.New())
		return &id
	}

	tests := []struct {
		name        string
		input       api.PostRegisterJSONBody
		mockSetup   func(*MockUserRepository)
		expected    api.User
		expectedErr error
	}{
		{
			name: "successful creation - moderator",
			input: api.PostRegisterJSONBody{
				Email:    openapi_types.Email("moderator@example.com"),
				Password: "securePassword123",
				Role:     api.PostRegisterJSONBodyRole(api.UserRoleModerator),
			},
			mockSetup: func(m *MockUserRepository) {
				testUUID := uuid.New()
				m.On("EmailExists", "moderator@example.com").Return(false, nil)
				m.On("Create", "moderator@example.com", mock.Anything, "moderator").Return(testUUID, nil)
			},
			expected: api.User{
				Email: openapi_types.Email("moderator@example.com"),
				Id:    newUUID(),
				Role:  api.UserRoleModerator,
			},
			expectedErr: nil,
		},
		{
			name: "successful creation - employee",
			input: api.PostRegisterJSONBody{
				Email:    openapi_types.Email("employee@example.com"),
				Password: "employeePass123",
				Role:     api.PostRegisterJSONBodyRole(api.UserRoleEmployee),
			},
			mockSetup: func(m *MockUserRepository) {
				testUUID := uuid.New()
				m.On("EmailExists", "employee@example.com").Return(false, nil)
				m.On("Create", "employee@example.com", mock.Anything, "employee").Return(testUUID, nil)
			},
			expected: api.User{
				Email: openapi_types.Email("employee@example.com"),
				Id:    newUUID(),
				Role:  api.UserRoleEmployee,
			},
			expectedErr: nil,
		},
		{
			name: "email already exists",
			input: api.PostRegisterJSONBody{
				Email:    openapi_types.Email("existing@example.com"),
				Password: "password123",
				Role:     api.PostRegisterJSONBodyRole(api.UserRoleEmployee),
			},
			mockSetup: func(m *MockUserRepository) {
				m.On("EmailExists", "existing@example.com").Return(true, nil)
			},
			expected:    api.User{},
			expectedErr: errs.ErrEmailExists,
		},
		{
			name: "password too long",
			input: api.PostRegisterJSONBody{
				Email:    openapi_types.Email("test@example.com"),
				Password: string(make([]byte, 73)), // 73 bytes password
				Role:     api.PostRegisterJSONBodyRole(api.UserRoleEmployee),
			},
			mockSetup:   func(m *MockUserRepository) {},
			expected:    api.User{},
			expectedErr: errs.ErrPasswordTooLong,
		},
		{
			name: "repository error on email check",
			input: api.PostRegisterJSONBody{
				Email:    openapi_types.Email("test@example.com"),
				Password: "password123",
				Role:     api.PostRegisterJSONBodyRole(api.UserRoleModerator),
			},
			mockSetup: func(m *MockUserRepository) {
				m.On("EmailExists", "test@example.com").Return(false, errors.New("db connection error"))
			},
			expected:    api.User{},
			expectedErr: errors.New("service.user.CreateUser: db connection error"),
		},
		{
			name: "repository error on user creation",
			input: api.PostRegisterJSONBody{
				Email:    openapi_types.Email("test@example.com"),
				Password: "password123",
				Role:     api.PostRegisterJSONBodyRole(api.UserRoleEmployee),
			},
			mockSetup: func(m *MockUserRepository) {
				testID := uuid.New()
				m.On("EmailExists", "test@example.com").Return(false, nil)
				m.On("Create", "test@example.com", mock.Anything, "employee").Return(testID, errors.New("create failed"))
			},
			expected:    api.User{},
			expectedErr: errors.New("service.user.CreateUser: create failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			userService := service.NewUserService(mockRepo)
			result, err := userService.CreateUser(tt.input)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				assert.Equal(t, api.User{}, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.input.Email, result.Email)
				assert.Equal(t, api.UserRole(tt.input.Role), result.Role)
				assert.NotNil(t, result.Id)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_Login(t *testing.T) {
	tests := []struct {
		name        string
		input       api.PostLoginJSONBody
		mockSetup   func(*MockUserRepository)
		expectedErr error
		checkToken  func(t *testing.T, token string)
	}{
		{
			name: "successful login - moderator",
			input: api.PostLoginJSONBody{
				Email:    openapi_types.Email("moderator@example.com"),
				Password: "correct_password",
			},
			mockSetup: func(m *MockUserRepository) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.DefaultCost)
				m.On("Login", "moderator@example.com").Return(string(hashedPassword), "moderator", nil)
			},
			expectedErr: nil,
			checkToken: func(t *testing.T, token string) {
				userService := service.NewUserService(nil)
				role, err := userService.ParseToken(token)
				assert.NoError(t, err)
				assert.Equal(t, api.UserRoleModerator, role)
			},
		},
		{
			name: "successful login - employee",
			input: api.PostLoginJSONBody{
				Email:    openapi_types.Email("employee@example.com"),
				Password: "correct_password",
			},
			mockSetup: func(m *MockUserRepository) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.DefaultCost)
				m.On("Login", "employee@example.com").Return(string(hashedPassword), "employee", nil)
			},
			expectedErr: nil,
			checkToken: func(t *testing.T, token string) {
				userService := service.NewUserService(nil)
				role, err := userService.ParseToken(token)
				assert.NoError(t, err)
				assert.Equal(t, api.UserRoleEmployee, role)
			},
		},
		{
			name: "wrong credentials",
			input: api.PostLoginJSONBody{
				Email:    openapi_types.Email("user@example.com"),
				Password: "wrong_password",
			},
			mockSetup: func(m *MockUserRepository) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.DefaultCost)
				m.On("Login", "user@example.com").Return(string(hashedPassword), "employee", nil)
			},
			expectedErr: errs.ErrWrongCreds,
			checkToken:  nil,
		},
		{
			name: "user not found",
			input: api.PostLoginJSONBody{
				Email:    openapi_types.Email("nonexistent@example.com"),
				Password: "password123",
			},
			mockSetup: func(m *MockUserRepository) {
				m.On("Login", "nonexistent@example.com").Return("", "", errs.ErrWrongCreds)
			},
			expectedErr: errs.ErrWrongCreds,
			checkToken:  nil,
		},
		{
			name: "repository error",
			input: api.PostLoginJSONBody{
				Email:    openapi_types.Email("error@example.com"),
				Password: "password123",
			},
			mockSetup: func(m *MockUserRepository) {
				m.On("Login", "error@example.com").Return("", "", errors.New("database error"))
			},
			expectedErr: errors.New("database error"),
			checkToken:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			userService := service.NewUserService(mockRepo)
			token, err := userService.Login(tt.input)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				if tt.checkToken != nil {
					tt.checkToken(t, token)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_ParseToken(t *testing.T) {
	userService := service.NewUserService(nil)

	tests := []struct {
		name        string
		token       string
		expected    api.UserRole
		expectError bool // Just check if error occurs, not specific message
	}{
		{
			name:        "valid moderator token",
			token:       generateTestToken(t, "moderator"),
			expected:    api.UserRoleModerator,
			expectError: false,
		},
		{
			name:        "valid employee token",
			token:       generateTestToken(t, "employee"),
			expected:    api.UserRoleEmployee,
			expectError: false,
		},
		{
			name:        "invalid token format",
			token:       "invalid.token.here",
			expected:    "",
			expectError: true,
		},
		{
			name:        "empty token",
			token:       "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "expired token",
			token:       generateExpiredToken(t, "moderator"),
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, err := userService.ParseToken(tt.token)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, api.UserRole(""), role)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, role)
			}
		})
	}
}

func TestUserService_GenerateToken(t *testing.T) {
	userService := service.NewUserService(nil)

	tests := []struct {
		name       string
		role       string
		checkToken func(t *testing.T, token string)
	}{
		{
			name: "generate moderator token",
			role: "moderator",
			checkToken: func(t *testing.T, token string) {
				role, err := userService.ParseToken(token)
				assert.NoError(t, err)
				assert.Equal(t, api.UserRoleModerator, role)
			},
		},
		{
			name: "generate employee token",
			role: "employee",
			checkToken: func(t *testing.T, token string) {
				role, err := userService.ParseToken(token)
				assert.NoError(t, err)
				assert.Equal(t, api.UserRoleEmployee, role)
			},
		},
		{
			name: "empty role",
			role: "",
			checkToken: func(t *testing.T, token string) {
				role, err := userService.ParseToken(token)
				assert.NoError(t, err)
				assert.Equal(t, api.UserRole(""), role)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := userService.GenerateToken(tt.role)
			assert.NoError(t, err)
			assert.NotEmpty(t, token)
			tt.checkToken(t, token)
		})
	}
}

// Helper functions
func generateTestToken(t *testing.T, role string) string {
	t.Helper()
	userService := service.NewUserService(nil)
	token, err := userService.GenerateToken(role)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}
	return token
}

func generateExpiredToken(t *testing.T, role string) string {
	t.Helper()

	// Create claims with the actual struct from service package
	claims := struct {
		jwt.StandardClaims
		UserRole string `json:"role"`
	}{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
			IssuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
		},
		UserRole: role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte("secretKey"))
	if err != nil {
		t.Fatalf("Failed to generate expired token: %v", err)
	}
	return tokenString
}
