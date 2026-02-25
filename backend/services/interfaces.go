package services

import (
	"github.com/Dailiduzhou/library_manage_sys/models"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/mock_$GOFILE -package=mocks
type BookService interface {
	CreateBook(title, author, summary, coverPath string, initialStock int) (*models.Book, error)
	GetBooks(title, author, summary string) ([]models.Book, error)
	UpdateBook(id uint, title, author, summary, coverPath string, stock, totalStock int) (*models.Book, error)
	DeleteBook(id uint) error
	GetBookByID(id uint) (*models.Book, error)
}

//go:generate mockgen -source=$GOFILE -destination=mocks/mock_$GOFILE -package=mocks
type UserService interface {
	Register(username, password string) (*models.User, error)
	Login(username, password string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
}

//go:generate mockgen -source=$GOFILE -destination=mocks/mock_$GOFILE -package=mocks
type BorrowService interface {
	BorrowBook(userID, bookID uint) (*models.BorrowRecord, error)
	ReturnBook(userID, bookID uint) (*models.BorrowRecord, error)
	GetUserRecords(userID uint) ([]models.BorrowRecord, error)
	GetAllRecords() ([]models.BorrowRecord, error)
	GetRecordsByUserID(userID uint) ([]models.BorrowRecord, error)
}
