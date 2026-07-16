package invoices

import (
	"context"

	"github.com/alkuinvito/ai-assistant/internal/uploads"
)

type InvoiceService interface {
	CreateInvoiceUploadUrl(ctx context.Context, req *uploads.DTOCreatePresignedPutUrl) (uploads.DTOPresignedPutUrl, error)
}

type invoiceService struct {
	uploadService uploads.UploadService
}

func NewInvoiceService(uploadService uploads.UploadService) InvoiceService {
	return &invoiceService{uploadService}
}

func (s *invoiceService) CreateInvoiceUploadUrl(ctx context.Context, req *uploads.DTOCreatePresignedPutUrl) (uploads.DTOPresignedPutUrl, error) {
	return s.uploadService.CreatePresignedPutUploadUrl(ctx, "invoices", req.Filename, req.Size, 3600)
}
