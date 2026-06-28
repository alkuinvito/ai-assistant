package routers

import (
	"errors"

	"github.com/alkuinvito/ai-assistant/internal/auth"
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

type Router struct {
	log         *logrus.Logger
	middlewares middlewares.Middleware
	userRouter  users.UserRouter
	authRouter  auth.AuthRouter
}

func NewRouter(
	log *logrus.Logger,
	middlewares middlewares.Middleware,
	userRouter users.UserRouter,
	authRouter auth.AuthRouter,
) *Router {
	return &Router{
		log:         log,
		middlewares: middlewares,
		userRouter:  userRouter,
		authRouter:  authRouter,
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
			code := fiber.StatusInternalServerError

			var appError *service_error.ServiceError
			if errors.As(err, &appError) {
				code = appError.StatusCode()
			} else {
				appError = service_error.NewInternalServerError(c.Context(), "unknown server error", err)
			}

			return c.Status(code).JSON(response.NewError(appError))
		},
		StructValidator: &structValidator{validate: validator.New()},
	})

	// Middlewares
	app.Use(cors.New())
	app.Use(r.middlewares.RequestId())
	app.Use(r.middlewares.Logger())
	app.Use(recoverer.New())

	// Health check
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(response.New("ok"))
	})

	// Routes
	api := app.Group("/api")
	r.authRouter.Handle(api)
	r.userRouter.Handle(api)

	return app
}
