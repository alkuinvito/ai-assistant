package invoices

import "github.com/gofiber/fiber/v3"

type InvoiceRouter interface {
	Handle(router fiber.Router)
}

type invoiceRouter struct {
	handler InvoiceHandler
}

func NewInvoiceRouter(handler InvoiceHandler) InvoiceRouter {
	return &invoiceRouter{handler}
}

func (r *invoiceRouter) Handle(router fiber.Router) {
	router.Route("/invoices", func(group fiber.Router) {
		group.Post("/upload", r.handler.CreateInvoiceUploadUrl)
	})
}
