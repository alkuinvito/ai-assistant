package uploads

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	Id        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserId    uuid.UUID `gorm:"type:uuid;not null"`
	Filename  string    `gorm:"not null"`
	Path      string    `gorm:"not null"`
	Size      float64   `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`

	// Invoice invoices.Invoice `gorm:"foreignKey:FileId;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

type UserQuota struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	MaxBytes  int64     `gorm:"not null;default:536870912"` // 512 MB
	UsedBytes int64     `gorm:"not null;default:0"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
