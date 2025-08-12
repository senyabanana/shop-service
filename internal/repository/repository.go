package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/senyabanana/shop-service/internal/entity"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock.go

type UserRepository interface {
	CreateUser(ctx context.Context, user entity.User) (int64, error)
	GetUser(ctx context.Context, username string) (entity.User, error)
	GetUserBalance(ctx context.Context, userID int64) (int64, error)
	UpdateCoins(ctx context.Context, userID, amount int64) error
}

type TransactionRepository interface {
	GetReceivedTransactions(ctx context.Context, userID int64) ([]entity.TransactionDetail, error)
	GetSentTransactions(ctx context.Context, userID int64) ([]entity.TransactionDetail, error)
	InsertTransaction(ctx context.Context, fromUserID, toUserID, amount int64) error
}

type InventoryRepository interface {
	GetItem(ctx context.Context, itemName string) (entity.MerchItems, error)
	GetUserInventory(ctx context.Context, userID int64) ([]entity.InventoryItem, error)
	GetInventoryItem(ctx context.Context, userID, merchID int64) (int, error)
	UpdateInventoryItem(ctx context.Context, userID, merchID int64) error
	InsertInventoryItem(ctx context.Context, userID, merchID int64) error
}

type Repository struct {
	UserRepository
	TransactionRepository
	InventoryRepository
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		UserRepository:        NewUserPostgres(db),
		TransactionRepository: NewTransactionPostgres(db),
		InventoryRepository:   NewInventoryPostgres(db),
	}
}
