package invoices

import (
	"time"

	"github.com/google/uuid"
)

type Invoice struct {
	Id             uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserId         uuid.UUID `gorm:"type:uuid;not null"`
	FileId         uuid.UUID `gorm:"type:uuid;not null"`
	VendorName     string    `gorm:"not null"`
	Description    string
	Currency       string
	Tax            float64
	AdditionalFees float64
	TotalAmount    float64 `gorm:"not null"`
	IssuedAt       time.Time
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`

	Items []InvoiceItem `gorm:"foreignKey:InvoiceId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type InvoiceItem struct {
	Id         uuid.UUID `gorm:"type:uuid;primaryKey"`
	InvoiceId  uuid.UUID `gorm:"type:uuid;not null"`
	Name       string
	Qty        int
	Unit       string
	UnitPrice  float64
	TotalPrice float64
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}
