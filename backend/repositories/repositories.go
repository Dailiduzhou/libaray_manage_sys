package repositories

import "gorm.io/gorm"

// NewGormBookRepository creates a new BookRepository using GORM
func NewGormBookRepository(db *gorm.DB) BookRepository {
	return &gormBookRepository{db: db}
}

// NewGormUserRepository creates a new UserRepository using GORM
func NewGormUserRepository() UserRepository {
	return &GormUserRepository{}
}

// NewGormBorrowRepository creates a new BorrowRepository using GORM
func NewGormBorrowRepository(db *gorm.DB) BorrowRepository {
	return &gormBorrowRepository{db: db}
}
