package roles

import (
	"github.com/alkuinvito/ai-assistant/pkg/request"
	"github.com/alkuinvito/ai-assistant/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type RoleHandler interface {
	AssignPermissionsToRole(c fiber.Ctx) error
	AssignRolesToUser(c fiber.Ctx) error
	CreatePermission(c fiber.Ctx) error
	CreateRole(c fiber.Ctx) error
	DeletePermission(c fiber.Ctx) error
	DeleteRole(c fiber.Ctx) error
	GetUserAccess(c fiber.Ctx) error
	ListPermissions(c fiber.Ctx) error
	ListRoles(c fiber.Ctx) error
}

type roleHandler struct {
	service RoleService
}

func NewRoleHandler(service RoleService) RoleHandler {
	return &roleHandler{service: service}
}

func (h *roleHandler) AssignPermissionsToRole(c fiber.Ctx) error {
	req := new(DTOAssignPermissionsRequest)
	if err := request.ValidateRequest(c, req); err != nil {
		return err
	}

	if err := h.service.AssignPermissionsToRole(c, c.Params("id"), req); err != nil {
		return err
	}

	return response.New(c, "permissions assigned successfully")
}

func (h *roleHandler) AssignRolesToUser(c fiber.Ctx) error {
	req := new(DTOAssignRoleRequest)
	if err := request.ValidateRequest(c, req); err != nil {
		return err
	}

	if err := h.service.AssignRolesToUser(c, c.Params("userId"), req); err != nil {
		return err
	}

	return response.New(c, "roles assigned successfully")
}

func (h *roleHandler) CreatePermission(c fiber.Ctx) error {
	req := new(DTOCreatePermissionRequest)
	if err := request.ValidateRequest(c, req); err != nil {
		return err
	}

	result, err := h.service.CreatePermission(c, req)
	if err != nil {
		return err
	}

	return response.New(c, result)
}

func (h *roleHandler) CreateRole(c fiber.Ctx) error {
	req := new(DTOCreateRoleRequest)
	if err := request.ValidateRequest(c, req); err != nil {
		return err
	}

	result, err := h.service.CreateRole(c, req)
	if err != nil {
		return err
	}

	return response.New(c, result)
}

func (h *roleHandler) DeletePermission(c fiber.Ctx) error {
	if err := h.service.DeletePermission(c, c.Params("id")); err != nil {
		return err
	}

	return response.New(c, "permission deleted successfully")
}

func (h *roleHandler) DeleteRole(c fiber.Ctx) error {
	if err := h.service.DeleteRole(c, c.Params("id")); err != nil {
		return err
	}

	return response.New(c, "role deleted successfully")
}

func (h *roleHandler) GetUserAccess(c fiber.Ctx) error {
	result, err := h.service.GetUserAccess(c, c.Params("userId"))
	if err != nil {
		return err
	}

	return response.New(c, result)
}

func (h *roleHandler) ListPermissions(c fiber.Ctx) error {
	result, err := h.service.ListPermissions(c)
	if err != nil {
		return err
	}

	return response.New(c, result)
}

func (h *roleHandler) ListRoles(c fiber.Ctx) error {
	result, err := h.service.ListRoles(c)
	if err != nil {
		return err
	}

	return response.New(c, result)
}
