package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	BookStatusAvailable = 1 // 可借
	BookStatusBorrowed  = 2 // 已借出
)

type Book struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Summary   string `json:"summary"`
	CoverPath string `json:"cover_path"`

	InitialStock int `json:"initial_stock" gorm:"default:0" binding:"gte=0"`
	Stock        int `json:"stock" gorm:"default:0" binding:"gte=0"`
	TotalStock   int `json:"total_stock" gorm:"defualt:0" binding:"gte=0"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"unique;not null" json:"username"`
	Password  string    `gorm:"not null" json:"-"`
	Role      string    `gorm:"default:'user'" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BorrowRecord struct {
	gorm.Model
	UserID     uint      `json:"user_id"`
	BookID     uint      `json:"book_id"`
	BorrowDate time.Time `json:"borrow_date"`
	ReturnDate time.Time `json:"return_date"`
	Status     string    `json:"status"` // 状态: "borrowed" (借出中), "returned" (已归还)

	// 关联关系 (Preload用)
	User User `json:"user,omitempty"`
	Book Book `json:"book,omitempty"`
}
