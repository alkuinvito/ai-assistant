package invoices

import (
	"time"

	"github.com/google/uuid"
)

type DTOAnalyzeInvoiceRequest struct {
	FileId string `json:"file_id" validate:"required,uuid"`
}

type DTOAnalyzeInvoiceResponse struct {
	VendorName     string                   `json:"vendor_name"`
	Description    string                   `json:"description"`
	Currency       string                   `json:"currency"`
	Tax            float64                  `json:"tax"`
	AdditionalFees float64                  `json:"additional_fees"`
	TotalAmount    float64                  `json:"total_amount"`
	IssuedAt       *time.Time               `json:"issued_at,omitempty"`
	Items          []DTOInvoiceItemResponse `json:"items"`
}

type DTOInvoiceItemResponse struct {
	Name       string  `json:"name"`
	Qty        int     `json:"qty"`
	Unit       string  `json:"unit"`
	UnitPrice  float64 `json:"unit_price"`
	TotalPrice float64 `json:"total_price"`
}

type DTOSaveInvoiceRequest struct {
	FileId         string                      `json:"file_id" validate:"required,uuid"`
	VendorName     string                      `json:"vendor_name" validate:"required"`
	Description    string                      `json:"description"`
	Currency       string                      `json:"currency"`
	Tax            float64                     `json:"tax"`
	AdditionalFees float64                     `json:"additional_fees"`
	TotalAmount    float64                     `json:"total_amount" validate:"required"`
	IssuedAt       *time.Time                  `json:"issued_at,omitempty"`
	Items          []DTOSaveInvoiceItemRequest `json:"items"`
}

type DTOSaveInvoiceItemRequest struct {
	Name       string  `json:"name"`
	Qty        int     `json:"qty"`
	Unit       string  `json:"unit"`
	UnitPrice  float64 `json:"unit_price"`
	TotalPrice float64 `json:"total_price"`
}

type DTOSaveInvoiceResponse struct {
	Id      uuid.UUID `json:"id"`
	Message string    `json:"message"`
}
