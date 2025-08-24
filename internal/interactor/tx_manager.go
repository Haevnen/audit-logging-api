package interactor

//go:generate mockgen -source=tx_manager.go -destination=./mocks/mock_tx_manager.go -package=mocks

import (
	"context"

	"gorm.io/gorm"
)

type transactionKeyType string

const txKey transactionKeyType = "tx-context-key"

type TxManager interface {
	TransactionExec(ctx context.Context, fn func(ctx context.Context) error) error
	GetTx(ctx context.Context) *gorm.DB
}

// TxManager provides transaction handling for use cases.
type txManager struct {
	db *gorm.DB
}

func NewTxManager(db *gorm.DB) *txManager {
	return &txManager{db: db}
}

// TransactionExec executes fn inside a transaction.
// If fn returns an error → rollback, else commit.
func (t *txManager) TransactionExec(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		return fn(txCtx) // rollback/commit managed by GORM
	})
}

// GetTx retrieves the *gorm.DB bound to the current transaction (if any).
// Falls back to base DB if no transaction is active.
func (t *txManager) GetTx(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx
	}
	return t.db
}
