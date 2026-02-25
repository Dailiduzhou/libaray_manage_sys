package repositories

import (
	"errors"

	"github.com/Dailiduzhou/library_manage_sys/models"
	"gorm.io/gorm"
)

// gormUserRepository implements UserRepository using GORM.
type gormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository creates a new UserRepository using GORM.
func NewGormUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}

// Create creates a new user in the database.
func (r *gormUserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// GetByID retrieves a user by ID.
func (r *gormUserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername retrieves a user by username.
func (r *gormUserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindAll retrieves all users.
func (r *gormUserRepository) FindAll() ([]*models.User, error) {
	var users []*models.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
