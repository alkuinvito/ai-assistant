package invoices

import (
	"github.com/alkuinvito/ai-assistant/internal/uploads"
	"github.com/alkuinvito/ai-assistant/pkg/request"
	"github.com/alkuinvito/ai-assistant/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type InvoiceHandler interface {
	AnalyzeInvoice(c fiber.Ctx) error
	CreateInvoiceUploadUrl(c fiber.Ctx) error
	SaveInvoice(c fiber.Ctx) error
}

type invoiceHandler struct {
	service InvoiceService
}

func NewInvoiceHandler(service InvoiceService) InvoiceHandler {
	return &invoiceHandler{service}
}

func (h *invoiceHandler) AnalyzeInvoice(c fiber.Ctx) error {
	req := new(DTOAnalyzeInvoiceRequest)
	if err := request.ValidateRequest(c, req); err != nil {
		return err
	}

	userId, ok := c.Locals("user_id").(string)
	if !ok || userId == "" {
		return fiber.ErrUnauthorized
	}

	result, err := h.service.AnalyzeInvoice(c, userId, req)
	if err != nil {
		return err
	}

	return response.New(c, result)
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

func (h *invoiceHandler) SaveInvoice(c fiber.Ctx) error {
	req := new(DTOSaveInvoiceRequest)
	if err := request.ValidateRequest(c, req); err != nil {
		return err
	}

	userId, ok := c.Locals("user_id").(string)
	if !ok || userId == "" {
		return fiber.ErrUnauthorized
	}

	result, err := h.service.SaveInvoice(c, userId, req)
	if err != nil {
		return err
	}

	return response.New(c, result)
}
