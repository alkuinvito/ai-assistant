package invoices

import (
	"gorm.io/gorm"
)

type InvoiceRepository interface {
	CreateInvoice(tx *gorm.DB, invoice *Invoice, items []InvoiceItem) error
}

type invoiceRepository struct{}

func NewInvoiceRepository() InvoiceRepository {
	return &invoiceRepository{}
}

func (r *invoiceRepository) CreateInvoice(tx *gorm.DB, invoice *Invoice, items []InvoiceItem) error {
	return tx.Transaction(func(d *gorm.DB) error {
		if err := d.Create(invoice).Error; err != nil {
			return err
		}

		for _, item := range items {
			if err := d.Create(&item).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
