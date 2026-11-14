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
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Middleware interface {
	ClientId() fiber.Handler
	JWTAuthentication() fiber.Handler
	Logger() fiber.Handler
	RequestId() fiber.Handler
	RequirePermission(permission string) fiber.Handler
	RequireRole(role string) fiber.Handler
}

type middleware struct {
	authService auth.AuthService
	logger      *logrus.Logger
}

func NewMiddleware(logger *logrus.Logger, authService auth.AuthService) Middleware {
	return &middleware{authService: authService, logger: logger}
}

const HeaderRequestID = "X-Request-ID"

// claimStringSlice reads a string-list claim from a JWT.
func claimStringSlice(claims jwt.Claims, key string) []string {
	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		return nil
	}

	raw, ok := mapClaims[key]
	if !ok {
		return nil
	}

	switch values := raw.(type) {
	case []string:
		return values
	case []any:
		result := make([]string, 0, len(values))
		for _, value := range values {
			if str, ok := value.(string); ok {
				result = append(result, str)
			}
		}
		return result
	default:
		return nil
	}
}

// generateRandomID creates a secure, random hexadecimal string
func generateRandomID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp string if crypto/rand fails
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("req_%s", hex.EncodeToString(bytes))
}

// hasLocalValue reports whether the string-list stored under key contains value.
func hasLocalValue(c fiber.Ctx, key, value string) bool {
	values, ok := c.Locals(key).([]string)
	if !ok {
		return false
	}

	for _, item := range values {
		if item == value {
			return true
		}
	}

	return false
}

// userIDFromContext extracts and parses the authenticated user id from the context.
func userIDFromContext(c fiber.Ctx) (uuid.UUID, error) {
	raw, ok := c.Locals("user_id").(string)
	if !ok || raw == "" {
		return uuid.Nil, service_error.NewUnauthorized(c, "unauthorized", nil)
	}

	uid, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, service_error.NewUnauthorized(c, "invalid user id", err)
	}

	return uid, nil
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
		fiber.StoreInContext(c, "roles", claimStringSlice(claims, "roles"))
		fiber.StoreInContext(c, "permissions", claimStringSlice(claims, "permissions"))

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

// RequirePermission rejects the request unless the access token grants the permission.
func (m *middleware) RequirePermission(permission string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if _, err := userIDFromContext(c); err != nil {
			return err
		}

		if !hasLocalValue(c, "permissions", permission) {
			return service_error.NewForbidden(c, "insufficient permissions", nil)
		}

		return c.Next()
	}
}

// RequireRole rejects the request unless the access token grants the role.
func (m *middleware) RequireRole(role string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if _, err := userIDFromContext(c); err != nil {
			return err
		}

		if !hasLocalValue(c, "roles", role) {
			return service_error.NewForbidden(c, "insufficient role", nil)
		}

		return c.Next()
	}
}
