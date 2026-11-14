package roles

import "github.com/gofiber/fiber/v3"

type RoleRouter interface {
	Handle(router fiber.Router)
}

type roleRouter struct {
	handler RoleHandler
}

func NewRoleRouter(handler RoleHandler) RoleRouter {
	return &roleRouter{handler: handler}
}

func (r *roleRouter) Handle(router fiber.Router) {
	router.Get("/roles", r.handler.ListRoles)
	router.Post("/roles", r.handler.CreateRole)
	router.Delete("/roles/:id", r.handler.DeleteRole)
	router.Post("/roles/:id/permissions", r.handler.AssignPermissionsToRole)

	router.Get("/permissions", r.handler.ListPermissions)
	router.Post("/permissions", r.handler.CreatePermission)
	router.Delete("/permissions/:id", r.handler.DeletePermission)

	router.Post("/users/:userId/roles", r.handler.AssignRolesToUser)
	router.Get("/users/:userId/access", r.handler.GetUserAccess)
}
