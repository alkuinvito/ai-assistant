package utils

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

// Stringify converts any value to a string
func Stringify(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(x)
	case fmt.Stringer:
		return x.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// GetRequestID returns the request ID from the context
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if val, ok := ctx.Value("request_id").(string); ok {
		return val
	}

	return ""
}

// GetIp returns the IP address from the fiber context
func GetIp(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if c, ok := ctx.(fiber.Ctx); ok {
		return c.IP()
	}

	return ""
}
