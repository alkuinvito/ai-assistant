package utils

import (
	"context"
	"fmt"
	"strconv"
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

var TraceIDKey = "trace_id"

// GetTraceID returns the trace ID from the context
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if val, ok := ctx.Value("trace_id").(string); ok {
		return val
	}

	return ""
}
