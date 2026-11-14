package users

import (
	"github.com/alkuinvito/ai-assistant/pkg/request"
	"github.com/alkuinvito/ai-assistant/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type UserHandler interface {
	GetProfile(c fiber.Ctx) error
	UpdateProfile(c fiber.Ctx) error
}

type userHandler struct {
	service UserService
}

func NewUserHandler(service UserService) UserHandler {
	return &userHandler{service: service}
}

func (h *userHandler) GetProfile(c fiber.Ctx) error {
	userId, ok := c.Locals("user_id").(string)
	if !ok || userId == "" {
		return fiber.ErrUnauthorized
	}

	result, err := h.service.GetProfile(c, userId)
	if err != nil {
		return err
	}

	return response.New(c, result)
}

func (h *userHandler) UpdateProfile(c fiber.Ctx) error {
	req := new(DTOUpdateProfileRequest)
	if err := request.ValidateRequest(c, req); err != nil {
		return err
	}

	userId, ok := c.Locals("user_id").(string)
	if !ok || userId == "" {
		return fiber.ErrUnauthorized
	}

	result, err := h.service.UpdateProfile(c, userId, req)
	if err != nil {
		return err
	}

	return response.New(c, result)
}
