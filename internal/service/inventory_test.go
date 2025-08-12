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

func TestInventoryService_BuyItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockInventoryRepo := mocks.NewMockInventoryRepository(ctrl)
	db, mock, _ := sqlmock.New()
	mockDB := sqlx.NewDb(db, testDriverName)
	mockTrManager := manager.Must(trmsqlx.NewDefaultFactory(mockDB))
	mockLog := logrus.New()

	service := NewInventoryService(mockUserRepo, mockInventoryRepo, mockTrManager, mockLog)

	tests := []struct {
		name         string
		userID       int64
		itemName     string
		mockBehavior func()
		wantErr      error
	}{
		{
			name:     "Success",
			userID:   1,
			itemName: "cup",
			mockBehavior: func() {
				mock.ExpectBegin()
				mockInventoryRepo.EXPECT().GetItem(gomock.Any(), "cup").Return(entity.MerchItems{ID: 10, ItemType: "cup", Price: 50}, nil)
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), int64(1)).Return(int64(100), nil)
				mockUserRepo.EXPECT().UpdateCoins(gomock.Any(), int64(1), int64(-50)).Return(nil)
				mockInventoryRepo.EXPECT().GetInventoryItem(gomock.Any(), int64(1), int64(10)).Return(1, nil)
				mockInventoryRepo.EXPECT().UpdateInventoryItem(gomock.Any(), int64(1), int64(10)).Return(nil)
				mock.ExpectCommit()
			},
			wantErr: nil,
		},
		{
			name:     "Item not found",
			userID:   1,
			itemName: "UnknownItem",
			mockBehavior: func() {
				mock.ExpectBegin()
				mockInventoryRepo.EXPECT().GetItem(gomock.Any(), "UnknownItem").Return(entity.MerchItems{}, entity.ErrItemNotFound)
				mock.ExpectRollback()
			},
			wantErr: entity.ErrItemNotFound,
		},
		{
			name:     "Error fetching balance",
			userID:   1,
			itemName: "cup",
			mockBehavior: func() {
				mock.ExpectBegin()
				mockInventoryRepo.EXPECT().GetItem(gomock.Any(), "cup").Return(entity.MerchItems{ID: 10, ItemType: "cup", Price: 50}, nil)
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), int64(1)).Return(int64(0), errors.New("db error"))
				mock.ExpectRollback()
			},
			wantErr: errors.New("db error"),
		},
		{
			name:     "Insufficient balance",
			userID:   1,
			itemName: "cup",
			mockBehavior: func() {
				mock.ExpectBegin()
				mockInventoryRepo.EXPECT().GetItem(gomock.Any(), "cup").Return(entity.MerchItems{ID: 10, ItemType: "cup", Price: 100}, nil)
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), int64(1)).Return(int64(50), nil)
				mock.ExpectRollback()
			},
			wantErr: entity.ErrInsufficientBalance,
		},
		{
			name:     "Error updating balance",
			userID:   1,
			itemName: "cup",
			mockBehavior: func() {
				mock.ExpectBegin()
				mockInventoryRepo.EXPECT().GetItem(gomock.Any(), "cup").Return(entity.MerchItems{ID: 10, ItemType: "cup", Price: 50}, nil)
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), int64(1)).Return(int64(100), nil)
				mockUserRepo.EXPECT().UpdateCoins(gomock.Any(), int64(1), int64(-50)).Return(errors.New("db error"))
				mock.ExpectRollback()
			},
			wantErr: errors.New("db error"),
		},
		{
			name:     "Error updating inventory item",
			userID:   1,
			itemName: "cup",
			mockBehavior: func() {
				mock.ExpectBegin()
				mockInventoryRepo.EXPECT().GetItem(gomock.Any(), "cup").Return(entity.MerchItems{ID: 10, ItemType: "cup", Price: 50}, nil)
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), int64(1)).Return(int64(100), nil)
				mockUserRepo.EXPECT().UpdateCoins(gomock.Any(), int64(1), int64(-50)).Return(nil)
				mockInventoryRepo.EXPECT().GetInventoryItem(gomock.Any(), int64(1), int64(10)).Return(1, nil)
				mockInventoryRepo.EXPECT().UpdateInventoryItem(gomock.Any(), int64(1), int64(10)).Return(errors.New("db error"))
				mock.ExpectRollback()
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()
			err := service.BuyItem(context.Background(), tt.userID, tt.itemName)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
