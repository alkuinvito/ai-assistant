package users

import (
	"time"

	"github.com/alkuinvito/ai-assistant/pkg/database"
	"github.com/google/uuid"
)

type User struct {
	Id         uuid.UUID          `gorm:"type:uuid;primaryKey"`
	Email      database.Encrypted `gorm:"type:bytes"`
	EmailHash  string             `gorm:"uniqueIndex;not null"`
	Name       string             `gorm:"not null"`
	Password   string             `gorm:"not null"`
	IsActive   bool               `gorm:"default:true"`
	IsVerified bool               `gorm:"default:false"`
	CreatedAt  time.Time          `gorm:"autoCreateTime"`
	UpdatedAt  time.Time          `gorm:"autoUpdateTime"`
}
