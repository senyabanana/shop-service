package service

import (
	"context"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/sirupsen/logrus"

	"github.com/senyabanana/shop-service/internal/entity"
	"github.com/senyabanana/shop-service/internal/repository"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

type Authorization interface {
	GetUser(ctx context.Context, username string) (entity.User, error)
	CreateUser(ctx context.Context, username, password string) error
	GenerateToken(ctx context.Context, username, password string) (string, error)
	ParseToken(accessToken string) (int64, error)
}

type Transaction interface {
	GetUserInfo(ctx context.Context, userID int64) (entity.InfoResponse, error)
	SendCoin(ctx context.Context, fromUserID int64, toUsername string, amount int64) error
}

type Inventory interface {
	BuyItem(ctx context.Context, userID int64, itemName string) error
}

type Service struct {
	Authorization
	Transaction
	Inventory
}

func NewService(repos *repository.Repository, trManager *manager.Manager, jwtSecretKey string, log *logrus.Logger) *Service {
	return &Service{
		Authorization: NewAuthService(repos.UserRepository, trManager, jwtSecretKey, log),
		Transaction:   NewTransactionService(repos.UserRepository, repos.TransactionRepository, repos.InventoryRepository, trManager, log),
		Inventory:     NewInventoryService(repos.UserRepository, repos.InventoryRepository, trManager, log),
	}
}
