package auth

import (
	"github.com/alkuinvito/ai-assistant/pkg/request"
	"github.com/alkuinvito/ai-assistant/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type AuthHandler interface {
	Login(c fiber.Ctx) error
	Register(c fiber.Ctx) error
}

type authHandler struct {
	service AuthService
}

func NewAuthHandler(service AuthService) AuthHandler {
	return &authHandler{service: service}
}

func (h *authHandler) Login(c fiber.Ctx) error {
	req := new(DTOLoginRequest)
	if err := request.ValidateRequest(c, req); err != nil {
		return err
	}

	data, err := h.service.Login(c, req)
	if err != nil {
		return err
	}

	return response.New(data).JSON(c)
}

func (h *authHandler) Register(c fiber.Ctx) error {
	req := new(DTORegisterRequest)
	if err := request.ValidateRequest(c, req); err != nil {
		return err
	}

	err := h.service.Register(c, req)
	if err != nil {
		return err
	}

	return response.New("User registered successfully").JSON(c)
}
