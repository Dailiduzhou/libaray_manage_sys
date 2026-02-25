package services

import (
	"errors"
	"sort"
	"strings"

	"github.com/Dailiduzhou/library_manage_sys/models"
	"github.com/Dailiduzhou/library_manage_sys/repositories"
)

// Custom errors for BookService
var (
	ErrBookNotFound      = errors.New("book not found")
	ErrBookAlreadyExists = errors.New("book already exists")
	ErrStockInvalid      = errors.New("stock cannot be greater than total stock")
	ErrBookBorrowed      = errors.New("book is still borrowed")
)

// bookService implements BookService interface
type bookService struct {
	repo repositories.BookRepository
}

// CreateBook creates a new book with duplicate checking
func (s *bookService) CreateBook(title, author, summary, coverPath string, initialStock int) (*models.Book, error) {
	books, err := s.repo.ListBooks()
	if err != nil {
		return nil, err
	}

	// Preserve existing semantics: title + author must be unique.
	for _, existingBook := range books {
		if strings.EqualFold(existingBook.Title, title) && strings.EqualFold(existingBook.Author, author) {
			return nil, ErrBookAlreadyExists
		}
	}

	finalSummary := summary
	if finalSummary == "" {
		finalSummary = models.DefaultSummary
	}

	finalCoverPath := coverPath
	if finalCoverPath == "" {
		finalCoverPath = models.DefaultCoverPath
	}

	// Create new book
	newBook := &models.Book{
		Title:        title,
		Author:       author,
		Summary:      finalSummary,
		CoverPath:    finalCoverPath,
		InitialStock: initialStock,
		Stock:        initialStock,
		TotalStock:   initialStock,
	}

	if err := s.repo.CreateBook(newBook); err != nil {
		return nil, err
	}

	return newBook, nil
}

// GetBooks retrieves books by filtering criteria
func (s *bookService) GetBooks(title, author, summary string) ([]models.Book, error) {
	books, err := s.repo.ListBooks()
	if err != nil {
		return nil, err
	}

	titleLower := strings.ToLower(title)
	authorLower := strings.ToLower(author)
	summaryLower := strings.ToLower(summary)

	result := make([]models.Book, 0, len(books))
	for _, b := range books {
		if title != "" && !strings.Contains(strings.ToLower(b.Title), titleLower) {
			continue
		}
		if author != "" && !strings.Contains(strings.ToLower(b.Author), authorLower) {
			continue
		}
		if summary != "" && !strings.Contains(strings.ToLower(b.Summary), summaryLower) {
			continue
		}
		result = append(result, *b)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID > result[j].ID
	})

	return result, nil
}

// UpdateBook updates an existing book with validation
func (s *bookService) UpdateBook(id uint, title, author, summary, coverPath string, stock, totalStock int) (*models.Book, error) {
	book, err := s.repo.GetBookByID(id)
	if err != nil {
		return nil, err
	}
	if book == nil {
		return nil, ErrBookNotFound
	}

	if title != "" {
		book.Title = title
	}
	if author != "" {
		book.Author = author
	}
	if summary != "" {
		book.Summary = summary
	}
	if coverPath != "" {
		book.CoverPath = coverPath
	}

	if stock >= 0 || totalStock > 0 {
		newStock := book.Stock
		newTotalStock := book.TotalStock
		if stock >= 0 {
			newStock = stock
		}
		if totalStock > 0 {
			newTotalStock = totalStock
		}
		if newStock > newTotalStock {
			return nil, ErrStockInvalid
		}
		book.Stock = newStock
		book.TotalStock = newTotalStock
	}

	if err := s.repo.UpdateBook(book); err != nil {
		return nil, err
	}

	return book, nil
}

// DeleteBook deletes a book with stock validation
func (s *bookService) DeleteBook(id uint) error {
	book, err := s.repo.GetBookByID(id)
	if err != nil {
		return err
	}
	if book == nil {
		return ErrBookNotFound
	}

	if book.Stock != book.TotalStock {
		return ErrBookBorrowed
	}

	return s.repo.DeleteBook(id)
}

// GetBookByID retrieves a book by its ID
func (s *bookService) GetBookByID(id uint) (*models.Book, error) {
	book, err := s.repo.GetBookByID(id)
	if err != nil {
		return nil, err
	}
	if book == nil {
		return nil, ErrBookNotFound
	}
	return book, nil
}
