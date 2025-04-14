package repository

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ST359/pvz-service/internal/app_errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserPostgres_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserPostgres(db)
	testUUID := uuid.New()
	email := "test@example.com"
	passwordHash := "hashedpassword"
	role := "admin"

	tests := []struct {
		name        string
		email       string
		password    string
		role        string
		mockSetup   func()
		expected    uuid.UUID
		expectedErr error
	}{
		{
			name:     "successful creation",
			email:    email,
			password: passwordHash,
			role:     role,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(testUUID)
				mock.ExpectQuery("INSERT INTO users").
					WithArgs(email, passwordHash, role).
					WillReturnRows(rows)
			},
			expected:    testUUID,
			expectedErr: nil,
		},
		{
			name:     "database error",
			email:    email,
			password: passwordHash,
			role:     role,
			mockSetup: func() {
				mock.ExpectQuery("INSERT INTO users").
					WithArgs(email, passwordHash, role).
					WillReturnError(sql.ErrConnDone)
			},
			expected:    uuid.Nil,
			expectedErr: errors.New("sql: connection is already closed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := repo.Create(tt.email, tt.password, tt.role)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserPostgres_Login(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserPostgres(db)
	email := "test@example.com"
	passwordHash := "hashedpassword"
	role := "admin"

	tests := []struct {
		name      string
		email     string
		mockSetup func()
		expected  struct {
			hash string
			role string
		}
		expectedErr error
	}{
		{
			name:  "successful login",
			email: email,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"password_hash", "role"}).
					AddRow(passwordHash, role)
				mock.ExpectQuery("SELECT password_hash, role FROM users").
					WithArgs(email).
					WillReturnRows(rows)
			},
			expected: struct {
				hash string
				role string
			}{
				hash: passwordHash,
				role: role,
			},
			expectedErr: nil,
		},
		{
			name:  "user not found",
			email: email,
			mockSetup: func() {
				mock.ExpectQuery("SELECT password_hash, role FROM users").
					WithArgs(email).
					WillReturnError(sql.ErrNoRows)
			},
			expected: struct {
				hash string
				role string
			}{},
			expectedErr: app_errors.ErrWrongCreds,
		},
		{
			name:  "database error",
			email: email,
			mockSetup: func() {
				mock.ExpectQuery("SELECT password_hash, role FROM users").
					WithArgs(email).
					WillReturnError(sql.ErrConnDone)
			},
			expected: struct {
				hash string
				role string
			}{},
			expectedErr: errors.New("sql: connection is already closed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			hash, role, err := repo.Login(tt.email)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				if tt.expectedErr == app_errors.ErrWrongCreds {
					assert.ErrorIs(t, err, app_errors.ErrWrongCreds)
				} else {
					assert.Contains(t, err.Error(), tt.expectedErr.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.hash, hash)
				assert.Equal(t, tt.expected.role, role)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserPostgres_EmailExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserPostgres(db)
	email := "test@example.com"

	tests := []struct {
		name        string
		email       string
		mockSetup   func()
		expected    bool
		expectedErr error
	}{
		{
			name:  "email exists",
			email: email,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\)>0 FROM users").
					WithArgs(email).
					WillReturnRows(rows)
			},
			expected:    true,
			expectedErr: nil,
		},
		{
			name:  "email doesn't exist",
			email: email,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\)>0 FROM users").
					WithArgs(email).
					WillReturnRows(rows)
			},
			expected:    false,
			expectedErr: nil,
		},
		{
			name:  "database error",
			email: email,
			mockSetup: func() {
				mock.ExpectQuery("SELECT COUNT\\(\\*\\)>0 FROM users").
					WithArgs(email).
					WillReturnError(sql.ErrConnDone)
			},
			expected:    false,
			expectedErr: errors.New("sql: connection is already closed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := repo.EmailExists(tt.email)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
