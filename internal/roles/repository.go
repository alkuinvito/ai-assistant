package roles

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RoleRepository interface {
	assignPermissionsToRole(tx *gorm.DB, records []RolePermission) error
	assignRolesToUser(tx *gorm.DB, records []UserRole) error
	createPermission(tx *gorm.DB, permission *Permission) error
	createRole(tx *gorm.DB, role *Role) error
	deletePermission(tx *gorm.DB, id uuid.UUID) error
	deleteRole(tx *gorm.DB, id uuid.UUID) error
	getPermissionByName(tx *gorm.DB, name string) (*Permission, error)
	getPermissionsByIds(tx *gorm.DB, ids []uuid.UUID) ([]Permission, error)
	getPermissionsByRoleId(tx *gorm.DB, roleId uuid.UUID) ([]Permission, error)
	getRoleById(tx *gorm.DB, id uuid.UUID) (*Role, error)
	getRoleByName(tx *gorm.DB, name string) (*Role, error)
	getRolePermissions(tx *gorm.DB, roleIds []uuid.UUID) ([]RolePermission, error)
	getRolesByIds(tx *gorm.DB, ids []uuid.UUID) ([]Role, error)
	getUserRoles(tx *gorm.DB, userId uuid.UUID) ([]UserRole, error)
	listPermissions(tx *gorm.DB) ([]Permission, error)
	listRoles(tx *gorm.DB) ([]Role, error)
}

type roleRepository struct{}

func NewRoleRepository() RoleRepository {
	return &roleRepository{}
}

func (r *roleRepository) assignPermissionsToRole(tx *gorm.DB, records []RolePermission) error {
	if len(records) == 0 {
		return nil
	}

	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&records).Error
}

func (r *roleRepository) assignRolesToUser(tx *gorm.DB, records []UserRole) error {
	if len(records) == 0 {
		return nil
	}

	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&records).Error
}

func (r *roleRepository) createPermission(tx *gorm.DB, permission *Permission) error {
	return tx.Create(permission).Error
}

func (r *roleRepository) createRole(tx *gorm.DB, role *Role) error {
	return tx.Create(role).Error
}

func (r *roleRepository) deletePermission(tx *gorm.DB, id uuid.UUID) error {
	return tx.Transaction(func(d *gorm.DB) error {
		var permission Permission
		if err := d.First(&permission, "id = ?", id).Error; err != nil {
			return err
		}

		if err := d.Where("permission_id = ?", id).Delete(&RolePermission{}).Error; err != nil {
			return err
		}

		return d.Delete(&permission).Error
	})
}

func (r *roleRepository) deleteRole(tx *gorm.DB, id uuid.UUID) error {
	return tx.Transaction(func(d *gorm.DB) error {
		var role Role
		if err := d.First(&role, "id = ?", id).Error; err != nil {
			return err
		}

		if err := d.Where("role_id = ?", id).Delete(&RolePermission{}).Error; err != nil {
			return err
		}

		if err := d.Where("role_id = ?", id).Delete(&UserRole{}).Error; err != nil {
			return err
		}

		return d.Delete(&role).Error
	})
}

func (r *roleRepository) getPermissionByName(tx *gorm.DB, name string) (*Permission, error) {
	var permission Permission
	if err := tx.First(&permission, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *roleRepository) getPermissionsByIds(tx *gorm.DB, ids []uuid.UUID) ([]Permission, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var permissions []Permission
	err := tx.Where("id IN ?", ids).Order("name").Find(&permissions).Error
	return permissions, err
}

func (r *roleRepository) getPermissionsByRoleId(tx *gorm.DB, roleId uuid.UUID) ([]Permission, error) {
	var permissions []Permission
	err := tx.Table("permissions").
		Select("permissions.*").
		Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Where("rp.role_id = ?", roleId).
		Order("permissions.name").
		Find(&permissions).Error
	return permissions, err
}

func (r *roleRepository) getRoleById(tx *gorm.DB, id uuid.UUID) (*Role, error) {
	var role Role
	if err := tx.First(&role, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) getRoleByName(tx *gorm.DB, name string) (*Role, error) {
	var role Role
	if err := tx.First(&role, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) getRolePermissions(tx *gorm.DB, roleIds []uuid.UUID) ([]RolePermission, error) {
	if len(roleIds) == 0 {
		return nil, nil
	}

	var records []RolePermission
	err := tx.Where("role_id IN ?", roleIds).Find(&records).Error
	return records, err
}

func (r *roleRepository) getRolesByIds(tx *gorm.DB, ids []uuid.UUID) ([]Role, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var roles []Role
	err := tx.Where("id IN ?", ids).Order("name").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) getUserRoles(tx *gorm.DB, userId uuid.UUID) ([]UserRole, error) {
	var records []UserRole
	err := tx.Where("user_id = ?", userId).Find(&records).Error
	return records, err
}

func (r *roleRepository) listPermissions(tx *gorm.DB) ([]Permission, error) {
	var permissions []Permission
	err := tx.Order("name").Find(&permissions).Error
	return permissions, err
}

func (r *roleRepository) listRoles(tx *gorm.DB) ([]Role, error) {
	var roles []Role
	err := tx.Order("name").Find(&roles).Error
	return roles, err
}
