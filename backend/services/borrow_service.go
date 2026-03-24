package services

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/Dailiduzhou/library_manage_sys/internal/events"
	"github.com/Dailiduzhou/library_manage_sys/internal/ports"
	"github.com/Dailiduzhou/library_manage_sys/models"
	"github.com/Dailiduzhou/library_manage_sys/pkg/logger"
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
	producer   ports.Producer

	borrowedTopic string
	returnedTopic string
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

	s.publishBorrowEvent(events.BorrowEvent{
		EventType:  events.BorrowedEventType,
		BorrowID:   borrowRecord.ID,
		UserID:     borrowRecord.UserID,
		BookID:     borrowRecord.BookID,
		OccurredAt: borrowRecord.BorrowDate,
	}, s.borrowedTopic)

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

	occurredAt := time.Now()
	if borrowRecord.ReturnDate != nil {
		occurredAt = *borrowRecord.ReturnDate
	}
	s.publishBorrowEvent(events.BorrowEvent{
		EventType:  events.ReturnedEventType,
		BorrowID:   borrowRecord.ID,
		UserID:     borrowRecord.UserID,
		BookID:     borrowRecord.BookID,
		OccurredAt: occurredAt,
	}, s.returnedTopic)

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

func (s *borrowService) publishBorrowEvent(evt events.BorrowEvent, topic string) {
	if s.producer == nil || topic == "" {
		return
	}

	payload, err := json.Marshal(evt)
	if err != nil {
		logger.Infof("kafka: failed to marshal borrow event: %v", err)
		return
	}

	key := strconv.FormatUint(uint64(evt.UserID), 10)
	if err := s.producer.SendMessage(topic, key, payload); err != nil {
		logger.Infof("kafka: failed to send borrow event: %v", err)
	}
}
