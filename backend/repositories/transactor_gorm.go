package repositories

import (
	"context"

	"gorm.io/gorm"
)

// gormTransactor implements Transactor using a *gorm.DB handle.
// It relies on GORM's Transaction helper for commit/rollback semantics
// and supports optional nested transactions by propagating an existing
// *gorm.DB from the context when present.
type gormTransactor struct {
	db *gorm.DB
}

// context key type to avoid collisions.
type txContextKey struct{}

var txKey = txContextKey{}

// NewGormTransactor constructs a Transactor backed by the provided *gorm.DB.
func NewGormTransactor(db *gorm.DB) Transactor {
	return &gormTransactor{db: db}
}

// TxFromContext returns the transaction-bound *gorm.DB propagated by gormTransactor.
// It enables repository interfaces that still require *gorm.DB to participate in the same transaction.
func TxFromContext(ctx context.Context) (*gorm.DB, bool) {
	if ctx == nil {
		return nil, false
	}
	tx, ok := ctx.Value(txKey).(*gorm.DB)
	if !ok || tx == nil {
		return nil, false
	}
	return tx, true
}

// gormTxRepositories provides transaction-scoped repositories bound to a *gorm.DB transaction.
type gormTxRepositories struct {
	books   BookRepository
	borrows BorrowRepository
}

func (r *gormTxRepositories) Books() BookRepository {
	return r.books
}

func (r *gormTxRepositories) Borrows() BorrowRepository {
	return r.borrows
}

// WithinTransaction executes fn within a single logical transaction boundary.
//
// It reuses an existing *gorm.DB stored in ctx (if present) to provide
// optional nested transaction semantics without committing or rolling back
// inner scopes independently. Otherwise, it starts a new GORM transaction
// via db.WithContext(ctx).Transaction.
func (t *gormTransactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context, repos TxRepositories) error) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// Check for existing transaction in context to support nested semantics.
	if existingTx, ok := ctx.Value(txKey).(*gorm.DB); ok && existingTx != nil {
		repos := &gormTxRepositories{
			books:   NewGormBookRepository(existingTx),
			borrows: NewGormBorrowRepository(existingTx),
		}
		return fn(ctx, repos)
	}

	// Start a new transaction using GORM's Transaction helper.
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		repos := &gormTxRepositories{
			books:   NewGormBookRepository(tx),
			borrows: NewGormBorrowRepository(tx),
		}
		return fn(txCtx, repos)
	})
}
