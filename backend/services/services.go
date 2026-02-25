package services

import "github.com/Dailiduzhou/library_manage_sys/repositories"

// This file re-exports all service constructors for convenient access.
// Instead of importing individual service files, users can import this package
// and access all services from a single location.

// NewBookService creates a new BookService instance
func NewBookService(bookRepo repositories.BookRepository) *bookService {
	return &bookService{
		repo: bookRepo,
	}
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{
		repo: userRepo,
	}
}

// NewBorrowService creates a new BorrowService instance
func NewBorrowService(
	borrowRepo repositories.BorrowRepository,
	bookRepo repositories.BookRepository,
	tx repositories.Transactor,
) BorrowService {
	return &borrowService{
		borrowRepo: borrowRepo,
		bookRepo:   bookRepo,
		tx:         tx,
	}
}
