package repositories

import (
	"errors"

	"github.com/Dailiduzhou/library_manage_sys/models"
	"gorm.io/gorm"
)

// GormUserRepository implements UserRepository using GORM
type GormUserRepository struct{}



// Create creates a new user in the database
func (r *GormUserRepository) Create(db *gorm.DB, user *models.User) error {
	return db.Create(user).Error
}

// GetByID retrieves a user by ID
func (r *GormUserRepository) GetByID(db *gorm.DB, id uint) (*models.User, error) {
	var user models.User
	err := db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *GormUserRepository) GetByUsername(db *gorm.DB, username string) (*models.User, error) {
	var user models.User
	err := db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindAll retrieves all users
func (r *GormUserRepository) FindAll(db *gorm.DB) ([]*models.User, error) {
	var users []*models.User
	err := db.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}
