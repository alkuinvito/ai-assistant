package users

import "github.com/gofiber/fiber/v3"

type UserRouter interface {
	Handle(router fiber.Router)
}

type userRouter struct {
	handler UserHandler
}

func NewUserRouter(handler UserHandler) UserRouter {
	return &userRouter{handler: handler}
}

func (r *userRouter) Handle(router fiber.Router) {
	router.Route("/users", func(group fiber.Router) {
	})
}
