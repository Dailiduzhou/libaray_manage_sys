package repositories

import (
	"errors"

	"github.com/Dailiduzhou/library_manage_sys/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// gormBorrowRepository implements BorrowRepository using GORM.
type gormBorrowRepository struct {
	db *gorm.DB
}

// NewGormBorrowRepository creates a new BorrowRepository using GORM.
func NewGormBorrowRepository(db *gorm.DB) BorrowRepository {
	return &gormBorrowRepository{db: db}
}

// Create creates a new borrow record in the database.
func (r *gormBorrowRepository) Create(record *models.BorrowRecord) error {
	return r.db.Omit("User", "Book").Create(record).Error
}

// GetByID retrieves a borrow record by ID.
func (r *gormBorrowRepository) GetByID(id uint) (*models.BorrowRecord, error) {
	var record models.BorrowRecord
	err := r.db.First(&record, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

// GetByUserID retrieves all borrow records for a user.
func (r *gormBorrowRepository) GetByUserID(userID uint) ([]*models.BorrowRecord, error) {
	var records []*models.BorrowRecord
	if err := r.db.Where("user_id = ?", userID).Order("id DESC").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// GetByBookID retrieves all borrow records for a book.
func (r *gormBorrowRepository) GetByBookID(bookID uint) ([]*models.BorrowRecord, error) {
	var records []*models.BorrowRecord
	if err := r.db.Where("book_id = ?", bookID).Order("id DESC").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// GetByUserAndBook retrieves a borrow record by user ID and book ID.
func (r *gormBorrowRepository) GetByUserAndBook(userID uint, bookID uint) (*models.BorrowRecord, error) {
	var record models.BorrowRecord
	err := r.db.Where("user_id = ? AND book_id = ?", userID, bookID).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

// FindAll retrieves all borrow records.
func (r *gormBorrowRepository) FindAll() ([]*models.BorrowRecord, error) {
	var records []*models.BorrowRecord
	if err := r.db.Order("id DESC").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// LockBorrowedRecordForUpdate loads an active borrowed record with a row lock.
func (r *gormBorrowRepository) LockBorrowedRecordForUpdate(userID uint, bookID uint) (*models.BorrowRecord, error) {
	var record models.BorrowRecord
	err := r.db.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND book_id = ? AND status = ?", userID, bookID, "borrowed").
		First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

// UpdateBorrowRecord persists an updated borrow record.
func (r *gormBorrowRepository) UpdateBorrowRecord(record *models.BorrowRecord) error {
	return r.db.Save(record).Error
}
