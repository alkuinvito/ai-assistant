package roles

import (
	"context"
	"errors"

	"github.com/alkuinvito/ai-assistant/pkg/service_error"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleService interface {
	AssignPermissionsToRole(ctx context.Context, roleId string, req *DTOAssignPermissionsRequest) error
	AssignRoleByName(ctx context.Context, userId string, roleName string) error
	AssignRolesToUser(ctx context.Context, userId string, req *DTOAssignRoleRequest) error
	CreatePermission(ctx context.Context, req *DTOCreatePermissionRequest) (*DTOPermissionResponse, error)
	CreateRole(ctx context.Context, req *DTOCreateRoleRequest) (*DTORoleResponse, error)
	DeletePermission(ctx context.Context, id string) error
	DeleteRole(ctx context.Context, id string) error
	GetUserAccess(ctx context.Context, userId string) (*DTOUserAccessResponse, error)
	ListPermissions(ctx context.Context) ([]DTOPermissionResponse, error)
	ListRoles(ctx context.Context) ([]DTORoleResponse, error)
	UserHasPermission(ctx context.Context, userId uuid.UUID, permission string) (bool, error)
	UserHasRole(ctx context.Context, userId uuid.UUID, role string) (bool, error)
}

type roleService struct {
	db   *gorm.DB
	repo RoleRepository
}

func NewRoleService(db *gorm.DB, repo RoleRepository) RoleService {
	return &roleService{db: db, repo: repo}
}

func (s *roleService) AssignPermissionsToRole(ctx context.Context, roleId string, req *DTOAssignPermissionsRequest) error {
	db := s.db.WithContext(ctx)

	rid, err := uuid.Parse(roleId)
	if err != nil {
		return service_error.NewBadRequest(ctx, "invalid role id", err)
	}

	if _, err := s.repo.getRoleById(db, rid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return service_error.NewNotFound(ctx, "role not found", err)
		}
		return service_error.NewInternalServerError(ctx, "failed to retrieve role", err)
	}

	records := make([]RolePermission, len(req.PermissionIds))
	for i, raw := range req.PermissionIds {
		pid, err := uuid.Parse(raw)
		if err != nil {
			return service_error.NewBadRequest(ctx, "invalid permission id", err)
		}

		signature, err := signRolePermission(rid, pid)
		if err != nil {
			return service_error.NewInternalServerError(ctx, "failed to sign role permission", err)
		}

		records[i] = RolePermission{RoleId: rid, PermissionId: pid, Signature: signature}
	}

	if err := s.repo.assignPermissionsToRole(db, records); err != nil {
		return service_error.NewInternalServerError(ctx, "failed to assign permissions", err)
	}

	return nil
}

func (s *roleService) AssignRoleByName(ctx context.Context, userId string, roleName string) error {
	db := s.db.WithContext(ctx)

	uid, err := uuid.Parse(userId)
	if err != nil {
		return service_error.NewBadRequest(ctx, "invalid user id", err)
	}

	role, err := s.repo.getRoleByName(db, roleName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return service_error.NewNotFound(ctx, "role not found", err)
		}
		return service_error.NewInternalServerError(ctx, "failed to retrieve role", err)
	}

	signature, err := signUserRole(uid, role.Id)
	if err != nil {
		return service_error.NewInternalServerError(ctx, "failed to sign user role", err)
	}

	records := []UserRole{{UserId: uid, RoleId: role.Id, Signature: signature}}
	if err := s.repo.assignRolesToUser(db, records); err != nil {
		return service_error.NewInternalServerError(ctx, "failed to assign role", err)
	}

	return nil
}

func (s *roleService) AssignRolesToUser(ctx context.Context, userId string, req *DTOAssignRoleRequest) error {
	db := s.db.WithContext(ctx)

	uid, err := uuid.Parse(userId)
	if err != nil {
		return service_error.NewBadRequest(ctx, "invalid user id", err)
	}

	records := make([]UserRole, len(req.RoleIds))
	for i, raw := range req.RoleIds {
		rid, err := uuid.Parse(raw)
		if err != nil {
			return service_error.NewBadRequest(ctx, "invalid role id", err)
		}

		signature, err := signUserRole(uid, rid)
		if err != nil {
			return service_error.NewInternalServerError(ctx, "failed to sign user role", err)
		}

		records[i] = UserRole{UserId: uid, RoleId: rid, Signature: signature}
	}

	if err := s.repo.assignRolesToUser(db, records); err != nil {
		return service_error.NewInternalServerError(ctx, "failed to assign roles", err)
	}

	return nil
}

func (s *roleService) CreatePermission(ctx context.Context, req *DTOCreatePermissionRequest) (*DTOPermissionResponse, error) {
	db := s.db.WithContext(ctx)

	if _, err := s.repo.getPermissionByName(db, req.Name); err == nil {
		return nil, service_error.NewConflict(ctx, "permission already exists", nil)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, service_error.NewInternalServerError(ctx, "failed to check permission", err)
	}

	permission := &Permission{
		Id:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
	}

	signature, err := signPermission(permission.Id, permission.Name)
	if err != nil {
		return nil, service_error.NewInternalServerError(ctx, "failed to sign permission", err)
	}
	permission.Signature = signature

	if err := s.repo.createPermission(db, permission); err != nil {
		return nil, service_error.NewInternalServerError(ctx, "failed to create permission", err)
	}

	return &DTOPermissionResponse{
		Id:          permission.Id,
		Name:        permission.Name,
		Description: permission.Description,
	}, nil
}

func (s *roleService) CreateRole(ctx context.Context, req *DTOCreateRoleRequest) (*DTORoleResponse, error) {
	db := s.db.WithContext(ctx)

	if _, err := s.repo.getRoleByName(db, req.Name); err == nil {
		return nil, service_error.NewConflict(ctx, "role already exists", nil)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, service_error.NewInternalServerError(ctx, "failed to check role", err)
	}

	role := &Role{
		Id:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
	}

	signature, err := signRole(role.Id, role.Name)
	if err != nil {
		return nil, service_error.NewInternalServerError(ctx, "failed to sign role", err)
	}
	role.Signature = signature

	if err := s.repo.createRole(db, role); err != nil {
		return nil, service_error.NewInternalServerError(ctx, "failed to create role", err)
	}

	return &DTORoleResponse{
		Id:          role.Id,
		Name:        role.Name,
		Description: role.Description,
		Permissions: []DTOPermissionResponse{},
	}, nil
}

func (s *roleService) DeletePermission(ctx context.Context, id string) error {
	db := s.db.WithContext(ctx)

	pid, err := uuid.Parse(id)
	if err != nil {
		return service_error.NewBadRequest(ctx, "invalid permission id", err)
	}

	if err := s.repo.deletePermission(db, pid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return service_error.NewNotFound(ctx, "permission not found", err)
		}
		return service_error.NewInternalServerError(ctx, "failed to delete permission", err)
	}

	return nil
}

func (s *roleService) DeleteRole(ctx context.Context, id string) error {
	db := s.db.WithContext(ctx)

	rid, err := uuid.Parse(id)
	if err != nil {
		return service_error.NewBadRequest(ctx, "invalid role id", err)
	}

	if err := s.repo.deleteRole(db, rid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return service_error.NewNotFound(ctx, "role not found", err)
		}
		return service_error.NewInternalServerError(ctx, "failed to delete role", err)
	}

	return nil
}

func (s *roleService) GetUserAccess(ctx context.Context, userId string) (*DTOUserAccessResponse, error) {
	db := s.db.WithContext(ctx)

	uid, err := uuid.Parse(userId)
	if err != nil {
		return nil, service_error.NewBadRequest(ctx, "invalid user id", err)
	}

	resolvedRoles, resolvedPermissions, err := s.getVerifiedAccess(ctx, db, uid)
	if err != nil {
		return nil, err
	}

	roleNames := make([]string, len(resolvedRoles))
	for i, role := range resolvedRoles {
		roleNames[i] = role.Name
	}

	permissionNames := make([]string, len(resolvedPermissions))
	for i, permission := range resolvedPermissions {
		permissionNames[i] = permission.Name
	}

	return &DTOUserAccessResponse{
		Roles:       roleNames,
		Permissions: permissionNames,
	}, nil
}

func (s *roleService) ListPermissions(ctx context.Context) ([]DTOPermissionResponse, error) {
	db := s.db.WithContext(ctx)

	permissions, err := s.repo.listPermissions(db)
	if err != nil {
		return nil, service_error.NewInternalServerError(ctx, "failed to list permissions", err)
	}

	result := make([]DTOPermissionResponse, len(permissions))
	for i, permission := range permissions {
		result[i] = DTOPermissionResponse{
			Id:          permission.Id,
			Name:        permission.Name,
			Description: permission.Description,
		}
	}

	return result, nil
}

func (s *roleService) ListRoles(ctx context.Context) ([]DTORoleResponse, error) {
	db := s.db.WithContext(ctx)

	roles, err := s.repo.listRoles(db)
	if err != nil {
		return nil, service_error.NewInternalServerError(ctx, "failed to list roles", err)
	}

	result := make([]DTORoleResponse, len(roles))
	for i, role := range roles {
		permissions, err := s.repo.getPermissionsByRoleId(db, role.Id)
		if err != nil {
			return nil, service_error.NewInternalServerError(ctx, "failed to retrieve role permissions", err)
		}

		permissionDTOs := make([]DTOPermissionResponse, len(permissions))
		for j, permission := range permissions {
			permissionDTOs[j] = DTOPermissionResponse{
				Id:          permission.Id,
				Name:        permission.Name,
				Description: permission.Description,
			}
		}

		result[i] = DTORoleResponse{
			Id:          role.Id,
			Name:        role.Name,
			Description: role.Description,
			Permissions: permissionDTOs,
		}
	}

	return result, nil
}

func (s *roleService) UserHasPermission(ctx context.Context, userId uuid.UUID, permission string) (bool, error) {
	db := s.db.WithContext(ctx)

	_, permissions, err := s.getVerifiedAccess(ctx, db, userId)
	if err != nil {
		return false, err
	}

	for _, item := range permissions {
		if item.Name == permission {
			return true, nil
		}
	}

	return false, nil
}

func (s *roleService) UserHasRole(ctx context.Context, userId uuid.UUID, role string) (bool, error) {
	db := s.db.WithContext(ctx)

	roles, _, err := s.getVerifiedAccess(ctx, db, userId)
	if err != nil {
		return false, err
	}

	for _, item := range roles {
		if item.Name == role {
			return true, nil
		}
	}

	return false, nil
}

// getVerifiedAccess resolves a user's roles and permissions from the database,
// verifying the tamper-detection signature on every assignment row.
func (s *roleService) getVerifiedAccess(ctx context.Context, db *gorm.DB, userId uuid.UUID) ([]Role, []Permission, error) {
	userRoles, err := s.repo.getUserRoles(db, userId)
	if err != nil {
		return nil, nil, service_error.NewInternalServerError(ctx, "failed to retrieve user roles", err)
	}

	roleIds := make([]uuid.UUID, len(userRoles))
	for i, record := range userRoles {
		if err := verifyUserRole(record); err != nil {
			return nil, nil, service_error.NewInternalServerError(ctx, "rbac integrity verification failed", err)
		}
		roleIds[i] = record.RoleId
	}

	resolvedRoles, err := s.repo.getRolesByIds(db, roleIds)
	if err != nil {
		return nil, nil, service_error.NewInternalServerError(ctx, "failed to retrieve roles", err)
	}

	for _, role := range resolvedRoles {
		if err := verifyRole(role); err != nil {
			return nil, nil, service_error.NewInternalServerError(ctx, "rbac integrity verification failed", err)
		}
	}

	rolePermissions, err := s.repo.getRolePermissions(db, roleIds)
	if err != nil {
		return nil, nil, service_error.NewInternalServerError(ctx, "failed to retrieve role permissions", err)
	}

	permissionIdSet := make(map[uuid.UUID]struct{}, len(rolePermissions))
	for _, record := range rolePermissions {
		if err := verifyRolePermission(record); err != nil {
			return nil, nil, service_error.NewInternalServerError(ctx, "rbac integrity verification failed", err)
		}
		permissionIdSet[record.PermissionId] = struct{}{}
	}

	permissionIds := make([]uuid.UUID, 0, len(permissionIdSet))
	for id := range permissionIdSet {
		permissionIds = append(permissionIds, id)
	}

	resolvedPermissions, err := s.repo.getPermissionsByIds(db, permissionIds)
	if err != nil {
		return nil, nil, service_error.NewInternalServerError(ctx, "failed to retrieve permissions", err)
	}

	for _, permission := range resolvedPermissions {
		if err := verifyPermission(permission); err != nil {
			return nil, nil, service_error.NewInternalServerError(ctx, "rbac integrity verification failed", err)
		}
	}

	return resolvedRoles, resolvedPermissions, nil
}
