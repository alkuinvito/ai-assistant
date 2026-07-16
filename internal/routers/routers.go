package routers

import (
	"errors"

	"github.com/alkuinvito/ai-assistant/internal/auth"
	"github.com/alkuinvito/ai-assistant/internal/invoices"
	"github.com/alkuinvito/ai-assistant/internal/middlewares"
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
	log           *logrus.Logger
	middlewares   middlewares.Middleware
	authRouter    auth.AuthRouter
	invoiceRouter invoices.InvoiceRouter
	userRouter    users.UserRouter
}

func NewRouter(
	log *logrus.Logger,
	middlewares middlewares.Middleware,
	authRouter auth.AuthRouter,
	invoiceRouter invoices.InvoiceRouter,
	userRouter users.UserRouter,
) *Router {
	return &Router{
		log:           log,
		middlewares:   middlewares,
		authRouter:    authRouter,
		invoiceRouter: invoiceRouter,
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
	r.userRouter.Handle(authGuard)
	r.invoiceRouter.Handle(authGuard)

	return app
}
