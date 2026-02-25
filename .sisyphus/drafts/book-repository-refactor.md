# Draft: Book Repository Refactor

## Requirements (confirmed)
- Move all Book-related DB operations from service/controller layers into a proper repository-backed design.
- Define a clear `BookRepository` interface with no `*gorm.DB` in method signatures.
- Implement a concrete `gormBookRepository` struct that owns an internal `db *gorm.DB` and encapsulates all GORM usage.
- Provide a constructor `NewGormBookRepository(db *gorm.DB) BookRepository` and adjust wiring accordingly.
- Ensure all Book-related DB operations, including those currently in `BorrowService`, go through `BookRepository`.
- Keep Book business rules (duplicate detection, stock validation, delete constraints) in the service layer, not in the repository.

## Technical Decisions
- `BookRepository` is an ORM-agnostic interface; only the `gormBookRepository` implementation knows about GORM and `*gorm.DB`.
- `gormBookRepository` holds a base `db *gorm.DB` connection and may internally create transaction-bound instances as needed.
- Introduce a `Transactor`/Unit-of-Work abstraction that starts GORM transactions and provides transaction-scoped repository instances to services.
- Expose explicit lock-aware methods in `BookRepository` (e.g. `GetByIDForUpdate`, `FindAvailableCopyForUpdate`) to support row-level locking without leaking ORM-specific APIs.
- `BorrowService` and other services that mutate Book state will orchestrate use cases inside `Transactor` transactions, using only repository methods.
- Map DB and GORM-specific errors to domain-level errors (e.g. `ErrBookNotFound`, `ErrBookUnavailable`, `ErrConcurrentModification`) inside the repository implementation.

## Research Findings
- `Book` model and related DTOs live in `backend/models/models.go` and `backend/models/requests.go`.
- `BookService` in `backend/services/book_service.go` currently performs all GORM operations directly via `config.DB`.
- `BookRepository` and `GormBookRepository` already exist in `backend/repositories/interfaces.go` and `backend/repositories/book_repo.go`, along with mocks and integration tests.
- Controllers (`BookHandler`) mostly delegate to `BookService` but have a few direct `config.DB` usages and legacy controller functions still using raw GORM.
- `BorrowService` contains cross-aggregate transactional logic (Book + BorrowRecord) and currently manipulates `Book` rows directly with GORM, including row locking.
- `main.go` wires `bookService := services.NewBookService()` and `bookRepo := &repositories.GormBookRepository{}`, injecting both into `BookHandler` but with `bookService` not yet using `bookRepo`.

## Open Questions
- Exact shape of the `Transactor`/Unit-of-Work abstraction (interface names, context usage, and how repository sets are provided).
- Which Book operations require lock-aware repository methods vs. simple read methods (e.g. only stock updates and delete, or also some search flows).
- Whether to model concurrency using row-level locking only or to introduce optimistic concurrency via version fields.
- How aggressively to deprecate/remove legacy controller functions that use `config.DB` for Book operations (immediate removal vs. soft deprecation).

## Scope Boundaries
- INCLUDE: Refactor Book-related DB access in `BookService`, `BookHandler`, and `BorrowService` to use a properly injected `BookRepository` based on GORM.
- INCLUDE: Define and document the `gormBookRepository` struct, its constructor, and any necessary interface adjustments, including lock-aware methods.
- INCLUDE: Introduce a `Transactor`/Unit-of-Work abstraction to coordinate transactions across multiple repositories.
- INCLUDE: Update wiring in `main.go` and service constructors in `backend/services/services.go` to inject `BookRepository` and `Transactor` where needed.
- EXCLUDE: Full refactor of User and Borrow repositories/services beyond the Book-related parts (may be planned but not implemented in this plan unless explicitly requested).
- EXCLUDE: Non-database concerns such as file storage helpers, authentication, or unrelated endpoints.