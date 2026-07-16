package utils

import (
	"context"
	"fmt"
	"strconv"
	"time"

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

// DetectMimeType returns a MIME type based on the file extension.
func DetectMimeType(filename string) string {
	switch {
	case len(filename) > 4 && filename[len(filename)-4:] == ".pdf":
		return "application/pdf"
	case len(filename) > 4 && filename[len(filename)-4:] == ".jpg":
		return "image/jpeg"
	case len(filename) > 5 && filename[len(filename)-5:] == ".jpeg":
		return "image/jpeg"
	case len(filename) > 4 && filename[len(filename)-4:] == ".png":
		return "image/png"
	default:
		return "application/octet-stream"
	}
}

// ParseDate tries common date layouts.
func ParseDate(s string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"02/01/2006",
		"01/02/2006",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", s)
}
