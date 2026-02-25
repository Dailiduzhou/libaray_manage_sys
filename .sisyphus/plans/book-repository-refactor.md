# Book Repository & Transaction Refactor

## TL;DR
> **Summary**: Refactor the Book-related persistence layer so that all database operations for books (including those in `BookService`, `BookHandler`, and `BorrowService`) go exclusively through a clean `BookRepository` interface implemented by `gormBookRepository` with an internal `*gorm.DB`, coordinated via a transaction/Unit-of-Work abstraction.
> **Deliverables**:
> - ORM-agnostic `BookRepository` interface
> - `gormBookRepository` implementation with `db *gorm.DB` and lock-aware methods
> - `Transactor`/Unit-of-Work abstraction providing transaction-scoped repositories
> - Refactored `BookService`, `BorrowService`, and `BookHandler` with no direct GORM/`config.DB` usage
> - Updated wiring in `main.go` and service constructors
> **Effort**: Large
> **Parallel**: YES - 4 waves
> **Critical Path**: Define interfaces → Implement `gormBookRepository` + `Transactor` → Refactor services → Clean up controllers/tests

## Context
### Original Request
Move database operations from the Service layer into the Repository layer, using Book as the example entity, and define a `gormBookRepository` with an internal `db *gorm.DB` and a constructor `NewGormBookRepository(db *gorm.DB) BookRepository` where all GORM CRUD logic is implemented.

### Interview Summary
- The existing codebase already has `BookRepository` and `GormBookRepository` implementations and integration tests, but `BookService` still uses `config.DB` directly.
- Controllers (`BookHandler`) delegate to `BookService` but leak some direct DB usage; legacy controller functions still use raw GORM.
- `BorrowService` currently uses GORM directly to manipulate `Book` rows, including transactional updates and row-level locks.
- You decided that:
  - `BookRepository` methods must NOT expose `*gorm.DB` in their signatures; only `gormBookRepository` owns `db *gorm.DB` internally.
  - All Book-related DB operations, including those in `BorrowService`, must go through `BookRepository`.

### Metis Review (gaps addressed)
- Identified need for a `Transactor`/Unit-of-Work abstraction to coordinate transactions across repositories without exposing GORM to services.
- Highlighted the importance of lock-aware repository methods (`GetByIDForUpdate`, `FindAvailableCopyForUpdate`) for stock updates and delete constraints.
- Flagged concurrency edge cases (deadlocks, nested transactions, long-lived transactions, lock timeouts) and the need for a consistent lock acquisition order and clear transactional APIs.
- Recommended strict layering guardrails: no `config.DB` or `gorm` imports in services/handlers, repository methods returning domain errors instead of raw DB errors, and context-aware APIs.
- Emphasized acceptance criteria around architectural invariants, repository coverage, transaction behavior, locking correctness, and observability.

## Work Objectives
### Core Objective
Establish a clean, testable, and transaction-safe persistence layer for Book entities such that all Book-related DB operations are routed through a `BookRepository` abstraction implemented by `gormBookRepository`, with services orchestrating business rules via a `Transactor`/Unit-of-Work abstraction and no direct GORM usage outside the repository layer.

### Deliverables
- Updated `BookRepository` interface in `backend/repositories/interfaces.go` with ORM-agnostic method signatures, including lock-aware operations.
- Concrete `gormBookRepository` struct in `backend/repositories/book_repo.go` with `db *gorm.DB`, implementing all `BookRepository` methods using GORM.
- Constructor `NewGormBookRepository(db *gorm.DB) BookRepository` (and any transaction-bound variants) in `backend/repositories/repositories.go` or equivalent.
- `Transactor`/Unit-of-Work abstraction in a suitable package (e.g., `backend/repositories` or `backend/infra`) with GORM-backed implementation.
- Refactored `BookService` to depend on `BookRepository` (and `Transactor` where needed) instead of `config.DB`.
- Refactored `BorrowService` to use `BookRepository` for all Book-related DB operations.
- Refactored `BookHandler` and removal/rewire of legacy controller functions so they no longer access `config.DB`.
- Updated wiring in `backend/main.go` and `backend/services/services.go` to construct repositories, services, and transactors appropriately.
- Updated and/or new tests (unit + integration) covering repository behavior, transactor semantics, and key Book/Borrow flows.

### Definition of Done
- No direct references to `config.DB` or `gorm` in `BookService`, `BorrowService`, `BookHandler`, or any Book-related controllers/services.
- All Book-related DB access is performed via `BookRepository` (and other repositories for non-Book tables) in application code.
- `gormBookRepository` and `Transactor` tests verify transaction commit/rollback semantics, error propagation, and locking behavior.
- Existing Book repository integration tests are updated and passing, and new tests cover lock-aware methods and transaction usage.
- Borrow/return flows are verified to maintain correct stock/inventory invariants under concurrent access.

### Must Have
- Clean `BookRepository` interface with no ORM-specific types.
- `gormBookRepository` implementation fully encapsulating GORM usage, including row-locking and error mapping.
- `Transactor` abstraction coordinating transactions, providing transaction-bound repository instances, and handling rollback on errors/panics.
- Services and handlers using only repositories and transactors for DB interaction.
- Comprehensive tests for critical Book and Borrow flows, including locking and concurrency-sensitive paths.

### Must NOT Have
- Any new direct usage of `config.DB` or `gorm` in services, controllers, or other application layers.
- Repository methods that expose `*gorm.DB` or other GORM types in their signatures.
- Business rules (e.g., stock validation, delete constraints) buried inside repository implementations; these must remain in services.
- Hidden nested transactions or ambiguous transaction semantics; behavior must be explicit and well-documented.

## Verification Strategy
> ZERO HUMAN INTERVENTION — all verification is agent-executed.
- Test decision: tests-after using existing Go test infrastructure (`go test ./...`) with focus on repository and service packages.
- QA policy: Each TODO below includes targeted tests or test updates to validate correctness.
- Evidence: `.sisyphus/evidence/task-{N}-{slug}.txt` or `.log` files capturing `go test` outputs, along with any concurrency stress test logs.

## Execution Strategy
### Parallel Execution Waves
Wave 1: Interface & abstraction design
- Define/refine `BookRepository` interface and `Transactor` abstraction.
- Sketch `gormBookRepository` structure and transaction binding model.

Wave 2: Infrastructure implementation
- Implement `gormBookRepository` with internal `db *gorm.DB` and lock-aware methods.
- Implement `Transactor`/Unit-of-Work with GORM-backed logic and basic tests.

Wave 3: Service & handler refactor
- Refactor `BookService` and `BorrowService` to use `BookRepository` + `Transactor`.
- Clean up `BookHandler` and legacy controller functions.

Wave 4: Testing & hardening
- Update/add tests for repository, transactor, BookService, BorrowService, and controllers.
- Concurrency/locking verification and final regression suite.

### Dependency Matrix
- Wave 1 must complete before Waves 2–4.
- Wave 2 is prerequisite for Wave 3 (services depend on concrete repository/transactor implementations).
- Wave 3 must complete before final verification in Wave 4.

### Agent Dispatch Summary
- Wave 1: `ultrabrain` or `deep` for design-heavy tasks.
- Wave 2: `quick`/`unspecified-high` for focused implementation of infra components.
- Wave 3: `deep` for careful refactor of services/handlers.
- Wave 4: `unspecified-high` for thorough testing and concurrency checks.

## TODOs

- [ ] 1. Analyze existing BookRepository and GormBookRepository

  **What to do**: Review `backend/repositories/interfaces.go` and `backend/repositories/book_repo.go` to understand the current `BookRepository` interface, `GormBookRepository` implementation, and related tests, identifying methods to keep, methods to change, and gaps for lock-aware operations.
  **Must NOT do**: Do not change any code yet; this task is analysis-only.

  **Recommended Agent Profile**:
  - Category: `quick` — Reason: Single-pass code reading and comparison.
  - Skills: [] — No special skills required beyond repo exploration.
  - Omitted: [`git-master`] — No git history manipulation required at this stage.

  **Parallelization**: Can Parallel: YES | Wave 1 | Blocks: [2, 3] | Blocked By: []

  **References**:
  - Interface: `backend/repositories/interfaces.go` — current `BookRepository` methods and signatures.
  - Implementation: `backend/repositories/book_repo.go` — `GormBookRepository` struct and methods.
  - Tests: `backend/repositories/integration_test.go` — current expectations for repository behavior.

  **Acceptance Criteria**:
  - [ ] A short design note (e.g., `.sisyphus/evidence/task-1-analysis.txt`) summarizing existing BookRepository methods, their GORM usage, and which methods need to be updated/extended for the new abstraction.

  **QA Scenarios**
  ```
  Scenario: Document existing BookRepository shape
    Tool: Bash
    Steps: Inspect interfaces.go, book_repo.go, and integration_test.go; write a summary file under .sisyphus/evidence.
    Expected: Summary includes all current methods, their signatures, and high-level behavior.
    Evidence: .sisyphus/evidence/task-1-analysis.txt

  Scenario: Identify missing lock-aware operations
    Tool: Bash
    Steps: Compare service use cases (update/delete/borrow flows) with BookRepository methods; list any operations requiring locking not present in the interface.
    Expected: Clear list of potential new lock-aware methods.
    Evidence: .sisyphus/evidence/task-1-lock-gaps.txt
  ```

  **Commit**: NO | Message: N/A | Files: []

- [ ] 2. Design ORM-agnostic BookRepository interface

  **What to do**: Refine/replace the existing `BookRepository` interface in `backend/repositories/interfaces.go` to remove `*gorm.DB` parameters and define a clean set of methods (CRUD, search, and lock-aware operations) aligned with current and future Book use cases.
  **Must NOT do**: Do not introduce GORM-specific types or transaction concepts into the interface.

  **Recommended Agent Profile**:
  - Category: `deep` — Reason: Requires careful API design balancing current needs with future extensibility.
  - Skills: []
  - Omitted: [`git-master`] — No history rewriting.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: [4, 5, 6, 7] | Blocked By: [1]

  **References**:
  - `backend/repositories/interfaces.go` — current `BookRepository` and related repo interfaces.
  - `backend/services/book_service.go` — all existing Book use cases to ensure coverage.
  - `backend/services/borrow_service.go` — Book-related borrow/return flows requiring locking.

  **Acceptance Criteria**:
  - [ ] `BookRepository` interface has no `*gorm.DB` parameters or GORM types.
  - [ ] Methods cover: create, get by ID, update, delete, list/search, and any required lock-aware operations for stock updates and delete constraints.
  - [ ] Interface documentation (comments) clearly distinguishes read-only methods from lock-aware/transactional methods.

  **QA Scenarios**
  ```
  Scenario: Verify BookRepository interface is ORM-agnostic
    Tool: Bash
    Steps: Use ripgrep to search for "*gorm.DB" and "gorm." within interfaces.go.
    Expected: No occurrences within the BookRepository definition.
    Evidence: .sisyphus/evidence/task-2-orm-agnostic.txt

  Scenario: Cross-check interface methods against service use cases
    Tool: Bash
    Steps: Enumerate all Book-related methods in BookService and BorrowService; verify each can be expressed via BookRepository methods.
    Expected: No Book-related DB operation in services is left without a corresponding repository method.
    Evidence: .sisyphus/evidence/task-2-usecase-coverage.txt
  ```

  **Commit**: YES | Message: `refactor(repository): define orm-agnostic BookRepository interface` | Files: [`backend/repositories/interfaces.go`]

- [ ] 3. Design Transactor / Unit-of-Work abstraction

  **What to do**: Define a `Transactor` interface (and, if needed, a `UnitOfWork` or `Repositories` aggregate) that allows services to execute functions within a transaction and access transaction-scoped repository instances.
  **Must NOT do**: Do not expose `*gorm.DB` or GORM-specific transaction types in the public interface.

  **Recommended Agent Profile**:
  - Category: `deep` — Reason: Involves transaction semantics, concurrency, and API design.
  - Skills: []

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: [4, 5, 6, 7] | Blocked By: [1]

  **References**:
  - `backend/services/borrow_service.go` — current transactional patterns for borrow/return flows.
  - `backend/services/book_service.go` — current use of `config.DB.Transaction` and `clause.Locking`.
  - GORM docs — transaction and context patterns (consult via external docs if needed).

  **Acceptance Criteria**:
  - [ ] `Transactor` interface is defined in an appropriate package (e.g., `backend/repositories` or `backend/infra`) with a method like `WithinTransaction(ctx context.Context, fn func(repos Repositories) error) error`.
  - [ ] `Repositories` struct (or equivalent) exposes `BookRepository` (and potentially other repos) without GORM types.
  - [ ] Interface docs clearly describe behavior on success, error, and panic (commit vs rollback) and mention nesting semantics.

  **QA Scenarios**
  ```
  Scenario: Confirm Transactor interface is ORM-free
    Tool: Bash
    Steps: Search the file defining Transactor for "gorm" or "*gorm.DB" references.
    Expected: No ORM-specific types in the interface.
    Evidence: .sisyphus/evidence/task-3-orm-free.txt

  Scenario: Check transaction semantics documentation
    Tool: Bash
    Steps: Inspect comments on the Transactor interface and associated types.
    Expected: Clear description of commit/rollback behavior and nested transaction handling.
    Evidence: .sisyphus/evidence/task-3-semantics.txt
  ```

  **Commit**: YES | Message: `feat(tx): introduce Transactor and UnitOfWork interfaces` | Files: [`backend/repositories/transactor.go`] (or appropriate location)

- [ ] 4. Implement gormBookRepository with internal *gorm.DB

  **What to do**: Implement the `gormBookRepository` struct in `backend/repositories/book_repo.go` with an internal `db *gorm.DB` field and update all methods to satisfy the new `BookRepository` interface, including CRUD, search, and lock-aware operations, mapping DB errors to domain errors.
  **Must NOT do**: Do not leak `*gorm.DB` or GORM types through the `BookRepository` interface.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` — Reason: Requires careful implementation and error mapping.
  - Skills: []

  **Parallelization**: Can Parallel: NO | Wave 2 | Blocks: [8, 9, 10] | Blocked By: [2]

  **References**:
  - `backend/repositories/book_repo.go` — current `GormBookRepository` implementation.
  - `backend/repositories/integration_test.go` — integration tests for existing behavior.
  - `backend/models/models.go` — Book model definition.

  **Acceptance Criteria**:
  - [ ] `gormBookRepository` struct has a `db *gorm.DB` field and no exported ORM types.
  - [ ] All `BookRepository` methods are implemented, including lock-aware ones using appropriate GORM locking clauses.
  - [ ] Domain errors (`ErrBookNotFound`, `ErrBookInUse`, etc.) are returned instead of raw GORM/SQL errors wherever applicable.
  - [ ] Existing integration tests are updated to use the new constructor and all pass.

  **QA Scenarios**
  ```
  Scenario: Run BookRepository integration tests
    Tool: Bash
    Steps: Execute `go test ./backend/repositories -run TestBook*`.
    Expected: All Book-related repository tests pass.
    Evidence: .sisyphus/evidence/task-4-integration-tests.txt

  Scenario: Verify lock-aware methods use correct clauses
    Tool: Bash
    Steps: Inspect gormBookRepository methods implementing *ForUpdate operations for usage of clause.Locking or equivalent.
    Expected: Locking is applied only in designated methods and not in read-only ones.
    Evidence: .sisyphus/evidence/task-4-locking-impl.txt
  ```

  **Commit**: YES | Message: `feat(repository): implement gormBookRepository with internal db` | Files: [`backend/repositories/book_repo.go`, `backend/repositories/repositories.go`, `backend/repositories/integration_test.go`]

- [ ] 5. Implement GORM-backed Transactor

  **What to do**: Implement a concrete GORM-based `Transactor` that wraps `*gorm.DB.Transaction`, constructs transaction-scoped repository instances (including `gormBookRepository`), and ensures proper commit/rollback behavior on success, error, and panic.
  **Must NOT do**: Do not expose GORM transaction handles to services or leak `*gorm.DB` outside this implementation.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` — Reason: Requires careful handling of transaction semantics and error propagation.
  - Skills: []

  **Parallelization**: Can Parallel: NO | Wave 2 | Blocks: [8, 9, 10] | Blocked By: [3, 4]

  **References**:
  - `backend/config/config.go` — `config.DB` definition and setup.
  - GORM docs — `db.Transaction` usage and panic handling.
  - Newly defined `Transactor` and `Repositories` interfaces/types.

  **Acceptance Criteria**:
  - [ ] `gormTransactor` (or similar) uses `config.DB` (or injected `*gorm.DB`) to run transactions.
  - [ ] Within a transaction, repository instances (including `gormBookRepository`) are constructed with the transaction handle.
  - [ ] On error or panic in the callback, the transaction is rolled back and error/panic propagated appropriately.
  - [ ] Unit tests validate commit, rollback on error, and rollback on panic.

  **QA Scenarios**
  ```
  Scenario: Verify transaction commit/rollback behavior
    Tool: Bash
    Steps: Write and run tests where the Transactor callback performs a simple insert and returns nil/error/panics.
    Expected: Data committed only on nil, rolled back on error/panic.
    Evidence: .sisyphus/evidence/task-5-tx-tests.txt

  Scenario: Ensure repository instances use transaction handle
    Tool: Bash
    Steps: Instrument or assert within tests that BookRepository methods called inside the transaction use the tx-specific DB handle.
    Expected: No repository method uses the base DB when inside a transaction.
    Evidence: .sisyphus/evidence/task-5-tx-binding.txt
  ```

  **Commit**: YES | Message: `feat(tx): implement gorm-backed Transactor` | Files: [`backend/repositories/transactor_gorm.go`, tests]

- [ ] 6. Refactor BookService to use BookRepository + Transactor

  **What to do**: Refactor `backend/services/book_service.go` so that all Book-related DB operations use `BookRepository` and `Transactor` instead of `config.DB` and direct GORM calls, keeping business rules in the service layer.
  **Must NOT do**: Do not re-introduce `config.DB` or `gorm` imports into the service.

  **Recommended Agent Profile**:
  - Category: `deep` — Reason: Requires careful migration of logic while preserving behavior and error semantics.
  - Skills: []

  **Parallelization**: Can Parallel: NO | Wave 3 | Blocks: [11, 12, 13] | Blocked By: [2, 4, 5]

  **References**:
  - `backend/services/book_service.go` — current service implementation and error types.
  - `backend/repositories/interfaces.go` — new `BookRepository` interface.
  - `Transactor` implementation and types.

  **Acceptance Criteria**:
  - [ ] `bookService` struct holds a `BookRepository` (and `Transactor` if needed) instead of relying on `config.DB`.
  - [ ] All DB operations (create, list, update, delete, get by ID) are expressed via repository methods, with transactions used where necessary.
  - [ ] Business rules and domain errors remain in the service layer and are mapped as before.
  - [ ] No `config.DB` or `gorm` imports remain in `book_service.go`.

  **QA Scenarios**
  ```
  Scenario: Verify BookService has no direct DB access
    Tool: Bash
    Steps: Use ripgrep to search for "config.DB" and "gorm" in backend/services/book_service.go.
    Expected: No matches.
    Evidence: .sisyphus/evidence/task-6-no-gorm.txt

  Scenario: Run BookService-related tests
    Tool: Bash
    Steps: Execute `go test ./backend/services -run TestBook*` (or equivalent).
    Expected: All BookService tests pass.
    Evidence: .sisyphus/evidence/task-6-service-tests.txt
  ```

  **Commit**: YES | Message: `refactor(service): move BookService DB logic to BookRepository` | Files: [`backend/services/book_service.go`, `backend/services/services.go`]

- [ ] 7. Refactor BorrowService to use BookRepository for Book operations

  **What to do**: Refactor `backend/services/borrow_service.go` so that all Book-related DB operations (including row-locking for borrow/return flows) use `BookRepository` within `Transactor` transactions, eliminating direct GORM usage for Book entities.
  **Must NOT do**: Do not change Borrow-related repository patterns beyond what is necessary to route Book operations through `BookRepository`.

  **Recommended Agent Profile**:
  - Category: `deep` — Reason: Cross-aggregate transactions and locking.
  - Skills: []

  **Parallelization**: Can Parallel: NO | Wave 3 | Blocks: [11, 12, 13] | Blocked By: [2, 3, 4, 5]

  **References**:
  - `backend/services/borrow_service.go` — current implementation with GORM-based transactions and locks.
  - `BookRepository` interface — especially lock-aware methods.
  - `Transactor` and `Repositories` types.

  **Acceptance Criteria**:
  - [ ] `BorrowService` no longer imports `gorm` or uses `config.DB` directly.
  - [ ] All Book-related DB actions in borrow/return flows use `BookRepository` within `Transactor.WithinTransaction` calls.
  - [ ] Locking semantics are preserved or improved, and documentation is updated accordingly.
  - [ ] Borrow-related unit tests and integration tests pass.

  **QA Scenarios**
  ```
  Scenario: Verify BorrowService has no direct DB access to Book
    Tool: Bash
    Steps: Search borrow_service.go for "config.DB", "gorm", and direct Book model queries.
    Expected: No direct GORM usage for Book remains.
    Evidence: .sisyphus/evidence/task-7-no-gorm.txt

  Scenario: Run BorrowService tests
    Tool: Bash
    Steps: Execute `go test ./backend/services -run TestBorrow*` (or equivalent), plus any integration tests involving borrow flows.
    Expected: All tests pass with new repository-based implementation.
    Evidence: .sisyphus/evidence/task-7-service-tests.txt
  ```

  **Commit**: YES | Message: `refactor(service): route BorrowService Book operations through BookRepository` | Files: [`backend/services/borrow_service.go`, possibly `backend/services/services.go`]

- [ ] 8. Refactor BookHandler and remove direct DB leaks

  **What to do**: Update `backend/controllers/controller.go` (Book-related handler methods) to ensure no direct `config.DB` usage remains and that all Book operations go through `BookService` (which now uses `BookRepository`). Deprecate or remove legacy controller functions that directly use GORM for Book operations.
  **Must NOT do**: Do not reintroduce DB access in handlers.

  **Recommended Agent Profile**:
  - Category: `quick` — Reason: Mostly wiring and cleanup.
  - Skills: []

  **Parallelization**: Can Parallel: YES | Wave 3 | Blocks: [11, 12, 13] | Blocked By: [6]

  **References**:
  - `backend/controllers/controller.go` — BookHandler methods and legacy controller functions.
  - `backend/routes/book.go` — which handlers are actually registered.

  **Acceptance Criteria**:
  - [ ] `BookHandler` uses `BookService` only and has no direct `config.DB` or `gorm` usage.
  - [ ] Legacy controller functions that bypass `BookService` are either removed or clearly marked as deprecated and updated to call `BookService`.
  - [ ] All Book-related routes still function correctly.

  **QA Scenarios**
  ```
  Scenario: Verify no DB access in BookHandler
    Tool: Bash
    Steps: Search controller.go for "config.DB" and "gorm" within Book-related sections.
    Expected: No matches.
    Evidence: .sisyphus/evidence/task-8-no-gorm.txt

  Scenario: Smoke test Book HTTP endpoints
    Tool: interactive_bash or Bash with HTTP client
    Steps: Use curl or a small test client to hit create, list, update, delete endpoints.
    Expected: All endpoints respond with expected status codes and data shapes.
    Evidence: .sisyphus/evidence/task-8-http-smoke.txt
  ```

  **Commit**: YES | Message: `refactor(controller): remove direct DB usage from BookHandler` | Files: [`backend/controllers/controller.go`, possibly `backend/routes/book.go`]

- [ ] 9. Update wiring in main.go and services.go

  **What to do**: Update `backend/main.go` and `backend/services/services.go` to construct `gormBookRepository` via `NewGormBookRepository(config.DB)`, instantiate `Transactor`, and inject these into `BookService`, `BorrowService`, and `BookHandler` appropriately.
  **Must NOT do**: Do not bypass the DI/wiring layer with ad-hoc instantiations in other parts of the application.

  **Recommended Agent Profile**:
  - Category: `quick` — Reason: Focused wiring changes.
  - Skills: []

  **Parallelization**: Can Parallel: YES | Wave 3 | Blocks: [11, 12, 13] | Blocked By: [4, 5, 6, 7]

  **References**:
  - `backend/main.go` — current wiring for services and handlers.
  - `backend/services/services.go` — constructors for services.
  - New `BookRepository` and `Transactor` constructors.

  **Acceptance Criteria**:
  - [ ] `main.go` uses `NewGormBookRepository(config.DB)` (or equivalent) and constructs a `Transactor` instance.
  - [ ] `NewBookService` (and `NewBorrowService` as needed) accept `BookRepository` and `Transactor` as dependencies.
  - [ ] No direct `&GormBookRepository{}` instantiations remain outside the repository wiring.

  **QA Scenarios**
  ```
  Scenario: Build the application
    Tool: Bash
    Steps: Run `go build ./...` from the backend root.
    Expected: Build succeeds with no missing dependencies.
    Evidence: .sisyphus/evidence/task-9-build.txt

  Scenario: Quick runtime smoke test
    Tool: Bash
    Steps: Run the application and hit a couple of Book endpoints.
    Expected: Application starts without panic, and endpoints work.
    Evidence: .sisyphus/evidence/task-9-runtime-smoke.txt
  ```

  **Commit**: YES | Message: `chore(wiring): inject BookRepository and Transactor into services` | Files: [`backend/main.go`, `backend/services/services.go`]

- [ ] 10. Expand and adjust tests for Book & Borrow flows

  **What to do**: Update existing tests and add new ones for `BookRepository`, `Transactor`, `BookService`, `BorrowService`, and `BookHandler`, focusing on DB interaction via repositories, transaction behavior, error mapping, and concurrency-sensitive scenarios.
  **Must NOT do**: Do not remove valuable existing tests without replacing their coverage.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` — Reason: Coordination of multiple test layers and concurrency considerations.
  - Skills: []

  **Parallelization**: Can Parallel: YES | Wave 4 | Blocks: [11, 12, 13] | Blocked By: [4, 5, 6, 7, 8, 9]

  **References**:
  - `backend/repositories/integration_test.go` — existing repository tests.
  - Service and controller test files for Books and Borrows.
  - New interfaces and implementations.

  **Acceptance Criteria**:
  - [ ] All existing Book and Borrow tests compile and pass after refactor.
  - [ ] New tests cover transaction commit/rollback, lock-aware repository methods, and error mapping.
  - [ ] At least one test simulates concurrent borrow/return operations to verify invariants.

  **QA Scenarios**
  ```
  Scenario: Run full test suite
    Tool: Bash
    Steps: Execute `go test ./...` from the project root.
    Expected: All tests pass.
    Evidence: .sisyphus/evidence/task-10-full-tests.txt

  Scenario: Concurrency stress test for borrow/return
    Tool: Bash
    Steps: Implement a test that runs multiple borrow/return operations in parallel against a small pool of books.
    Expected: No negative stock, no inconsistent borrow records, and no panics.
    Evidence: .sisyphus/evidence/task-10-concurrency.txt
  ```

  **Commit**: YES | Message: `test: cover BookRepository and transaction-based Book/Borrow flows` | Files: [relevant *_test.go files]

## Final Verification Wave
- [ ] F1. Plan Compliance Audit — oracle
- [ ] F2. Code Quality Review — unspecified-high
- [ ] F3. Real Manual QA — unspecified-high (+ playwright if UI)
- [ ] F4. Scope Fidelity Check — deep

## Commit Strategy
- Use small, focused commits aligned with TODO tasks (interface design, repository implementation, transactor, service refactors, wiring, tests).
- Avoid mixing structural refactors (e.g., legacy controller removal) with core repository/transaction changes in the same commit.

## Success Criteria
- All Book-related DB operations are routed through `BookRepository` with `gormBookRepository` as the only GORM-aware implementation.
- `BookService`, `BorrowService`, and `BookHandler` are free of direct GORM and `config.DB` usage.
- Transactional flows for borrow/return and stock updates are correct, robust, and verified via tests, including concurrency scenarios.
- The application builds and runs successfully, and all tests pass.
