package users

import "github.com/gofiber/fiber/v3"

// UserGuards carries the authorization handlers applied to user routes.
// They are injected by the router assembly so this package does not depend on
// the middleware package (which would create an import cycle through auth).
type UserGuards struct {
	ProfileWrite fiber.Handler
}

type UserRouter interface {
	Handle(router fiber.Router, guards UserGuards)
}

type userRouter struct {
	handler UserHandler
}

func NewUserRouter(handler UserHandler) UserRouter {
	return &userRouter{handler: handler}
}

func (r *userRouter) Handle(router fiber.Router, guards UserGuards) {
	router.Route("/users", func(group fiber.Router) {
		group.Get("/me", r.handler.GetProfile)
		group.Put("/me", guards.ProfileWrite, r.handler.UpdateProfile)
	})
}
