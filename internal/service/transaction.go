package service

import (
	"context"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/sirupsen/logrus"

	"github.com/senyabanana/shop-service/internal/entity"
	"github.com/senyabanana/shop-service/internal/repository"
)

type TransactionService struct {
	userRepo        repository.UserRepository
	transactionRepo repository.TransactionRepository
	inventoryRepo   repository.InventoryRepository
	trManager       *manager.Manager
	log             *logrus.Logger
}

func NewTransactionService(
	userRepo repository.UserRepository,
	transactionRepo repository.TransactionRepository,
	inventoryRepo repository.InventoryRepository,
	trManager *manager.Manager,
	log *logrus.Logger) *TransactionService {
	return &TransactionService{
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		inventoryRepo:   inventoryRepo,
		trManager:       trManager,
		log:             log,
	}
}

func (s *TransactionService) GetUserInfo(ctx context.Context, userID int64) (entity.InfoResponse, error) {
	s.log.Infof("Fetching user info for userID: %d", userID)

	var info entity.InfoResponse

	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		var err error

		info.Coins, err = s.userRepo.GetUserBalance(ctx, userID)
		if err != nil {
			s.log.Errorf("Failed to get user balance for userID %d: %v", userID, err)
			return err
		}

		info.Inventory, err = s.inventoryRepo.GetUserInventory(ctx, userID)
		if err != nil {
			s.log.Errorf("Failed to get user inventory for userID %d: %v", userID, err)
			return err
		}

		info.CoinHistory.Received, err = s.transactionRepo.GetReceivedTransactions(ctx, userID)
		if err != nil {
			s.log.Errorf("Failed to get received transactions for userID %d: %v", userID, err)
			return err
		}

		info.CoinHistory.Sent, err = s.transactionRepo.GetSentTransactions(ctx, userID)
		if err != nil {
			s.log.Errorf("Failed to get sent transactions for userID %d: %v", userID, err)
			return err
		}

		return nil
	})

	if err != nil {
		return entity.InfoResponse{}, err
	}

	if info.Inventory == nil {
		info.Inventory = make([]entity.InventoryItem, 0)
	}
	if info.CoinHistory.Received == nil {
		info.CoinHistory.Received = make([]entity.TransactionDetail, 0)
	}
	if info.CoinHistory.Sent == nil {
		info.CoinHistory.Sent = make([]entity.TransactionDetail, 0)
	}

	s.log.Infof("Successfully fetched user info for userID: %d", userID)
	return info, nil
}

func (s *TransactionService) SendCoin(ctx context.Context, fromUserID int64, toUsername string, amount int64) error {
	s.log.Infof("User %d is sending %d coins to %s", fromUserID, amount, toUsername)

	return s.trManager.Do(ctx, func(ctx context.Context) error {
		toUser, err := s.userRepo.GetUser(ctx, toUsername)
		if err != nil {
			s.log.Warnf("SendCoin failed: recipient %s not found", toUsername)
			return entity.ErrRecipientNotFound
		}

		toUserID := toUser.ID
		if fromUserID == toUserID {
			s.log.Warnf("SendCoin failed: user %d tried to send coins to themselves", fromUserID)
			return entity.ErrSendThemselves
		}

		balance, err := s.userRepo.GetUserBalance(ctx, fromUserID)
		if err != nil {
			s.log.Errorf("SendCoin failed: failed to fetch balance for user %d: %v", fromUserID, err)
			return err
		}
		if balance < amount {
			s.log.Warnf("SendCoin failed: insufficient balance for user %d", fromUserID)
			return entity.ErrInsufficientBalance
		}

		err = s.userRepo.UpdateCoins(ctx, fromUserID, -amount)
		if err != nil {
			s.log.Errorf("SendCoin failed: failed to decrease balance for user %d: %v", fromUserID, err)
			return err
		}

		err = s.userRepo.UpdateCoins(ctx, toUserID, amount)
		if err != nil {
			s.log.Errorf("SendCoin failed: failed to increase balance for user %d: %v", toUserID, err)
			return err
		}

		err = s.transactionRepo.InsertTransaction(ctx, fromUserID, toUserID, amount)
		if err != nil {
			s.log.Errorf("SendCoin failed: failed to insert transaction record: %v", err)
			return err
		}

		s.log.Infof("Transaction successful: %d coins from %d to %s", amount, fromUserID, toUsername)
		return nil
	})
}
