package middlewares

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/alkuinvito/ai-assistant/pkg/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/sirupsen/logrus"
)

type Middleware interface {
	Logger() fiber.Handler
	RequestId() fiber.Handler
}

type middleware struct {
	logger *logrus.Logger
}

func NewMiddleware(logger *logrus.Logger) Middleware {
	return &middleware{logger: logger}
}

const HeaderRequestID = "X-Request-ID"

// generateRandomID creates a secure, random hexadecimal string
func generateRandomID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp string if crypto/rand fails
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("req_%s", hex.EncodeToString(bytes))
}

func (m *middleware) Logger() fiber.Handler {
	return logger.New(logger.Config{
		LoggerFunc: func(c fiber.Ctx, data *logger.Data, cfg *logger.Config) error {
			requestId := utils.GetRequestID(c)

			// Safely read response metadata
			statusCode := c.Response().StatusCode()
			latency := time.Since(data.Start)

			if statusCode >= 400 {
				m.logger.WithFields(logrus.Fields{
					"req_id": requestId,
					"data": map[string]any{
						"latency": latency.String(),
						"ip":      c.IP(),
					},
				}).Errorf("%d | %s %s", statusCode, c.Method(), c.Path())
			} else {
				m.logger.WithFields(logrus.Fields{
					"req_id": requestId,
					"data": map[string]any{
						"latency": latency.String(),
						"ip":      c.IP(),
					},
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
