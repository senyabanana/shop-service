package repository

import (
	"context"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/jmoiron/sqlx"

	"github.com/senyabanana/shop-service/internal/entity"
)

type TransactionPostgres struct {
	db     *sqlx.DB
	getter *trmsqlx.CtxGetter
}

func NewTransactionPostgres(db *sqlx.DB) *TransactionPostgres {
	return &TransactionPostgres{
		db:     db,
		getter: trmsqlx.DefaultCtxGetter,
	}
}

func (r *TransactionPostgres) GetReceivedTransactions(ctx context.Context, userID int64) ([]entity.TransactionDetail, error) {
	var received []entity.TransactionDetail
	query := `
		SELECT u.username AS from_user, t.amount
		FROM transactions AS t
		JOIN users AS u ON t.from_user = u.id
		WHERE t.to_user = $1`

	return received, r.getter.DefaultTrOrDB(ctx, r.db).SelectContext(ctx, &received, query, userID)
}

func (r *TransactionPostgres) GetSentTransactions(ctx context.Context, userID int64) ([]entity.TransactionDetail, error) {
	var sent []entity.TransactionDetail
	query := `
		SELECT u.username AS to_user, t.amount
		FROM transactions AS t
		JOIN users AS u ON t.to_user = u.id
		WHERE t.from_user = $1`

	return sent, r.getter.DefaultTrOrDB(ctx, r.db).SelectContext(ctx, &sent, query, userID)
}

func (r *TransactionPostgres) InsertTransaction(ctx context.Context, fromUserID, toUserID, amount int64) error {
	query := `INSERT INTO transactions (from_user, to_user, amount) VALUES ($1, $2, $3)`
	_, err := r.getter.DefaultTrOrDB(ctx, r.db).ExecContext(ctx, query, fromUserID, toUserID, amount)

	return err
}
