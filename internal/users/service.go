package users

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService interface {
	CreateUser(tx *gorm.DB, user *User) error
	GetUserByEmail(tx *gorm.DB, email string) (*User, error)
	GetUserById(tx *gorm.DB, id uuid.UUID) (*User, error)
}

type userService struct {
	db   *gorm.DB
	repo UserRepository
}

func NewUserService(db *gorm.DB, repo UserRepository) UserService {
	return &userService{db: db, repo: repo}
}

func (s *userService) CreateUser(tx *gorm.DB, user *User) error {
	return s.repo.create(tx, user)
}

func (s *userService) GetUserById(tx *gorm.DB, id uuid.UUID) (*User, error) {
	return s.repo.getById(tx, id)
}

func (s *userService) GetUserByEmail(tx *gorm.DB, email string) (*User, error) {
	return s.repo.getByEmail(tx, email)
}
