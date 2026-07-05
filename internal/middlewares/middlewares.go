package middlewares

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alkuinvito/ai-assistant/internal/auth"
	"github.com/alkuinvito/ai-assistant/pkg/service_error"
	"github.com/alkuinvito/ai-assistant/pkg/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/sirupsen/logrus"
)

type Middleware interface {
	ClientId() fiber.Handler
	JWTAuthentication() fiber.Handler
	Logger() fiber.Handler
	RequestId() fiber.Handler
}

type middleware struct {
	logger      *logrus.Logger
	authService auth.AuthService
}

func NewMiddleware(logger *logrus.Logger, authService auth.AuthService) Middleware {
	return &middleware{logger: logger, authService: authService}
}

const HeaderRequestID = "X-Request-ID"

// generateRandomID creates a secure, random hexadecimal string
func generateRandomID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp string if crypto/rand fails
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("req_%s", hex.EncodeToString(bytes))
}

func (m *middleware) ClientId() fiber.Handler {
	return func(c fiber.Ctx) error {
		clientId := c.Get("X-Client-ID")
		if clientId == "" {
			return service_error.NewUnauthorized(c, "invalid client id", nil)
		}

		appClientId := os.Getenv("APP_CLIENT_ID")
		webClientId := os.Getenv("WEB_CLIENT_ID")

		switch clientId {
		case appClientId:
			fiber.StoreInContext(c, "client_type", "app")
		case webClientId:
			fiber.StoreInContext(c, "client_type", "web")
		default:
			return service_error.NewUnauthorized(c, "invalid client id", nil)
		}

		return c.Next()
	}
}

func (m *middleware) JWTAuthentication() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			return service_error.NewUnauthorized(c, "invalid authorization header", nil)
		}

		claims, err := m.authService.VerifyAccessToken(headerParts[1])
		if err != nil {
			return service_error.NewUnauthorized(c, "invalid or expired access token", err)
		}

		userId, err := claims.GetSubject()
		if err != nil {
			return service_error.NewInternalServerError(c, "failed to get user id", err)
		}

		fiber.StoreInContext(c, "user_id", userId)
		return c.Next()
	}
}

func (m *middleware) Logger() fiber.Handler {
	return logger.New(logger.Config{
		LoggerFunc: func(c fiber.Ctx, data *logger.Data, cfg *logger.Config) error {
			requestId := utils.GetRequestID(c)

			// Safely read response metadata
			statusCode := c.Response().StatusCode()
			latency := time.Since(data.Start)

			if statusCode >= 500 {
				m.logger.WithFields(logrus.Fields{
					"req_id":   requestId,
					"duration": latency.Microseconds(),
				}).Errorf("%d | %s %s", statusCode, c.Method(), c.Path())
			} else if statusCode >= 400 {
				m.logger.WithFields(logrus.Fields{
					"req_id":   requestId,
					"duration": latency.Microseconds(),
				}).Warnf("%d | %s %s", statusCode, c.Method(), c.Path())
			} else {
				m.logger.WithFields(logrus.Fields{
					"req_id":   requestId,
					"duration": latency.Microseconds(),
				}).Infof("%d | %s %s", statusCode, c.Method(), c.Path())
			}

			return nil
		},
	})
}

// RequestId initializes the request ID for the request lifecycle
func (m *middleware) RequestId() fiber.Handler {
	return func(c fiber.Ctx) error {
		requestId := c.Get(HeaderRequestID)
		if requestId == "" {
			requestId = generateRandomID()
		}

		fiber.StoreInContext(c, "request_id", requestId)
		c.Set(HeaderRequestID, requestId)

		return c.Next()
	}
}
