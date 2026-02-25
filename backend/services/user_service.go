package services

import (
	"errors"

	"github.com/Dailiduzhou/library_manage_sys/models"
	"github.com/Dailiduzhou/library_manage_sys/repositories"
	"github.com/Dailiduzhou/library_manage_sys/utils"
)

// Custom errors for UserService
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

// userService implements UserService interface
type userService struct {
	repo repositories.UserRepository
}

// Register creates a new user account
func (s *userService) Register(username, password string) (*models.User, error) {
	existingUser, err := s.repo.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create new user
	newUser := &models.User{
		Username: username,
		Password: hashedPassword,
		Role:     "user",
	}

	if err := s.repo.Create(newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

// Login authenticates a user with username and password
func (s *userService) Login(username, password string) (*models.User, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Verify password
	err = utils.ComparePassword(user.Password, password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (s *userService) GetUserByUsername(username string) (*models.User, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}
