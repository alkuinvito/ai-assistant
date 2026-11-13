package invoices

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/alkuinvito/ai-assistant/internal/uploads"
	"github.com/alkuinvito/ai-assistant/pkg/service_error"
	"github.com/alkuinvito/ai-assistant/pkg/storage"
	"github.com/alkuinvito/ai-assistant/pkg/utils"
	"github.com/google/uuid"
	"github.com/zendev-sh/goai"
	"github.com/zendev-sh/goai/provider"
	"github.com/zendev-sh/goai/provider/openrouter"
	"gorm.io/gorm"
)

type InvoiceService interface {
	AnalyzeInvoice(ctx context.Context, userId string, req *DTOAnalyzeInvoiceRequest) (*DTOAnalyzeInvoiceResponse, error)
	CreateInvoiceUploadUrl(ctx context.Context, userId string, req *uploads.DTOCreatePresignedPutUrl) (uploads.DTOPresignedPutUrl, error)
	SaveInvoice(ctx context.Context, userId string, req *DTOSaveInvoiceRequest) (*DTOSaveInvoiceResponse, error)
}

type invoiceService struct {
	db            *gorm.DB
	repo          InvoiceRepository
	uploadService uploads.UploadService
	storage       storage.Storage
}

func NewInvoiceService(db *gorm.DB, repo InvoiceRepository, uploadService uploads.UploadService, s storage.Storage) InvoiceService {
	return &invoiceService{db, repo, uploadService, s}
}

func (s *invoiceService) AnalyzeInvoice(ctx context.Context, userId string, req *DTOAnalyzeInvoiceRequest) (*DTOAnalyzeInvoiceResponse, error) {
	uid, err := uuid.Parse(userId)
	if err != nil {
		return nil, service_error.NewUnauthorized(ctx, "invalid user id", err)
	}

	fileId, err := uuid.Parse(req.FileId)
	if err != nil {
		return nil, service_error.NewBadRequest(ctx, "invalid file id", err)
	}

	db := s.db.WithContext(ctx)

	file, err := s.uploadService.GetFileByIDAndUser(ctx, db, fileId, uid)
	if err != nil {
		return nil, err
	}

	// Download file from R2
	fileBytes, err := s.storage.GetObject(ctx, "invoices", file.Path)
	if err != nil {
		return nil, service_error.NewInternalServerError(ctx, "failed to download file from storage", err)
	}

	// Encode to base64 data URL for the AI model
	encoded := base64.StdEncoding.EncodeToString(fileBytes)
	mimeType := utils.DetectMimeType(file.Filename)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)

	// Call goai to extract structured invoice data
	model := openrouter.Chat("google/gemini-2.5-pro")

	userMessage := provider.Message{
		Role: provider.RoleUser,
		Content: []provider.Part{
			{
				Type: provider.PartText,
				Text: "Extract all financial metadata fields from this invoice document.",
			},
			{
				Type: provider.PartImage,
				URL:  dataURL,
			},
		},
	}

	result, err := goai.GenerateObject[InvoiceExtraction](
		ctx,
		model,
		goai.WithSystem("You are a precise financial data ingestion parser. Extract every field you can find. Return numbers without formatting (no commas, no currency symbols). Use ISO 8601 for dates."),
		goai.WithMessages(userMessage),
		goai.WithMaxOutputTokens(2000),
		goai.WithTemperature(0.0),
	)
	if err != nil {
		return nil, service_error.NewInternalServerError(ctx, "failed to analyze invoice with AI", err)
	}

	extraction := result.Object

	// Parse the issued_at date
	var issuedAt *time.Time
	if extraction.IssuedAt != "" {
		parsed, err := utils.ParseDate(extraction.IssuedAt)
		if err == nil {
			issuedAt = &parsed
		}
	}

	// Build item response — no DB writes, return to frontend for review
	itemResponses := make([]DTOInvoiceItemResponse, len(extraction.Items))
	for i, item := range extraction.Items {
		itemResponses[i] = DTOInvoiceItemResponse{
			Name:       item.Name,
			Qty:        item.Qty,
			Unit:       item.Unit,
			UnitPrice:  item.UnitPrice,
			TotalPrice: item.TotalPrice,
		}
	}

	return &DTOAnalyzeInvoiceResponse{
		VendorName:     extraction.VendorName,
		Description:    extraction.Description,
		Currency:       extraction.Currency,
		Tax:            extraction.Tax,
		AdditionalFees: extraction.AdditionalFees,
		TotalAmount:    extraction.TotalAmount,
		IssuedAt:       issuedAt,
		Items:          itemResponses,
	}, nil
}

func (s *invoiceService) CreateInvoiceUploadUrl(ctx context.Context, userId string, req *uploads.DTOCreatePresignedPutUrl) (uploads.DTOPresignedPutUrl, error) {
	uid, err := uuid.Parse(userId)
	if err != nil {
		return uploads.DTOPresignedPutUrl{}, service_error.NewUnauthorized(ctx, "invalid user id", err)
	}

	return s.uploadService.CreatePresignedPutUploadUrl(ctx, "invoices", req.Filename, req.Size, uid, 3600)
}

func (s *invoiceService) SaveInvoice(ctx context.Context, userId string, req *DTOSaveInvoiceRequest) (*DTOSaveInvoiceResponse, error) {
	uid, err := uuid.Parse(userId)
	if err != nil {
		return nil, service_error.NewUnauthorized(ctx, "invalid user id", err)
	}

	fileId, err := uuid.Parse(req.FileId)
	if err != nil {
		return nil, service_error.NewBadRequest(ctx, "invalid file id", err)
	}

	db := s.db.WithContext(ctx)

	// Verify the file exists and belongs to the user
	file, err := s.uploadService.GetFileByIDAndUser(ctx, db, fileId, uid)
	if err != nil {
		return nil, err
	}

	// Build invoice model from verified data
	invoiceId := uuid.New()
	invoice := &Invoice{
		Id:             invoiceId,
		UserId:         uid,
		FileId:         fileId,
		VendorName:     req.VendorName,
		Description:    req.Description,
		Currency:       req.Currency,
		Tax:            req.Tax,
		AdditionalFees: req.AdditionalFees,
		TotalAmount:    req.TotalAmount,
		IssuedAt: func() time.Time {
			if req.IssuedAt != nil {
				return *req.IssuedAt
			}
			return time.Time{}
		}(),
	}

	items := make([]InvoiceItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = InvoiceItem{
			Id:         uuid.New(),
			InvoiceId:  invoiceId,
			Name:       item.Name,
			Qty:        item.Qty,
			Unit:       item.Unit,
			UnitPrice:  item.UnitPrice,
			TotalPrice: item.TotalPrice,
		}
	}

	if err := s.repo.CreateInvoice(db, invoice, items); err != nil {
		return nil, service_error.NewInternalServerError(ctx, "failed to save invoice", err)
	}

	// Increment the user's upload quota with the confirmed file size
	if err := s.uploadService.IncrementUsedBytes(ctx, uid, int64(file.Size)); err != nil {
		// Invoice was saved successfully, but quota increment failed — log for auditing
		_ = service_error.NewInternalServerError(ctx, "failed to increment upload quota", err)
	}

	return &DTOSaveInvoiceResponse{
		Id:      invoiceId,
		Message: "invoice saved successfully",
	}, nil
}
