package roles

import "github.com/google/uuid"

type DTOAssignPermissionsRequest struct {
	PermissionIds []string `json:"permission_ids" validate:"required,min=1,dive,uuid"`
}

type DTOAssignRoleRequest struct {
	RoleIds []string `json:"role_ids" validate:"required,min=1,dive,uuid"`
}

type DTOCreatePermissionRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=255"`
}

type DTOCreateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=255"`
}

type DTOPermissionResponse struct {
	Id          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type DTORoleResponse struct {
	Id          uuid.UUID               `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Permissions []DTOPermissionResponse `json:"permissions"`
}

type DTOUserAccessResponse struct {
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}
