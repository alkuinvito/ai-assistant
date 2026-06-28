package response

import (
	"errors"

	"github.com/alkuinvito/ai-assistant/pkg/request"
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

type Response[T any] struct {
	Data   T              `json:"data"`
	Error  *ErrorResponse `json:"error,omitempty"`
	Meta   *Meta          `json:"meta,omitempty"`
	status int            `json:"-"`
}

func New[T any](data T) *Response[T] {
	return &Response[T]{
		status: fiber.StatusOK,
		Data:   data,
	}
}

func NewError(err error) *Response[any] {
	resp := &Response[any]{}
	return resp.WithError(err)
}

func (r *Response[T]) WithError(err error) *Response[T] {
	if err == nil {
		return r
	}

	var appError *service_error.ServiceError
	if errors.As(err, &appError) {
		r.status = appError.StatusCode()
		r.Error = &ErrorResponse{
			Message: appError.Error(),
		}
	} else {
		r.status = fiber.StatusInternalServerError
		r.Error = &ErrorResponse{
			Message: err.Error(),
		}
	}
	return r
}

func (r *Response[T]) WithMeta(meta *Meta) *Response[T] {
	r.Meta = meta
	return r
}

func (r *Response[T]) WithPagination(params *request.RequestParams, totalCount int64) *Response[T] {
	r.Meta = &Meta{
		Total: totalCount,
		Page:  params.Page,
		Limit: params.Limit,
	}
	return r
}
