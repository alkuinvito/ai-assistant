package roles

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	Id          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"uniqueIndex;not null"`
	Description string
	Signature   string    `gorm:"not null;default:''"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

type Permission struct {
	Id          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"uniqueIndex;not null"`
	Description string
	Signature   string    `gorm:"not null;default:''"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}

type RolePermission struct {
	RoleId       uuid.UUID `gorm:"type:uuid;primaryKey"`
	PermissionId uuid.UUID `gorm:"type:uuid;primaryKey"`
	Signature    string    `gorm:"not null;default:''"`
}

type UserRole struct {
	UserId    uuid.UUID `gorm:"type:uuid;primaryKey"`
	RoleId    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Signature string    `gorm:"not null;default:''"`
}
