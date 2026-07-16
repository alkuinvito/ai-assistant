package invoices

import (
	"github.com/alkuinvito/ai-assistant/internal/uploads"
	"github.com/alkuinvito/ai-assistant/pkg/request"
	"github.com/alkuinvito/ai-assistant/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type InvoiceHandler interface {
	CreateInvoiceUploadUrl(c fiber.Ctx) error
}

type invoiceHandler struct {
	service InvoiceService
}

func NewInvoiceHandler(service InvoiceService) InvoiceHandler {
	return &invoiceHandler{service}
}

func (h *invoiceHandler) CreateInvoiceUploadUrl(c fiber.Ctx) error {
	req := new(uploads.DTOCreatePresignedPutUrl)
	if err := request.ValidateRequest(c, req); err != nil {
		return err
	}

	result, err := h.service.CreateInvoiceUploadUrl(c, req)
	if err != nil {
		return err
	}

	return response.New(c, result)
}
