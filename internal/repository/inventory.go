package repository

import (
	"context"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/jmoiron/sqlx"

	"github.com/senyabanana/shop-service/internal/entity"
)

type InventoryPostgres struct {
	db     *sqlx.DB
	getter *trmsqlx.CtxGetter
}

func NewInventoryPostgres(db *sqlx.DB) *InventoryPostgres {
	return &InventoryPostgres{
		db:     db,
		getter: trmsqlx.DefaultCtxGetter,
	}
}

func (r *InventoryPostgres) GetItem(ctx context.Context, itemName string) (entity.MerchItems, error) {
	var item entity.MerchItems
	query := `SELECT id, item_type, price FROM merch_items WHERE item_type = $1`

	return item, r.getter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &item, query, itemName)
}

func (r *InventoryPostgres) GetUserInventory(ctx context.Context, userID int64) ([]entity.InventoryItem, error) {
	var inventory []entity.InventoryItem
	query := `
		SELECT mi.item_type AS type, SUM(i.quantity) AS quantity
		FROM inventory AS i
		JOIN merch_items AS mi ON i.merch_id = mi.id
		WHERE i.user_id = $1
		GROUP BY mi.item_type`

	return inventory, r.getter.DefaultTrOrDB(ctx, r.db).SelectContext(ctx, &inventory, query, userID)
}

func (r *InventoryPostgres) GetInventoryItem(ctx context.Context, userID, merchID int64) (int, error) {
	var quantity int
	query := `SELECT quantity FROM inventory WHERE user_id = $1 AND merch_id = $2`
	err := r.getter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &quantity, query, userID, merchID)
	if err != nil {
		return 0, err
	}

	return quantity, nil
}

func (r *InventoryPostgres) UpdateInventoryItem(ctx context.Context, userID, merchID int64) error {
	query := `UPDATE inventory SET quantity = quantity + 1 WHERE user_id = $1 AND merch_id = $2`
	_, err := r.getter.DefaultTrOrDB(ctx, r.db).ExecContext(ctx, query, userID, merchID)
	return err
}

func (r *InventoryPostgres) InsertInventoryItem(ctx context.Context, userID, merchID int64) error {
	query := `INSERT INTO inventory (user_id, merch_id, quantity) VALUES ($1, $2, 1)`
	_, err := r.getter.DefaultTrOrDB(ctx, r.db).ExecContext(ctx, query, userID, merchID)
	return err
}
