package invoices

import "github.com/gofiber/fiber/v3"

// InvoiceGuards carries the authorization handlers applied to invoice routes.
// They are injected by the router assembly so this package does not depend on
// the middleware package (which would create an import cycle through auth).
type InvoiceGuards struct {
	FileWrite    fiber.Handler
	InvoiceRead  fiber.Handler
	InvoiceWrite fiber.Handler
}

type InvoiceRouter interface {
	Handle(router fiber.Router, guards InvoiceGuards)
}

type invoiceRouter struct {
	handler InvoiceHandler
}

func NewInvoiceRouter(handler InvoiceHandler) InvoiceRouter {
	return &invoiceRouter{handler}
}

func (r *invoiceRouter) Handle(router fiber.Router, guards InvoiceGuards) {
	router.Route("/invoices", func(group fiber.Router) {
		group.Post("/upload", guards.FileWrite, r.handler.CreateInvoiceUploadUrl)
		group.Post("/analyze", guards.InvoiceRead, r.handler.AnalyzeInvoice)
		group.Post("/save", guards.InvoiceWrite, r.handler.SaveInvoice)
	})
}
