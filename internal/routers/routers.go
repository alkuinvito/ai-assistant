package routers

import (
	"errors"

	"github.com/alkuinvito/ai-assistant/internal/auth"
	"github.com/alkuinvito/ai-assistant/internal/invoices"
	"github.com/alkuinvito/ai-assistant/internal/middlewares"
	"github.com/alkuinvito/ai-assistant/internal/roles"
	"github.com/alkuinvito/ai-assistant/internal/users"
	"github.com/alkuinvito/ai-assistant/pkg/response"
	"github.com/alkuinvito/ai-assistant/pkg/service_error"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/sirupsen/logrus"
)

const maxBodySize = 10 * 1024 * 1024 // 10MB

type Router struct {
	authRouter    auth.AuthRouter
	invoiceRouter invoices.InvoiceRouter
	log           *logrus.Logger
	middlewares   middlewares.Middleware
	roleRouter    roles.RoleRouter
	userRouter    users.UserRouter
}

func NewRouter(
	authRouter auth.AuthRouter,
	invoiceRouter invoices.InvoiceRouter,
	log *logrus.Logger,
	middlewares middlewares.Middleware,
	roleRouter roles.RoleRouter,
	userRouter users.UserRouter,
) *Router {
	return &Router{
		authRouter:    authRouter,
		invoiceRouter: invoiceRouter,
		log:           log,
		middlewares:   middlewares,
		roleRouter:    roleRouter,
		userRouter:    userRouter,
	}
}

type structValidator struct {
	validate *validator.Validate
}

func (v *structValidator) Validate(out any) error {
	return v.validate.Struct(out)
}

func (r *Router) Handle() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			var appError *service_error.ServiceError
			if !errors.As(err, &appError) {
				appError = service_error.NewInternalServerError(c, "unknown server error", err)
			}

			return response.New(c, appError)
		},
		StructValidator: &structValidator{validate: validator.New()},
		BodyLimit:       maxBodySize,
	})

	// Middlewares
	app.Use(cors.New())
	app.Use(r.middlewares.RequestId())
	app.Use(r.middlewares.Logger())
	app.Use(recoverer.New())

	// Health check
	app.Get("/health", func(c fiber.Ctx) error {
		return response.New(c, "ok")
	})

	// Routes
	api := app.Group("/api")
	api.Use(r.middlewares.ClientId())
	r.authRouter.Handle(api)

	// JWT authenticated only
	authGuard := api.Use(r.middlewares.JWTAuthentication())
	r.invoiceRouter.Handle(authGuard, invoices.InvoiceGuards{
		FileWrite:    r.middlewares.RequirePermission("file:write"),
		InvoiceRead:  r.middlewares.RequirePermission("invoice:read"),
		InvoiceWrite: r.middlewares.RequirePermission("invoice:write"),
	})
	r.userRouter.Handle(authGuard, users.UserGuards{
		ProfileWrite: r.middlewares.RequirePermission("profile:write"),
	})

	// Admin only
	adminGuard := api.Group("/admin")
	adminGuard.Use(r.middlewares.JWTAuthentication(), r.middlewares.RequireRole("admin"))
	r.roleRouter.Handle(adminGuard)

	return app
}
