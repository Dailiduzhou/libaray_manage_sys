package services

import (
	"errors"
	"testing"

	"github.com/Dailiduzhou/library_manage_sys/models"
)

type stubBookRepository struct {
	createBookFn         func(book *models.Book) error
	getBookByIDFn        func(id uint) (*models.Book, error)
	listBooksFn          func() ([]*models.Book, error)
	findBooksByTitleFn   func(title string) ([]*models.Book, error)
	findBooksByAuthorFn  func(author string) ([]*models.Book, error)
	updateBookFn         func(book *models.Book) error
	updateStockWithTotal func(id uint, newStock, newTotalStock int) (*models.Book, error)
	deleteBookFn         func(id uint) error
	lockBookForUpdateFn  func(id uint) (*models.Book, error)
	saveLockedBookFn     func(book *models.Book) error
}

func (s *stubBookRepository) CreateBook(book *models.Book) error {
	if s.createBookFn != nil {
		return s.createBookFn(book)
	}
	return nil
}

func (s *stubBookRepository) GetBookByID(id uint) (*models.Book, error) {
	if s.getBookByIDFn != nil {
		return s.getBookByIDFn(id)
	}
	return nil, nil
}

func (s *stubBookRepository) ListBooks() ([]*models.Book, error) {
	if s.listBooksFn != nil {
		return s.listBooksFn()
	}
	return []*models.Book{}, nil
}

func (s *stubBookRepository) FindBooksByTitle(title string) ([]*models.Book, error) {
	if s.findBooksByTitleFn != nil {
		return s.findBooksByTitleFn(title)
	}
	return []*models.Book{}, nil
}

func (s *stubBookRepository) FindBooksByAuthor(author string) ([]*models.Book, error) {
	if s.findBooksByAuthorFn != nil {
		return s.findBooksByAuthorFn(author)
	}
	return []*models.Book{}, nil
}

func (s *stubBookRepository) UpdateBook(book *models.Book) error {
	if s.updateBookFn != nil {
		return s.updateBookFn(book)
	}
	return nil
}

func (s *stubBookRepository) UpdateBookStockWithTotal(id uint, newStock, newTotalStock int) (*models.Book, error) {
	if s.updateStockWithTotal != nil {
		return s.updateStockWithTotal(id, newStock, newTotalStock)
	}
	return nil, nil
}

func (s *stubBookRepository) DeleteBook(id uint) error {
	if s.deleteBookFn != nil {
		return s.deleteBookFn(id)
	}
	return nil
}

func (s *stubBookRepository) LockBookForUpdate(id uint) (*models.Book, error) {
	if s.lockBookForUpdateFn != nil {
		return s.lockBookForUpdateFn(id)
	}
	return nil, nil
}

func (s *stubBookRepository) SaveLockedBook(book *models.Book) error {
	if s.saveLockedBookFn != nil {
		return s.saveLockedBookFn(book)
	}
	return nil
}

func TestBookServiceCreateBookDuplicate(t *testing.T) {
	repo := &stubBookRepository{
		listBooksFn: func() ([]*models.Book, error) {
			return []*models.Book{
				{ID: 1, Title: "Go Programming", Author: "John Doe"},
			}, nil
		},
	}
	svc := &bookService{
		repo: repo,
	}

	book, err := svc.CreateBook("go programming", "john doe", "", "", 10)
	if !errors.Is(err, ErrBookAlreadyExists) {
		t.Fatalf("expected ErrBookAlreadyExists, got err=%v", err)
	}
	if book != nil {
		t.Fatalf("expected nil book on duplicate, got %+v", book)
	}
}

func TestBookServiceGetBooksFilterAndSort(t *testing.T) {
	repo := &stubBookRepository{
		listBooksFn: func() ([]*models.Book, error) {
			return []*models.Book{
				{ID: 1, Title: "Go Basics", Author: "John", Summary: "guide one"},
				{ID: 3, Title: "Go Advanced", Author: "John", Summary: "guide two"},
				{ID: 2, Title: "Rust Intro", Author: "John", Summary: "guide three"},
			}, nil
		},
	}
	svc := &bookService{
		repo: repo,
	}

	books, err := svc.GetBooks("go", "john", "guide")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(books) != 2 {
		t.Fatalf("expected 2 books, got %d", len(books))
	}
	if books[0].ID != 3 || books[1].ID != 1 {
		t.Fatalf("expected books sorted by id desc [3,1], got [%d,%d]", books[0].ID, books[1].ID)
	}
}

func TestBookServiceUpdateBookNotFound(t *testing.T) {
	repo := &stubBookRepository{
		getBookByIDFn: func(id uint) (*models.Book, error) {
			return nil, nil
		},
	}
	svc := &bookService{
		repo: repo,
	}

	book, err := svc.UpdateBook(99, "new", "", "", "", -1, 0)
	if !errors.Is(err, ErrBookNotFound) {
		t.Fatalf("expected ErrBookNotFound, got err=%v", err)
	}
	if book != nil {
		t.Fatalf("expected nil updated book, got %+v", book)
	}
}

func TestBookServiceUpdateBookStockInvalid(t *testing.T) {
	saved := false
	repo := &stubBookRepository{
		getBookByIDFn: func(id uint) (*models.Book, error) {
			return &models.Book{ID: id, Stock: 2, TotalStock: 2}, nil
		},
		updateBookFn: func(book *models.Book) error {
			saved = true
			return nil
		},
	}
	svc := &bookService{
		repo: repo,
	}

	book, err := svc.UpdateBook(1, "", "", "", "", 5, 4)
	if !errors.Is(err, ErrStockInvalid) {
		t.Fatalf("expected ErrStockInvalid, got err=%v", err)
	}
	if book != nil {
		t.Fatalf("expected nil updated book, got %+v", book)
	}
	if saved {
		t.Fatal("expected UpdateBook not to be called when stock is invalid")
	}
}

func TestBookServiceDeleteBookBorrowed(t *testing.T) {
	repo := &stubBookRepository{
		getBookByIDFn: func(id uint) (*models.Book, error) {
			return &models.Book{ID: id, Stock: 1, TotalStock: 2}, nil
		},
	}
	svc := &bookService{
		repo: repo,
	}

	err := svc.DeleteBook(1)
	if !errors.Is(err, ErrBookBorrowed) {
		t.Fatalf("expected ErrBookBorrowed, got err=%v", err)
	}
}
