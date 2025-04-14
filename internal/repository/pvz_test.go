package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ST359/pvz-service/internal/api"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPVZPostgres_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPVZPostgres(db)

	tests := []struct {
		name        string
		input       api.PVZ
		mockSetup   func()
		expected    func(t *testing.T, result api.PVZ)
		expectedErr bool
	}{
		{
			name: "successful creation",
			input: api.PVZ{
				City: api.Moscow,
			},
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "registration_date", "city"}).
					AddRow(uuid.New(), time.Now(), api.Moscow)
				mock.ExpectQuery("INSERT INTO pvzs").
					WithArgs(api.Moscow).
					WillReturnRows(rows)
			},
			expected: func(t *testing.T, result api.PVZ) {
				assert.NotEqual(t, uuid.Nil, *result.Id)
				assert.Equal(t, api.Moscow, result.City)
				assert.NotNil(t, result.RegistrationDate)
			},
			expectedErr: false,
		},
		{
			name: "database error",
			input: api.PVZ{
				City: api.Moscow,
			},
			mockSetup: func() {
				mock.ExpectQuery("INSERT INTO pvzs").
					WithArgs(api.Moscow).
					WillReturnError(sql.ErrConnDone)
			},
			expected:    func(t *testing.T, result api.PVZ) {},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := repo.Create(tt.input)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Equal(t, api.PVZ{}, result)
			} else {
				assert.NoError(t, err)
				tt.expected(t, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPVZPostgres_GetByDate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPVZPostgres(db)

	now := time.Now()
	testUUID := uuid.New()

	tests := []struct {
		name        string
		params      api.GetPvzParams
		mockSetup   func()
		expectedLen int
		expectedErr bool
	}{
		{
			name: "successful get with date range",
			params: api.GetPvzParams{
				StartDate: &now,
				Limit:     ptrToInt(10),
			},
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"pvz_id", "city", "registration_date", "receptions"}).
					AddRow(testUUID, api.Moscow, now, []byte("[]"))
				mock.ExpectQuery("SELECT \\* FROM get_pvz_with_receptions_paginated").
					WithArgs(now, nil, 10, 0).
					WillReturnRows(rows)
			},
			expectedLen: 1,
			expectedErr: false,
		},
		{
			name: "empty result",
			params: api.GetPvzParams{
				Limit: ptrToInt(10),
			},
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"pvz_id", "city", "registration_date", "receptions"})
				mock.ExpectQuery("SELECT \\* FROM get_pvz_with_receptions_paginated").
					WithArgs(nil, nil, 10, 0).
					WillReturnRows(rows)
			},
			expectedLen: 0,
			expectedErr: false,
		},
		{
			name: "database error",
			params: api.GetPvzParams{
				Limit: ptrToInt(10),
			},
			mockSetup: func() {
				mock.ExpectQuery("SELECT \\* FROM get_pvz_with_receptions_paginated").
					WithArgs(nil, nil, 10, 0).
					WillReturnError(sql.ErrConnDone)
			},
			expectedLen: 0,
			expectedErr: true,
		},
		{
			name: "invalid JSON in receptions",
			params: api.GetPvzParams{
				Limit: ptrToInt(10),
			},
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"pvz_id", "city", "registration_date", "receptions"}).
					AddRow(testUUID, api.Moscow, now, []byte("{invalid}"))
				mock.ExpectQuery("SELECT \\* FROM get_pvz_with_receptions_paginated").
					WithArgs(nil, nil, 10, 0).
					WillReturnRows(rows)
			},
			expectedLen: 0,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := repo.GetByDate(tt.params)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLen, len(result))
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func ptrToInt(i int) *int {
	return &i
}
