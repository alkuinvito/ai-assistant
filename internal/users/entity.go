package users

import (
	"time"

	"github.com/google/uuid"
)

type DTOUpdateProfileRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

type DTOUserProfileResponse struct {
	Id         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	IsActive   bool      `json:"is_active"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
}
