package auth

import (
	"os"
	"strings"

	"github.com/alkuinvito/ai-assistant/pkg/request"
	"github.com/alkuinvito/ai-assistant/pkg/response"
	"github.com/alkuinvito/ai-assistant/pkg/service_error"
	"github.com/gofiber/fiber/v3"
)

const refreshCookieName = "aa.refresh_token"

type AuthHandler interface {
	Login(c fiber.Ctx) error
	Logout(c fiber.Ctx) error
	Refresh(c fiber.Ctx) error
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

	clientType, ok := c.Value("client_type").(string)
	if !ok || clientType == "" {
		return service_error.NewUnauthorized(c, "invalid client type", err)
	}

	if clientType == "web" {
		h.setRefreshCookie(c, data.RefreshToken)
		data.RefreshToken = ""
	}

	return response.New(c, data)
}

func (h *authHandler) Logout(c fiber.Ctx) error {
	clientType, _ := c.Value("client_type").(string)

	var refreshToken string
	if clientType == "web" {
		refreshToken = c.Cookies(refreshCookieName)
	} else {
		req := new(DTOLogoutRequest)
		_ = c.Bind().Body(req)
		refreshToken = req.RefreshToken
	}

	if err := h.service.Logout(c, refreshToken); err != nil {
		return err
	}

	if clientType == "web" {
		h.clearRefreshCookie(c)
	}

	return response.New(c, "logged out successfully")
}

func (h *authHandler) Refresh(c fiber.Ctx) error {
	clientType, ok := c.Value("client_type").(string)
	if !ok || clientType == "" {
		return service_error.NewUnauthorized(c, "invalid client type", nil)
	}

	var refreshToken string
	if clientType == "web" {
		refreshToken = c.Cookies(refreshCookieName)
	} else {
		req := new(DTORefreshRequest)
		if err := request.ValidateRequest(c, req); err != nil {
			return err
		}
		refreshToken = req.RefreshToken
	}

	if refreshToken == "" {
		return service_error.NewUnauthorized(c, "missing refresh token", nil)
	}

	data, err := h.service.Refresh(c, refreshToken)
	if err != nil {
		return err
	}

	if clientType == "web" {
		h.setRefreshCookie(c, data.RefreshToken)
		data.RefreshToken = ""
	}

	return response.New(c, data)
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

	return response.New(c, "user registered successfully")
}

func (h *authHandler) SendVerification(c fiber.Ctx) error {
	req := new(DTOSendVerificationRequest)
	if err := request.ValidateRequest(c, req); err != nil {
		return err
	}

	if err := h.service.GenerateAndSendVerificationToken(c, req.Email); err != nil {
		return err
	}

	return response.New(c, "verification email sent")
}

func (h *authHandler) VerifyEmail(c fiber.Ctx) error {
	req := new(DTOVerifyEmailRequest)
	if err := request.ValidateRequest(c, req); err != nil {
		return err
	}

	if err := h.service.VerifyEmail(c, req); err != nil {
		return err
	}

	return response.New(c, "email verified successfully")
}

func (h *authHandler) setRefreshCookie(c fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/api/auth",
		SameSite: cookieSameSite(),
		MaxAge:   h.service.GetRefreshTokenExpiresIn(),
		HTTPOnly: true,
		Secure:   os.Getenv("APP_ENV") != "development",
	})
}

func (h *authHandler) clearRefreshCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/api/auth",
		SameSite: cookieSameSite(),
		MaxAge:   -1,
		HTTPOnly: true,
		Secure:   os.Getenv("APP_ENV") != "development",
	})
}

// cookieSameSite returns the SameSite mode for the refresh cookie. Set
// COOKIE_SAME_SITE=none for cross-site web deployments (requires Secure).
func cookieSameSite() string {
	switch strings.ToLower(os.Getenv("COOKIE_SAME_SITE")) {
	case "none":
		return fiber.CookieSameSiteNoneMode
	case "strict":
		return fiber.CookieSameSiteStrictMode
	default:
		return fiber.CookieSameSiteLaxMode
	}
}
