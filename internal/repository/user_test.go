package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/senyabanana/shop-service/internal/entity"
)

const testDriverName = "sqlmock"

func TestUserPostgres_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, testDriverName)
	repo := NewUserPostgres(sqlxDB)

	tests := []struct {
		name         string
		inputUser    entity.User
		mockBehavior func()
		wantID       int64
		wantError    error
	}{
		{
			name: "Success",
			inputUser: entity.User{
				Username: "testuser",
				Password: "testpass",
				Coins:    1000,
			},
			mockBehavior: func() {
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("testuser", "testpass", int64(1000)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
			},
			wantID:    1,
			wantError: nil,
		},
		{
			name: "Username already exists",
			inputUser: entity.User{
				Username: "existuser",
				Password: "testpass",
				Coins:    1000,
			},
			mockBehavior: func() {
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("existuser", "testpass", int64(1000)).
					WillReturnError(errors.New("pq: duplicate key value violates unique constraint"))
			},
			wantID:    0,
			wantError: errors.New("pq: duplicate key value violates unique constraint"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			ctx := context.Background()
			id, err := repo.CreateUser(ctx, tt.inputUser)

			assert.Equal(t, tt.wantID, id)
			assert.Equal(t, tt.wantError, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserPostgres_GetUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, testDriverName)
	repo := NewUserPostgres(sqlxDB)

	tests := []struct {
		name         string
		username     string
		mockBehavior func()
		wantUser     entity.User
		wantError    error
	}{
		{
			name:     "Success",
			username: "testuser",
			mockBehavior: func() {
				mock.ExpectQuery("SELECT id, username, users.password_hash, coins FROM users WHERE username").
					WithArgs("testuser").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "coins"}).
						AddRow(int64(1), "testuser", "testpass", int64(1000)))
			},
			wantUser: entity.User{
				ID:       1,
				Username: "testuser",
				Password: "testpass",
				Coins:    1000,
			},
			wantError: nil,
		},
		{
			name:     "User Not Found",
			username: "unknown_user",
			mockBehavior: func() {
				mock.ExpectQuery("SELECT id, username, users.password_hash, coins FROM users WHERE username").
					WithArgs("unknown_user").
					WillReturnError(errors.New("sql: no rows in result set"))
			},
			wantUser:  entity.User{},
			wantError: errors.New("sql: no rows in result set"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			ctx := context.Background()
			user, err := repo.GetUser(ctx, tt.username)

			assert.Equal(t, tt.wantUser, user)
			assert.Equal(t, tt.wantError, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserPostgres_GetUserBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, testDriverName)
	repo := NewUserPostgres(sqlxDB)

	tests := []struct {
		name         string
		userID       int64
		mockBehavior func()
		wantCoins    int64
		wantError    error
	}{
		{
			name:   "Success",
			userID: 1,
			mockBehavior: func() {
				mock.ExpectQuery("SELECT coins FROM users WHERE id").
					WithArgs(int64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"coins"}).AddRow(int64(500)))
			},
			wantCoins: 500,
			wantError: nil,
		},
		{
			name:   "Error in Query",
			userID: 2,
			mockBehavior: func() {
				mock.ExpectQuery("SELECT coins FROM users WHERE id").
					WithArgs(2).
					WillReturnError(errors.New("database error"))
			},
			wantCoins: 0,
			wantError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			ctx := context.Background()
			coins, err := repo.GetUserBalance(ctx, tt.userID)

			assert.Equal(t, tt.wantCoins, coins)
			assert.Equal(t, tt.wantError, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserPostgres_UpdateCoins(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, testDriverName)
	repo := NewUserPostgres(sqlxDB)

	tests := []struct {
		name         string
		userID       int64
		amount       int64
		mockBehavior func()
		wantError    error
	}{
		{
			name:   "Success - Increase Coins",
			userID: 1,
			amount: 50,
			mockBehavior: func() {
				mock.ExpectExec("UPDATE users SET coins = coins [+-]").
					WithArgs(int64(50), int64(1)).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantError: nil,
		},
		{
			name:   "Success - Decrease Coins",
			userID: 1,
			amount: -50,
			mockBehavior: func() {
				mock.ExpectExec("UPDATE users SET coins = coins [+-]").
					WithArgs(int64(-50), int64(1)).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantError: nil,
		},
		{
			name:   "Insufficient Balance",
			userID: 2,
			amount: 1000,
			mockBehavior: func() {
				mock.ExpectExec("UPDATE users SET coins = coins [+-]").
					WithArgs(int64(1000), int64(2)).
					WillReturnResult(sqlmock.NewResult(1, 0))
			},
			wantError: errors.New("insufficient balance"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			ctx := context.Background()
			err := repo.UpdateCoins(ctx, tt.userID, tt.amount)

			assert.Equal(t, tt.wantError, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
