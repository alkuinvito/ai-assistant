package middlewares

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
)

type Middleware interface {
	TraceIDMiddleware() fiber.Handler
}

type middleware struct{}

func NewMiddleware() Middleware {
	return &middleware{}
}

const HeaderXTraceID = "X-Trace-ID"

// generateRandomID creates a secure, random hexadecimal string
func generateRandomID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp string if crypto/rand fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// TraceIDMiddleware initializes the trace ID for the request lifecycle
func (m *middleware) TraceIDMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		traceID := c.Get(HeaderXTraceID)
		if traceID == "" {
			traceID = generateRandomID()
		}

		fiber.StoreInContext(c, "trace_id", traceID)
		c.Set(HeaderXTraceID, traceID)

		return c.Next()
	}
}
