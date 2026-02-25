
[Task 2 - BookRepository interface]
- BookRepository is now ORM-agnostic: all methods accept domain types and primitives only, with no *gorm.DB or gorm-specific types in signatures.
- The interface distinguishes READ-ONLY queries (simple lookups/search) from LOCK-SENSITIVE operations used by stock and borrow/return flows, ready for later transactional implementations.


[Task 3 - Transactor design]
- Introduced ORM-agnostic Transactor and TxRepositories interfaces in backend/repositories/transactor.go, exposing BookRepository via a transaction-scoped aggregate suitable for service-layer units of work.
- Documented commit/rollback and nested transaction semantics (join existing transaction via context, outermost scope controls commit/rollback) to guide future concrete implementations.