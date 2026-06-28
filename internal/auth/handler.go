package auth

import (
	"github.com/alkuinvito/ai-assistant/pkg/request"
	"github.com/alkuinvito/ai-assistant/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type AuthHandler interface {
	Login(c fiber.Ctx) error
	Register(c fiber.Ctx) error
	SendVerification(c fiber.Ctx) error
	VerifyEmail(c fiber.Ctx) error
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

	return c.JSON(response.New(data))
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

	return c.JSON(response.New("user registered successfully"))
}

func (h *authHandler) SendVerification(c fiber.Ctx) error {
	req := new(DTOSendVerificationRequest)
	if err := request.ValidateRequest(c, req); err != nil {
		return err
	}

	if err := h.service.GenerateAndSendVerificationToken(c, req.Email); err != nil {
		return err
	}

	return c.JSON(response.New("verification email sent"))
}

func (h *authHandler) VerifyEmail(c fiber.Ctx) error {
	req := new(DTOVerifyEmailRequest)
	if err := request.ValidateRequest(c, req); err != nil {
		return err
	}

	if err := h.service.VerifyEmail(c, req); err != nil {
		return err
	}

	return c.JSON(response.New("email verified successfully"))
}
