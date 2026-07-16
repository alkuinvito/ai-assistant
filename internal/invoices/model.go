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

type InvoiceExtraction struct {
	VendorName     string                  `json:"vendor_name" jsonschema:"description=Name of the vendor or supplier issuing the invoice"`
	Description    string                  `json:"description" jsonschema:"description=Invoice description, notes, or memo"`
	Currency       string                  `json:"currency" jsonschema:"description=3-letter currency code (USD, EUR, IDR, etc.)"`
	Tax            float64                 `json:"tax" jsonschema:"description=Total tax amount"`
	AdditionalFees float64                 `json:"additional_fees" jsonschema:"description=Any additional fees or charges on the invoice"`
	TotalAmount    float64                 `json:"total_amount" jsonschema:"description=Total invoice amount including tax and fees"`
	IssuedAt       string                  `json:"issued_at" jsonschema:"description=Invoice issue date in ISO 8601 format (e.g. 2024-03-15 or 2024-03-15T00:00:00Z)"`
	Items          []InvoiceItemExtraction `json:"items" jsonschema:"description=Line items on the invoice"`
}

type InvoiceItemExtraction struct {
	Name       string  `json:"name" jsonschema:"description=Item or service name"`
	Qty        int     `json:"qty" jsonschema:"description=Quantity purchased"`
	Unit       string  `json:"unit" jsonschema:"description=Unit of measurement (pcs, hours, kg, etc.)"`
	UnitPrice  float64 `json:"unit_price" jsonschema:"description=Price per single unit"`
	TotalPrice float64 `json:"total_price" jsonschema:"description=Total price for this line item"`
}
