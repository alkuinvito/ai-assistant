package auth

import "github.com/gofiber/fiber/v3"

type AuthRouter interface {
	Handle(router fiber.Router)
}

type authRouter struct {
	handler AuthHandler
}

func NewAuthRouter(handler AuthHandler) AuthRouter {
	return &authRouter{handler: handler}
}

func (r *authRouter) Handle(router fiber.Router) {
	router.Route("/auth", func(group fiber.Router) {
		group.Post("/register", r.handler.Register)
		group.Post("/login", r.handler.Login)
	})
}
