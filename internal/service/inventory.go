package service

import (
	"context"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/sirupsen/logrus"

	"github.com/senyabanana/shop-service/internal/entity"
	"github.com/senyabanana/shop-service/internal/repository"
)

type InventoryService struct {
	userRepo      repository.UserRepository
	inventoryRepo repository.InventoryRepository
	trManager     *manager.Manager
	log           *logrus.Logger
}

func NewInventoryService(
	userRepo repository.UserRepository,
	inventoryRepo repository.InventoryRepository,
	trManager *manager.Manager,
	log *logrus.Logger) *InventoryService {
	return &InventoryService{
		userRepo:      userRepo,
		inventoryRepo: inventoryRepo,
		trManager:     trManager,
		log:           log,
	}
}

func (s *InventoryService) BuyItem(ctx context.Context, userID int64, itemName string) error {
	s.log.Infof("User %d is attempting to buy item: %s", userID, itemName)

	return s.trManager.Do(ctx, func(ctx context.Context) error {
		item, err := s.inventoryRepo.GetItem(ctx, itemName)
		if err != nil {
			s.log.Warnf("BuyItem failed: item %s not found", itemName)
			return entity.ErrItemNotFound
		}

		balance, err := s.userRepo.GetUserBalance(ctx, userID)
		if err != nil {
			s.log.Errorf("BuyItem failed: failed to fetch balance for user %d: %v", userID, err)
			return err
		}
		if balance < item.Price {
			s.log.Warnf("BuyItem failed: insufficient balance for user %d", userID)
			return entity.ErrInsufficientBalance
		}

		err = s.userRepo.UpdateCoins(ctx, userID, -item.Price)
		if err != nil {
			s.log.Errorf("BuyItem failed: error updating balance for user %d: %v", userID, err)
			return err
		}

		_, err = s.inventoryRepo.GetInventoryItem(ctx, userID, item.ID)
		if err != nil {
			s.log.Warnf("BuyItem: item %s not found in inventory for user %d, creating new entry", itemName, userID)

			err = s.inventoryRepo.InsertInventoryItem(ctx, userID, item.ID)
			if err != nil {
				s.log.Errorf("BuyItem failed: error inserting inventory item %s for user %d: %v", itemName, userID, err)
				return err
			}
		} else {
			err = s.inventoryRepo.UpdateInventoryItem(ctx, userID, item.ID)
			if err != nil {
				s.log.Errorf("BuyItem failed: error updating inventory item %s for user %d: %v", itemName, userID, err)
				return err
			}
		}

		s.log.Infof("User %d successfully purchased item: %s", userID, itemName)
		return nil
	})
}
