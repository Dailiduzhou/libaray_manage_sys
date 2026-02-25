package repositories

import "context"

// Transactor coordinates application-level transaction boundaries over repository operations.
//
// Commit / rollback semantics:
//   - If fn returns nil, the transaction MUST be committed.
//   - If fn returns a non-nil error or the context is cancelled or its deadline exceeded,
//     the transaction MUST be rolled back and the error (or a wrapped variant) returned.
//   - If fn panics, the transaction MUST be rolled back and the panic re-raised.
//
// Nested transaction semantics:
//   - If ctx already carries an active transaction started by this Transactor,
//     implementations SHOULD join / reuse that transaction instead of starting a new one.
//   - Nested calls to WithinTransaction MUST NOT commit or rollback independently;
//     commit/rollback is governed by the outermost transaction scope.
type Transactor interface {
	// WithinTransaction executes fn within a single logical transaction.
	//
	// Implementations MUST ensure that all repositories exposed via TxRepositories
	// share the same underlying transactional context for the lifetime of fn.
	WithinTransaction(ctx context.Context, fn func(ctx context.Context, repos TxRepositories) error) error
}

// TxRepositories provides transaction-scoped repository instances for use inside
// a Transactor-controlled transaction scope.
//
// Implementations should bind returned repositories to the current transaction
// so that LOCK-SENSITIVE operations (e.g. stock, borrow/return flows) participate
// in the same database transaction.
type TxRepositories interface {
	// Books returns a BookRepository whose LOCK-SENSITIVE operations participate
	// in the surrounding transaction started by Transactor.
	Books() BookRepository

	// Borrows returns a BorrowRepository bound to the same transaction.
	Borrows() BorrowRepository

	// Future methods may expose additional repositories (users, etc.)
	// as they are migrated to ORM-agnostic interfaces.
}
