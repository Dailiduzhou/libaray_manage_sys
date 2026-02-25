package repositories

import (
	"github.com/Dailiduzhou/library_manage_sys/models"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/mock_$GOFILE -package=mocks

// BookRepository defines the interface for book data access.
//
// Conventions:
//   - Methods that only read state and do not express locking or transactional intent are documented as READ-ONLY.
//   - Methods that participate in stock/borrow flows or require row-level consistency are documented as LOCK-SENSITIVE
//     to signal they are expected to be used within higher-level transactional contexts (e.g. borrow/return, stock updates).
type BookRepository interface {
	// CreateBook creates a new book aggregate. READ-ONLY on stock invariants;
	// callers are responsible for choosing initial stock values.
	CreateBook(book *models.Book) error

	// GetBookByID retrieves a single book by its ID. READ-ONLY.
	GetBookByID(id uint) (*models.Book, error)

	// ListBooks returns all books without filtering. READ-ONLY.
	ListBooks() ([]*models.Book, error)

	// FindBooksByTitle performs a fuzzy search by title. READ-ONLY.
	FindBooksByTitle(title string) ([]*models.Book, error)

	// FindBooksByAuthor performs a fuzzy search by author. READ-ONLY.
	FindBooksByAuthor(author string) ([]*models.Book, error)

	// UpdateBook applies general attribute changes (title, author, summary, cover path, etc.).
	// This method is READ-ONLY with respect to stock semantics; it does not express locking intent itself.
	UpdateBook(book *models.Book) error

	// UpdateBookStockWithTotal updates stock and total stock for a single book while enforcing domain invariants.
	// LOCK-SENSITIVE: intended for use in stock management flows where concurrent updates must be coordinated.
	// Implementations are expected to ensure that resulting stock does not violate configured invariants.
	UpdateBookStockWithTotal(id uint, newStock, newTotalStock int) (*models.Book, error)

	// DeleteBook removes a book by ID when it is safe to do so.
	// LOCK-SENSITIVE: callers use this in flows that must respect borrow/stock constraints.
	DeleteBook(id uint) error

	// LockBookForUpdate retrieves a book by ID with the intention of performing a subsequent lock-aware update.
	// LOCK-SENSITIVE: used by borrow/return and other stock flows that require row-level consistency.
	LockBookForUpdate(id uint) (*models.Book, error)

	// SaveLockedBook persists a previously locked book after in-memory changes, such as stock adjustments.
	// LOCK-SENSITIVE: callers are expected to invoke this only after obtaining a lock via LockBookForUpdate.
	SaveLockedBook(book *models.Book) error
}

//go:generate mockgen -source=$GOFILE -destination=mocks/mock_$GOFILE -package=mocks

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	FindAll() ([]*models.User, error)
}

//go:generate mockgen -source=$GOFILE -destination=mocks/mock_$GOFILE -package=mocks

// BorrowRepository defines the interface for borrow record data access.
//
// Conventions:
//   - General query methods are READ-ONLY and do not imply row-level locking.
//   - Methods used by borrow/return flows provide lock-sensitive variants so
//     service logic can keep transactional consistency without using ORM details.
type BorrowRepository interface {
	// Create persists a new borrow record. LOCK-SENSITIVE in borrow flow.
	Create(record *models.BorrowRecord) error

	// GetByID retrieves a record by primary key. READ-ONLY.
	GetByID(id uint) (*models.BorrowRecord, error)

	// GetByUserID lists all records by user. READ-ONLY.
	GetByUserID(userID uint) ([]*models.BorrowRecord, error)

	// GetByBookID lists all records by book. READ-ONLY.
	GetByBookID(bookID uint) ([]*models.BorrowRecord, error)

	// GetByUserAndBook retrieves one record by user and book. READ-ONLY.
	GetByUserAndBook(userID uint, bookID uint) (*models.BorrowRecord, error)

	// FindAll returns all borrow records. READ-ONLY.
	FindAll() ([]*models.BorrowRecord, error)

	// LockBorrowedRecordForUpdate retrieves an active borrowed record with row lock.
	// LOCK-SENSITIVE: intended for return flow updates.
	LockBorrowedRecordForUpdate(userID uint, bookID uint) (*models.BorrowRecord, error)

	// UpdateBorrowRecord persists status/return-date changes on a locked record.
	// LOCK-SENSITIVE: should be called after lock-sensitive read.
	UpdateBorrowRecord(record *models.BorrowRecord) error
}
