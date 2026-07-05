package response

import (
	"errors"

	"github.com/alkuinvito/ai-assistant/pkg/service_error"
	"github.com/gofiber/fiber/v3"
)

type ErrorResponse struct {
	Message string `json:"message"`
	// Details any    `json:"details"`
}

type Meta struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

type Response struct {
	ReqId  string         `json:"req_id,omitempty"`
	Data   any            `json:"data"`
	Error  *ErrorResponse `json:"error,omitempty"`
	Meta   *Meta          `json:"meta,omitempty"`
	status int            `json:"-"`
}

func New(c fiber.Ctx, data any, meta ...*Meta) error {
	res := &Response{
		status: fiber.StatusOK,
	}

	for _, m := range meta {
		res.Meta = m
	}

	if err, ok := data.(error); ok {
		// Include the request ID in the response for better traceability
		res.ReqId = c.Locals("request_id").(string)

		// Parse custom service errors and set the appropriate HTTP status code
		var appError *service_error.ServiceError
		if errors.As(err, &appError) {
			res.status = appError.StatusCode()
			res.Error = &ErrorResponse{
				Message: appError.Error(),
			}
		} else {
			res.status = fiber.StatusInternalServerError
			res.Error = &ErrorResponse{
				Message: err.Error(),
			}
		}

		return res.toJSON(c)
	}

	res.Data = data
	return res.toJSON(c)
}

func (r *Response) toJSON(c fiber.Ctx) error {
	return c.Status(r.status).JSON(r)
}
