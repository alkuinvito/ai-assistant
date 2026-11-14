package users

import (
	"context"
	"errors"

	"github.com/alkuinvito/ai-assistant/pkg/service_error"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService interface {
	CreateUser(tx *gorm.DB, user *User) error
	GetProfile(ctx context.Context, userId string) (*DTOUserProfileResponse, error)
	GetUserByEmail(tx *gorm.DB, email string) (*User, error)
	GetUserById(tx *gorm.DB, id uuid.UUID) (*User, error)
	UpdateProfile(ctx context.Context, userId string, req *DTOUpdateProfileRequest) (*DTOUserProfileResponse, error)
	VerifyUserEmail(tx *gorm.DB, email string) error
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

func (s *userService) GetProfile(ctx context.Context, userId string) (*DTOUserProfileResponse, error) {
	db := s.db.WithContext(ctx)

	uid, err := uuid.Parse(userId)
	if err != nil {
		return nil, service_error.NewBadRequest(ctx, "invalid user id", err)
	}

	user, err := s.repo.getById(db, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service_error.NewNotFound(ctx, "user not found", err)
		}
		return nil, service_error.NewInternalServerError(ctx, "failed to retrieve user", err)
	}

	return toUserProfileResponse(user), nil
}

func (s *userService) GetUserByEmail(tx *gorm.DB, email string) (*User, error) {
	return s.repo.getByEmail(tx, email)
}

func (s *userService) GetUserById(tx *gorm.DB, id uuid.UUID) (*User, error) {
	return s.repo.getById(tx, id)
}

func (s *userService) UpdateProfile(ctx context.Context, userId string, req *DTOUpdateProfileRequest) (*DTOUserProfileResponse, error) {
	db := s.db.WithContext(ctx)

	uid, err := uuid.Parse(userId)
	if err != nil {
		return nil, service_error.NewBadRequest(ctx, "invalid user id", err)
	}

	user, err := s.repo.getById(db, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service_error.NewNotFound(ctx, "user not found", err)
		}
		return nil, service_error.NewInternalServerError(ctx, "failed to retrieve user", err)
	}

	user.Name = req.Name

	if err := s.repo.update(db, user); err != nil {
		return nil, service_error.NewInternalServerError(ctx, "failed to update profile", err)
	}

	return toUserProfileResponse(user), nil
}

func (s *userService) VerifyUserEmail(tx *gorm.DB, email string) error {
	user, err := s.repo.getByEmail(tx, email)
	if err != nil {
		return err
	}

	user.IsVerified = true
	return s.repo.update(tx, user)
}

func toUserProfileResponse(user *User) *DTOUserProfileResponse {
	return &DTOUserProfileResponse{
		Id:         user.Id,
		Name:       user.Name,
		Email:      string(user.Email),
		IsActive:   user.IsActive,
		IsVerified: user.IsVerified,
		CreatedAt:  user.CreatedAt,
	}
}
