package repositories

import (
	"errors"

	"github.com/Dailiduzhou/library_manage_sys/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// gormBookRepository implements BookRepository using GORM
type gormBookRepository struct {
	db *gorm.DB
}

func newGormBookRepository(db *gorm.DB) BookRepository {
	return &gormBookRepository{db: db}
}

// CreateBook inserts a new book into the database
func (r *gormBookRepository) CreateBook(book *models.Book) error {
	return r.db.Create(book).Error
}

// GetBookByID retrieves a book by its ID
func (r *gormBookRepository) GetBookByID(id uint) (*models.Book, error) {
	var book models.Book
	if err := r.db.First(&book, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &book, nil
}

// ListBooks retrieves all books from the database
func (r *gormBookRepository) ListBooks() ([]*models.Book, error) {
	var books []*models.Book
	if err := r.db.Find(&books).Error; err != nil {
		return nil, err
	}
	return books, nil
}

// FindBooksByTitle retrieves books matching the given title (partial match)
func (r *gormBookRepository) FindBooksByTitle(title string) ([]*models.Book, error) {
	var books []*models.Book
	if err := r.db.Where("title LIKE ?", "%"+title+"%").Find(&books).Error; err != nil {
		return nil, err
	}
	return books, nil
}

// FindBooksByAuthor retrieves books matching the given author (partial match)
func (r *gormBookRepository) FindBooksByAuthor(author string) ([]*models.Book, error) {
	var books []*models.Book
	if err := r.db.Where("author LIKE ?", "%"+author+"%").Find(&books).Error; err != nil {
		return nil, err
	}
	return books, nil
}

// UpdateBook updates an existing book in the database
func (r *gormBookRepository) UpdateBook(book *models.Book) error {
	return r.db.Save(book).Error
}

// UpdateBookStockWithTotal updates stock and total stock atomically while enforcing invariants
func (r *gormBookRepository) UpdateBookStockWithTotal(id uint, newStock, newTotalStock int) (*models.Book, error) {
	var book models.Book

	if newTotalStock < 0 || newStock < 0 {
		return nil, errors.New("stock values cannot be negative")
	}

	if newStock > newTotalStock {
		return nil, errors.New("stock cannot be greater than total stock")
	}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&book, id).Error; err != nil {
			return err
		}

		book.Stock = newStock
		book.TotalStock = newTotalStock

		if err := tx.Save(&book).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &book, nil
}

// DeleteBook removes a book by its ID
func (r *gormBookRepository) DeleteBook(id uint) error {
	return r.db.Delete(&models.Book{}, id).Error
}

// LockBookForUpdate retrieves a book by ID with row-level locking for updates
func (r *gormBookRepository) LockBookForUpdate(id uint) (*models.Book, error) {
	var book models.Book
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&book, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &book, nil
}

// SaveLockedBook persists changes to a previously locked book
func (r *gormBookRepository) SaveLockedBook(book *models.Book) error {
	return r.db.Save(book).Error
}
