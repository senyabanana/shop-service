package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/dgrijalva/jwt-go"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/senyabanana/shop-service/internal/entity"
	mocks "github.com/senyabanana/shop-service/internal/repository/mocks"
)

const (
	testDriverName = "sqlmock"
	testUsername   = "testuser"
	testPassword   = "testpassword"
	testUserID     = int64(1)
	testJWTSecret  = "supersecret"
)

func TestAuthService_GetUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockLog := logrus.New()
	authService := NewAuthService(mockRepo, nil, testJWTSecret, mockLog)

	tests := []struct {
		name     string
		username string
		mockResp entity.User
		mockErr  error
		wantUser entity.User
		wantErr  error
	}{
		{
			name:     "Success",
			username: testUsername,
			mockResp: entity.User{ID: testUserID, Username: testUsername},
			mockErr:  nil,
			wantUser: entity.User{ID: testUserID, Username: testUsername},
			wantErr:  nil,
		},
		{
			name:     "User Not Found",
			username: "unknown",
			mockResp: entity.User{},
			mockErr:  entity.ErrUserNotFound,
			wantUser: entity.User{},
			wantErr:  entity.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.EXPECT().
				GetUser(gomock.Any(), tt.username).
				Return(tt.mockResp, tt.mockErr)

			user, err := authService.GetUser(context.Background(), tt.username)

			assert.Equal(t, tt.wantUser, user)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestAuthService_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	db, mock, _ := sqlmock.New()
	mockDB := sqlx.NewDb(db, testDriverName)
	mockTrManager := manager.Must(trmsqlx.NewDefaultFactory(mockDB))
	mockLog := logrus.New()
	authService := NewAuthService(mockRepo, mockTrManager, testJWTSecret, mockLog)

	tests := []struct {
		name       string
		username   string
		password   string
		mockErr    error
		wantErr    error
		wantCommit bool
	}{
		{
			name:       "Success",
			username:   testUsername,
			password:   testPassword,
			mockErr:    nil,
			wantErr:    nil,
			wantCommit: true,
		},
		{
			name:       "Username Already Exists",
			username:   testUsername,
			password:   testPassword,
			mockErr:    errors.New("user already exists"),
			wantErr:    errors.New("user already exists"),
			wantCommit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.EXPECT().
				CreateUser(gomock.Any(), gomock.Any()).
				Return(testUserID, tt.mockErr)

			err := authService.CreateUser(context.Background(), tt.username, tt.password)

			assert.Equal(t, tt.wantErr, err)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestAuthService_GenerateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockLog := logrus.New()
	authService := NewAuthService(mockRepo, nil, testJWTSecret, mockLog)

	tests := []struct {
		name      string
		username  string
		password  string
		mockUser  entity.User
		mockErr   error
		wantErr   error
		wantToken bool
	}{
		{
			name:     "Success",
			username: "validUser",
			password: "validPass",
			mockUser: entity.User{
				ID:       1,
				Username: "validUser",
				Password: generatePasswordHash("validPass"),
			},
			mockErr:   nil,
			wantErr:   nil,
			wantToken: true,
		},
		{
			name:      "User Not Found",
			username:  "unknownUser",
			password:  "anyPass",
			mockErr:   entity.ErrUserNotFound,
			wantErr:   entity.ErrUserNotFound,
			wantToken: false,
		},
		{
			name:     "Incorrect Password",
			username: "validUser",
			password: "wrongPass",
			mockUser: entity.User{
				ID:       1,
				Username: "validUser",
				Password: generatePasswordHash("validPass"),
			},
			mockErr:   nil,
			wantErr:   entity.ErrIncorrectPassword,
			wantToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.EXPECT().
				GetUser(gomock.Any(), tt.username).
				Return(tt.mockUser, tt.mockErr)

			token, err := authService.GenerateToken(context.Background(), tt.username, tt.password)

			assert.Equal(t, tt.wantErr, err)

			if tt.wantToken {
				assert.NotEmpty(t, token)
			} else {
				assert.Empty(t, token)
			}
		})
	}
}

func TestAuthService_ParseToken(t *testing.T) {
	mockLog := logrus.New()
	authService := NewAuthService(nil, nil, testJWTSecret, mockLog)

	validToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		1,
	})
	validTokenString, _ := validToken.SignedString([]byte(testJWTSecret))

	tests := []struct {
		name       string
		token      string
		wantUserID int64
		wantErr    error
	}{
		{
			name:       "Valid Token",
			token:      validTokenString,
			wantUserID: 1,
			wantErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := authService.ParseToken(tt.token)

			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				assert.Equal(t, tt.wantUserID, userID)
			} else {
				assert.Zero(t, userID)
			}
		})
	}
}
