package users

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	create(db *gorm.DB, user *User) error
	getById(db *gorm.DB, id uuid.UUID) (*User, error)
	getByEmail(db *gorm.DB, email string) (*User, error)
	update(db *gorm.DB, user *User) error
}

type userRepository struct{}

func NewUserRepository() UserRepository {
	return &userRepository{}
}

func (r *userRepository) create(db *gorm.DB, user *User) error {
	err := db.Create(user).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) getById(db *gorm.DB, id uuid.UUID) (*User, error) {
	var result *User

	err := db.First(&result, id).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *userRepository) getByEmail(db *gorm.DB, email string) (*User, error) {
	var result *User

	err := db.First(&result, "email = ?", email).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *userRepository) update(db *gorm.DB, user *User) error {
	return db.Save(user).Error
}
