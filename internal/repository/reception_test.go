package repository

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ST359/pvz-service/internal/api"
	errs "github.com/ST359/pvz-service/internal/app_errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReceptionPostgres_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewReceptionPostgres(db)
	pvzID := uuid.New()
	recID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		pvzID       uuid.UUID
		mockSetup   func()
		expected    api.Reception
		expectedErr error
	}{
		{
			name:  "successful creation",
			pvzID: pvzID,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "date", "pvz_id", "status"}).
					AddRow(recID, now, pvzID, "in_progress")
				mock.ExpectQuery("INSERT INTO receptions").
					WithArgs(pvzID).
					WillReturnRows(rows)
			},
			expected: api.Reception{
				Id:       &recID,
				DateTime: now,
				PvzId:    pvzID,
				Status:   "in_progress",
			},
			expectedErr: nil,
		},
		{
			name:  "database error",
			pvzID: pvzID,
			mockSetup: func() {
				mock.ExpectQuery("INSERT INTO receptions").
					WithArgs(pvzID).
					WillReturnError(sql.ErrConnDone)
			},
			expected:    api.Reception{},
			expectedErr: errors.New("repository.reception.Create: sql: connection is already closed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := repo.Create(tt.pvzID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReceptionPostgres_AddProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewReceptionPostgres(db)
	recID := uuid.New()
	prodID := uuid.New()
	now := time.Now()
	prodType := api.ProductTypeElectronics

	tests := []struct {
		name        string
		recID       uuid.UUID
		prodType    api.ProductType
		mockSetup   func()
		expected    api.Product
		expectedErr error
	}{
		{
			name:     "successful add product",
			recID:    recID,
			prodType: prodType,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "date", "reception_id", "type"}).
					AddRow(prodID, now, recID, prodType)
				mock.ExpectQuery("INSERT INTO products").
					WithArgs(recID, prodType).
					WillReturnRows(rows)
			},
			expected: api.Product{
				Id:          &prodID,
				DateTime:    &now,
				ReceptionId: recID,
				Type:        prodType,
			},
			expectedErr: nil,
		},
		{
			name:     "database error",
			recID:    recID,
			prodType: prodType,
			mockSetup: func() {
				mock.ExpectQuery("INSERT INTO products").
					WithArgs(recID, prodType).
					WillReturnError(sql.ErrConnDone)
			},
			expected:    api.Product{},
			expectedErr: errors.New("repository.reception.AddProduct: sql: connection is already closed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := repo.AddProduct(tt.recID, tt.prodType)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReceptionPostgres_GetReceptionInProgress(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewReceptionPostgres(db)
	pvzID := uuid.New()
	recID := uuid.New()

	tests := []struct {
		name        string
		pvzID       uuid.UUID
		mockSetup   func()
		expected    uuid.UUID
		expectedErr error
	}{
		{
			name:  "reception found",
			pvzID: pvzID,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(recID)
				mock.ExpectQuery("SELECT id FROM receptions").
					WithArgs(pvzID, "in_progress").
					WillReturnRows(rows)
			},
			expected:    recID,
			expectedErr: nil,
		},
		{
			name:  "no reception found",
			pvzID: pvzID,
			mockSetup: func() {
				mock.ExpectQuery("SELECT id FROM receptions").
					WithArgs(pvzID, "in_progress").
					WillReturnError(sql.ErrNoRows)
			},
			expected:    uuid.Nil,
			expectedErr: errs.ErrNoReceptionsInProgress,
		},
		{
			name:  "database error",
			pvzID: pvzID,
			mockSetup: func() {
				mock.ExpectQuery("SELECT id FROM receptions").
					WithArgs(pvzID, "in_progress").
					WillReturnError(sql.ErrConnDone)
			},
			expected:    uuid.Nil,
			expectedErr: errors.New("sql: connection is already closed"), // Compare against the root error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := repo.GetReceptionInProgress(tt.pvzID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				// Check that the error contains the expected message
				if tt.expectedErr == errs.ErrNoReceptionsInProgress {
					assert.ErrorIs(t, err, errs.ErrNoReceptionsInProgress)
				} else {
					assert.Contains(t, err.Error(), tt.expectedErr.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReceptionPostgres_DeleteLastProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewReceptionPostgres(db)
	recID := uuid.New()
	prodID := uuid.New()

	tests := []struct {
		name        string
		recID       uuid.UUID
		mockSetup   func()
		expectedErr error
	}{
		{
			name:  "successful delete",
			recID: recID,
			mockSetup: func() {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow(prodID)
				mock.ExpectQuery("SELECT id FROM products").
					WithArgs(recID).
					WillReturnRows(rows)
				mock.ExpectExec("DELETE FROM products").
					WithArgs(prodID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr: nil,
		},
		{
			name:  "no products found",
			recID: recID,
			mockSetup: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT id FROM products").
					WithArgs(recID).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectRollback()
			},
			expectedErr: errs.ErrNoProductsInReception,
		},
		{
			name:  "database error during select",
			recID: recID,
			mockSetup: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT id FROM products").
					WithArgs(recID).
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			expectedErr: errors.New("repository.pvz.DeleteLastProduct: sql: connection is already closed"),
		},
		{
			name:  "database error during delete",
			recID: recID,
			mockSetup: func() {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow(prodID)
				mock.ExpectQuery("SELECT id FROM products").
					WithArgs(recID).
					WillReturnRows(rows)
				mock.ExpectExec("DELETE FROM products").
					WithArgs(prodID).
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			expectedErr: errors.New("repository.pvz.DeleteLastProduct: sql: connection is already closed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := repo.DeleteLastProduct(tt.recID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReceptionPostgres_CloseLastReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewReceptionPostgres(db)
	recID := uuid.New()
	pvzID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		recID       uuid.UUID
		mockSetup   func()
		expected    api.Reception
		expectedErr error
	}{
		{
			name:  "successful close",
			recID: recID,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "date", "pvz_id", "status"}).
					AddRow(recID, now, pvzID, "close")
				mock.ExpectQuery("UPDATE receptions").
					WithArgs("close", recID).
					WillReturnRows(rows)
			},
			expected: api.Reception{
				Id:       &recID,
				DateTime: now,
				PvzId:    pvzID,
				Status:   "close",
			},
			expectedErr: nil,
		},
		{
			name:  "database error",
			recID: recID,
			mockSetup: func() {
				mock.ExpectQuery("UPDATE receptions").
					WithArgs("close", recID).
					WillReturnError(sql.ErrConnDone)
			},
			expected:    api.Reception{},
			expectedErr: errors.New("repository.pvz.CloseLastReception: sql: connection is already closed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := repo.CloseLastReception(tt.recID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
