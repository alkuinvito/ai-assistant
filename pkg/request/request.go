package request

import (
	"fmt"

	"github.com/alkuinvito/ai-assistant/pkg/service_error"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type RequestParams struct {
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
	Offset int    `query:"-"`
	Search string `query:"search"`
}

func ParseRequestParams(c fiber.Ctx) (*RequestParams, error) {
	params := &RequestParams{}

	if err := c.Bind().Query(&params); err != nil {
		return nil, err
	}

	params.Offset = (params.Page - 1) * params.Limit

	return params, nil
}

func ValidateRequest[T any](c fiber.Ctx, v T) error {
	if err := c.Bind().Body(v); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				errInstance := fmt.Errorf("%s", e.Error())
				return service_error.NewUnprocessableEntity(c, fmt.Sprintf("field validation for '%s' failed on the '%s' tag", e.Field(), e.Tag()), errInstance)
			}
		}

		return service_error.NewUnprocessableEntity(c, "invalid request body", err)
	}

	return nil
}
