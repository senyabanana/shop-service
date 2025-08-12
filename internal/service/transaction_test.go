package service

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/senyabanana/shop-service/internal/entity"
	mocks "github.com/senyabanana/shop-service/internal/repository/mocks"
)

func TestTransactionService_GetUserInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockTransactionRepo := mocks.NewMockTransactionRepository(ctrl)
	mockInventoryRepo := mocks.NewMockInventoryRepository(ctrl)
	db, mock, _ := sqlmock.New()
	mockDB := sqlx.NewDb(db, testDriverName)
	mockTrManager := manager.Must(trmsqlx.NewDefaultFactory(mockDB))
	mockLog := logrus.New()

	service := NewTransactionService(mockUserRepo, mockTransactionRepo, mockInventoryRepo, mockTrManager, mockLog)

	tests := []struct {
		name         string
		mockBehavior func()
		userID       int64
		wantInfo     entity.InfoResponse
		wantErr      error
	}{
		{
			name:   "Success",
			userID: 1,
			mockBehavior: func() {
				mock.ExpectBegin()
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), int64(1)).Return(int64(100), nil)
				mockInventoryRepo.EXPECT().GetUserInventory(gomock.Any(), int64(1)).Return([]entity.InventoryItem{}, nil)
				mockTransactionRepo.EXPECT().GetReceivedTransactions(gomock.Any(), int64(1)).Return([]entity.TransactionDetail{}, nil)
				mockTransactionRepo.EXPECT().GetSentTransactions(gomock.Any(), int64(1)).Return([]entity.TransactionDetail{}, nil)
				mock.ExpectCommit()
			},
			wantInfo: entity.InfoResponse{
				Coins:       100,
				Inventory:   []entity.InventoryItem{},
				CoinHistory: entity.CoinHistory{Received: []entity.TransactionDetail{}, Sent: []entity.TransactionDetail{}},
			},
			wantErr: nil,
		},
		{
			name:   "Error fetching balance",
			userID: 2,
			mockBehavior: func() {
				mock.ExpectBegin()
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), int64(2)).Return(int64(0), errors.New("db error"))
				mock.ExpectRollback()
			},
			wantInfo: entity.InfoResponse{},
			wantErr:  errors.New("db error"),
		},
		{
			name:   "Error fetching inventory",
			userID: 3,
			mockBehavior: func() {
				mock.ExpectBegin()
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), int64(3)).Return(int64(100), nil)
				mockInventoryRepo.EXPECT().GetUserInventory(gomock.Any(), int64(3)).Return(nil, errors.New("inventory error"))
				mock.ExpectRollback()
			},
			wantInfo: entity.InfoResponse{},
			wantErr:  errors.New("inventory error"),
		},
		{
			name:   "Error fetching received transactions",
			userID: 4,
			mockBehavior: func() {
				mock.ExpectBegin()
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), int64(4)).Return(int64(100), nil)
				mockInventoryRepo.EXPECT().GetUserInventory(gomock.Any(), int64(4)).Return([]entity.InventoryItem{}, nil)
				mockTransactionRepo.EXPECT().GetReceivedTransactions(gomock.Any(), int64(4)).Return(nil, errors.New("received transactions error"))
				mock.ExpectRollback()
			},
			wantInfo: entity.InfoResponse{},
			wantErr:  errors.New("received transactions error"),
		},
		{
			name:   "Error fetching sent transactions",
			userID: 5,
			mockBehavior: func() {
				mock.ExpectBegin()
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), int64(5)).Return(int64(100), nil)
				mockInventoryRepo.EXPECT().GetUserInventory(gomock.Any(), int64(5)).Return([]entity.InventoryItem{}, nil)
				mockTransactionRepo.EXPECT().GetReceivedTransactions(gomock.Any(), int64(5)).Return([]entity.TransactionDetail{}, nil)
				mockTransactionRepo.EXPECT().GetSentTransactions(gomock.Any(), int64(5)).Return(nil, errors.New("sent transactions error"))
				mock.ExpectRollback()
			},
			wantInfo: entity.InfoResponse{},
			wantErr:  errors.New("sent transactions error"),
		},
		{
			name:   "User has empty inventory and transactions",
			userID: 6,
			mockBehavior: func() {
				mock.ExpectBegin()
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), int64(6)).Return(int64(50), nil)
				mockInventoryRepo.EXPECT().GetUserInventory(gomock.Any(), int64(6)).Return(nil, nil)
				mockTransactionRepo.EXPECT().GetReceivedTransactions(gomock.Any(), int64(6)).Return(nil, nil)
				mockTransactionRepo.EXPECT().GetSentTransactions(gomock.Any(), int64(6)).Return(nil, nil)
				mock.ExpectCommit()
			},
			wantInfo: entity.InfoResponse{
				Coins:       50,
				Inventory:   []entity.InventoryItem{},
				CoinHistory: entity.CoinHistory{Received: []entity.TransactionDetail{}, Sent: []entity.TransactionDetail{}},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()
			info, err := service.GetUserInfo(context.Background(), tt.userID)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantInfo, info)
		})
	}
}

func TestTransactionService_SendCoin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockTransactionRepo := mocks.NewMockTransactionRepository(ctrl)

	db, mock, _ := sqlmock.New()
	mockDB := sqlx.NewDb(db, testDriverName)
	mockTrManager := manager.Must(trmsqlx.NewDefaultFactory(mockDB))

	mockLog := logrus.New()
	service := NewTransactionService(mockUserRepo, mockTransactionRepo, nil, mockTrManager, mockLog)

	tests := []struct {
		name         string
		mockBehavior func()
		fromUserID   int64
		toUsername   string
		amount       int64
		wantErr      error
	}{
		{
			name:       "Success",
			fromUserID: 1,
			toUsername: "recipient",
			amount:     50,
			mockBehavior: func() {
				mock.ExpectBegin()
				mockUserRepo.EXPECT().GetUser(gomock.Any(), "recipient").Return(entity.User{ID: 2, Username: "recipient"}, nil)
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), int64(1)).Return(int64(100), nil)
				mockUserRepo.EXPECT().UpdateCoins(gomock.Any(), int64(1), int64(-50)).Return(nil)
				mockUserRepo.EXPECT().UpdateCoins(gomock.Any(), int64(2), int64(50)).Return(nil)
				mockTransactionRepo.EXPECT().InsertTransaction(gomock.Any(), int64(1), int64(2), int64(50)).Return(nil)
				mock.ExpectCommit()
			},
			wantErr: nil,
		},
		{
			name:       "Recipient not found",
			fromUserID: 1,
			toUsername: "unknownUser",
			amount:     50,
			mockBehavior: func() {
				mock.ExpectBegin()
				mockUserRepo.EXPECT().GetUser(gomock.Any(), "unknownUser").Return(entity.User{}, entity.ErrRecipientNotFound)
				mock.ExpectRollback()
			},
			wantErr: entity.ErrRecipientNotFound,
		},
		{
			name:       "Sender tries to send to themselves",
			fromUserID: 1,
			toUsername: "sender",
			amount:     50,
			mockBehavior: func() {
				mock.ExpectBegin()
				mockUserRepo.EXPECT().GetUser(gomock.Any(), "sender").Return(entity.User{ID: 1, Username: "sender"}, nil)
				mock.ExpectRollback()
			},
			wantErr: entity.ErrSendThemselves,
		},
		{
			name:       "Insufficient balance",
			fromUserID: 1,
			toUsername: "recipient",
			amount:     200,
			mockBehavior: func() {
				mock.ExpectBegin()
				mockUserRepo.EXPECT().GetUser(gomock.Any(), "recipient").Return(entity.User{ID: 2, Username: "recipient"}, nil)
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), int64(1)).Return(int64(100), nil)
				mock.ExpectRollback()
			},
			wantErr: entity.ErrInsufficientBalance,
		},
		{
			name:       "Error fetching sender's balance",
			fromUserID: 1,
			toUsername: "recipient",
			amount:     50,
			mockBehavior: func() {
				mock.ExpectBegin()
				mockUserRepo.EXPECT().GetUser(gomock.Any(), "recipient").Return(entity.User{ID: 2, Username: "recipient"}, nil)
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), int64(1)).Return(int64(0), errors.New("db error"))
				mock.ExpectRollback()
			},
			wantErr: errors.New("db error"),
		},
		{
			name:       "Error decreasing sender's balance",
			fromUserID: 1,
			toUsername: "recipient",
			amount:     50,
			mockBehavior: func() {
				mock.ExpectBegin()
				mockUserRepo.EXPECT().GetUser(gomock.Any(), "recipient").Return(entity.User{ID: 2, Username: "recipient"}, nil)
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), int64(1)).Return(int64(100), nil)
				mockUserRepo.EXPECT().UpdateCoins(gomock.Any(), int64(1), int64(-50)).Return(errors.New("db error"))
				mock.ExpectRollback()
			},
			wantErr: errors.New("db error"),
		},
		{
			name:       "Error increasing recipient's balance",
			fromUserID: 1,
			toUsername: "recipient",
			amount:     50,
			mockBehavior: func() {
				mock.ExpectBegin()
				mockUserRepo.EXPECT().GetUser(gomock.Any(), "recipient").Return(entity.User{ID: 2, Username: "recipient"}, nil)
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), int64(1)).Return(int64(100), nil)
				mockUserRepo.EXPECT().UpdateCoins(gomock.Any(), int64(1), int64(-50)).Return(nil)
				mockUserRepo.EXPECT().UpdateCoins(gomock.Any(), int64(2), int64(50)).Return(errors.New("db error"))
				mock.ExpectRollback()
			},
			wantErr: errors.New("db error"),
		},
		{
			name:       "Error inserting transaction",
			fromUserID: 1,
			toUsername: "recipient",
			amount:     50,
			mockBehavior: func() {
				mock.ExpectBegin()
				mockUserRepo.EXPECT().GetUser(gomock.Any(), "recipient").Return(entity.User{ID: 2, Username: "recipient"}, nil)
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), int64(1)).Return(int64(100), nil)
				mockUserRepo.EXPECT().UpdateCoins(gomock.Any(), int64(1), int64(-50)).Return(nil)
				mockUserRepo.EXPECT().UpdateCoins(gomock.Any(), int64(2), int64(50)).Return(nil)
				mockTransactionRepo.EXPECT().InsertTransaction(gomock.Any(), int64(1), int64(2), int64(50)).Return(errors.New("db error"))
				mock.ExpectRollback()
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()
			err := service.SendCoin(context.Background(), tt.fromUserID, tt.toUsername, tt.amount)

			assert.Equal(t, tt.wantErr, err)
		})
	}
}
