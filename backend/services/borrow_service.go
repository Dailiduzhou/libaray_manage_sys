package services

import (
	"context"
	"errors"
	"time"

	"github.com/Dailiduzhou/library_manage_sys/models"
	"github.com/Dailiduzhou/library_manage_sys/repositories"
)

// BorrowService-specific errors
var (
	ErrNoStock        = errors.New("图书库存不足")
	ErrRecordNotFound = errors.New("借书记录查询失败")
)

// borrowService implements the BorrowService interface
type borrowService struct {
	borrowRepo repositories.BorrowRepository
	bookRepo   repositories.BookRepository
	tx         repositories.Transactor
}

// BorrowBook allows a user to borrow a book
func (s *borrowService) BorrowBook(userID, bookID uint) (*models.BorrowRecord, error) {
	var borrowRecord *models.BorrowRecord

	err := s.tx.WithinTransaction(context.Background(), func(ctx context.Context, repos repositories.TxRepositories) error {
		bookRepo := repos.Books()
		if bookRepo == nil {
			bookRepo = s.bookRepo
		}
		borrowRepo := repos.Borrows()
		if borrowRepo == nil {
			borrowRepo = s.borrowRepo
		}

		book, err := bookRepo.LockBookForUpdate(bookID)
		if err != nil {
			return err
		}
		if book == nil {
			return ErrBookNotFound
		}

		if book.Stock <= 0 {
			return ErrNoStock
		}

		book.Stock--
		if err := bookRepo.SaveLockedBook(book); err != nil {
			return err
		}

		borrowRecord = &models.BorrowRecord{
			UserID:     userID,
			BookID:     bookID,
			BorrowDate: time.Now(),
			ReturnDate: nil,
			Status:     "borrowed",
		}

		if err := borrowRepo.Create(borrowRecord); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return borrowRecord, nil
}

// ReturnBook allows a user to return a borrowed book
func (s *borrowService) ReturnBook(userID, bookID uint) (*models.BorrowRecord, error) {
	var borrowRecord models.BorrowRecord

	err := s.tx.WithinTransaction(context.Background(), func(ctx context.Context, repos repositories.TxRepositories) error {
		bookRepo := repos.Books()
		if bookRepo == nil {
			bookRepo = s.bookRepo
		}
		borrowRepo := repos.Borrows()
		if borrowRepo == nil {
			borrowRepo = s.borrowRepo
		}

		lockedRecord, err := borrowRepo.LockBorrowedRecordForUpdate(userID, bookID)
		if err != nil {
			return err
		}
		if lockedRecord == nil {
			return ErrRecordNotFound
		}
		borrowRecord = *lockedRecord

		book, err := bookRepo.LockBookForUpdate(bookID)
		if err != nil {
			return err
		}
		if book == nil {
			return ErrBookNotFound
		}

		book.Stock++
		if err := bookRepo.SaveLockedBook(book); err != nil {
			return err
		}

		now := time.Now()
		borrowRecord.ReturnDate = &now
		borrowRecord.Status = "returned"

		if err := borrowRepo.UpdateBorrowRecord(&borrowRecord); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &borrowRecord, nil
}

// GetUserRecords retrieves all borrow records for a specific user
func (s *borrowService) GetUserRecords(userID uint) ([]models.BorrowRecord, error) {
	records, err := s.borrowRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Convert []*models.BorrowRecord to []models.BorrowRecord
	result := make([]models.BorrowRecord, len(records))
	for i, r := range records {
		result[i] = *r
	}

	return result, nil
}

// GetAllRecords retrieves all borrow records
func (s *borrowService) GetAllRecords() ([]models.BorrowRecord, error) {
	records, err := s.borrowRepo.FindAll()
	if err != nil {
		return nil, err
	}

	// Convert []*models.BorrowRecord to []models.BorrowRecord
	result := make([]models.BorrowRecord, len(records))
	for i, r := range records {
		result[i] = *r
	}

	return result, nil
}

// GetRecordsByUserID retrieves all borrow records for a specific user (same as GetUserRecords)
func (s *borrowService) GetRecordsByUserID(userID uint) ([]models.BorrowRecord, error) {
	return s.GetUserRecords(userID)
}
