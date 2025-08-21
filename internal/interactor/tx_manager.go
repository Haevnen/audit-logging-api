package interactor

import (
	"context"

	"gorm.io/gorm"
)

type transactionKeyType string

const txKey transactionKeyType = "tx-context-key"

// TxManager provides transaction handling for use cases.
type TxManager struct {
	db *gorm.DB
}

func NewTxManager(db *gorm.DB) *TxManager {
	return &TxManager{db: db}
}

// TransactionExec executes fn inside a transaction.
// If fn returns an error → rollback, else commit.
func (t *TxManager) TransactionExec(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		return fn(txCtx) // rollback/commit managed by GORM
	})
}

// GetTx retrieves the *gorm.DB bound to the current transaction (if any).
// Falls back to base DB if no transaction is active.
func (t *TxManager) GetTx(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx
	}
	return t.db
}
