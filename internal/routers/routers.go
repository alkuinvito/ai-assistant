package routers

import (
	"errors"
	"time"

	"github.com/alkuinvito/ai-assistant/internal/auth"
	"github.com/alkuinvito/ai-assistant/internal/middlewares"
	"github.com/alkuinvito/ai-assistant/internal/users"
	"github.com/alkuinvito/ai-assistant/pkg/apperror"
	"github.com/alkuinvito/ai-assistant/pkg/response"
	"github.com/alkuinvito/ai-assistant/pkg/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/sirupsen/logrus"
)

type Router struct {
	log        *logrus.Logger
	middleware middlewares.Middleware
	userRouter users.UserRouter
	authRouter auth.AuthRouter
}

func NewRouter(
	log *logrus.Logger,
	middleware middlewares.Middleware,
	userRouter users.UserRouter,
	authRouter auth.AuthRouter,
) *Router {
	return &Router{
		log:        log,
		middleware: middleware,
		userRouter: userRouter,
		authRouter: authRouter,
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

			var appError *apperror.AppError
			if errors.As(err, &appError) {
				code = appError.StatusCode()
			} else {
				appError = apperror.NewInternalServerError(c.Context(), "Unknown server error", err)
			}

			return c.Status(code).JSON(response.NewError(appError))
		},
		StructValidator: &structValidator{validate: validator.New()},
	})

	// Middlewares
	app.Use(cors.New())
	app.Use(r.middleware.TraceIDMiddleware())
	app.Use(logger.New(logger.Config{
		LoggerFunc: func(c fiber.Ctx, data *logger.Data, cfg *logger.Config) error {
			traceId := utils.GetTraceID(c)

			// Safely read response metadata
			statusCode := c.Response().StatusCode()
			latency := time.Since(data.Start)

			if statusCode >= 400 {
				r.log.WithFields(logrus.Fields{
					"trace_id": traceId,
					"data": map[string]any{
						"latency": latency.String(),
						"ip":      c.IP(),
					},
				}).Errorf("%d | %s %s", statusCode, c.Method(), c.Path())
			} else {
				r.log.WithFields(logrus.Fields{
					"trace_id": traceId,
					"data": map[string]any{
						"latency": latency.String(),
						"ip":      c.IP(),
					},
				}).Infof("%d | %s %s", statusCode, c.Method(), c.Path())
			}

			return nil
		},
	}))
	app.Use(recoverer.New())

	// Routes
	api := app.Group("/api")
	r.authRouter.Handle(api)
	r.userRouter.Handle(api)

	return app
}
