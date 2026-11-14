package roles

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var defaultPermissions = []struct {
	Name        string
	Description string
}{
	{Name: "file:read", Description: "Read uploaded files"},
	{Name: "file:write", Description: "Upload and manage files"},
	{Name: "invoice:read", Description: "Read invoices"},
	{Name: "invoice:write", Description: "Create and manage invoices"},
	{Name: "permission:read", Description: "Read permissions"},
	{Name: "permission:write", Description: "Create and manage permissions"},
	{Name: "profile:write", Description: "Update own profile"},
	{Name: "role:read", Description: "Read roles"},
	{Name: "role:write", Description: "Create and manage roles"},
}

var defaultRoles = map[string][]string{
	"admin": {
		"file:read",
		"file:write",
		"invoice:read",
		"invoice:write",
		"permission:read",
		"permission:write",
		"profile:write",
		"role:read",
		"role:write",
	},
	"user": {
		"file:read",
		"file:write",
		"invoice:read",
		"invoice:write",
		"profile:write",
	},
}

// Seed creates the default permissions and roles. It is idempotent so it is
// safe to run on every migration.
func Seed(db *gorm.DB) error {
	for _, permission := range defaultPermissions {
		model := Permission{
			Id:          uuid.New(),
			Name:        permission.Name,
			Description: permission.Description,
		}

		signature, err := signPermission(model.Id, model.Name)
		if err != nil {
			return err
		}
		model.Signature = signature

		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoNothing: true,
		}).Create(&model).Error; err != nil {
			return err
		}
	}

	for roleName, permissionNames := range defaultRoles {
		roleModel := Role{
			Id:          uuid.New(),
			Name:        roleName,
			Description: roleName + " role",
		}

		signature, err := signRole(roleModel.Id, roleModel.Name)
		if err != nil {
			return err
		}
		roleModel.Signature = signature

		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoNothing: true,
		}).Create(&roleModel).Error; err != nil {
			return err
		}

		var storedRole Role
		if err := db.First(&storedRole, "name = ?", roleName).Error; err != nil {
			return err
		}

		var permissions []Permission
		if err := db.Where("name IN ?", permissionNames).Find(&permissions).Error; err != nil {
			return err
		}

		if len(permissions) == 0 {
			continue
		}

		records := make([]RolePermission, len(permissions))
		for i, permission := range permissions {
			recordSignature, err := signRolePermission(storedRole.Id, permission.Id)
			if err != nil {
				return err
			}

			records[i] = RolePermission{
				RoleId:       storedRole.Id,
				PermissionId: permission.Id,
				Signature:    recordSignature,
			}
		}

		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&records).Error; err != nil {
			return err
		}
	}

	if err := backfillPermissionSignatures(db); err != nil {
		return err
	}

	if err := backfillRoleSignatures(db); err != nil {
		return err
	}

	if err := backfillRolePermissionSignatures(db); err != nil {
		return err
	}

	return backfillUserRoleSignatures(db)
}

// backfillPermissionSignatures signs rows created before tamper detection was
// introduced. It only touches unsigned rows, so it never "re-blesses" rows
// that were modified after being signed.
func backfillPermissionSignatures(db *gorm.DB) error {
	var records []Permission
	if err := db.Where("signature = ''").Find(&records).Error; err != nil {
		return err
	}

	for _, record := range records {
		signature, err := signPermission(record.Id, record.Name)
		if err != nil {
			return err
		}

		if err := db.Model(&Permission{}).
			Where("id = ?", record.Id).
			Update("signature", signature).Error; err != nil {
			return err
		}
	}

	return nil
}

// backfillRoleSignatures signs roles created before tamper detection existed.
func backfillRoleSignatures(db *gorm.DB) error {
	var records []Role
	if err := db.Where("signature = ''").Find(&records).Error; err != nil {
		return err
	}

	for _, record := range records {
		signature, err := signRole(record.Id, record.Name)
		if err != nil {
			return err
		}

		if err := db.Model(&Role{}).
			Where("id = ?", record.Id).
			Update("signature", signature).Error; err != nil {
			return err
		}
	}

	return nil
}

// backfillRolePermissionSignatures signs rows created before tamper detection
// was introduced.
func backfillRolePermissionSignatures(db *gorm.DB) error {
	var records []RolePermission
	if err := db.Where("signature = ''").Find(&records).Error; err != nil {
		return err
	}

	for _, record := range records {
		signature, err := signRolePermission(record.RoleId, record.PermissionId)
		if err != nil {
			return err
		}

		if err := db.Model(&RolePermission{}).
			Where("role_id = ? AND permission_id = ?", record.RoleId, record.PermissionId).
			Update("signature", signature).Error; err != nil {
			return err
		}
	}

	return nil
}

// backfillUserRoleSignatures signs user-role rows created before tamper
// detection was introduced.
func backfillUserRoleSignatures(db *gorm.DB) error {
	var records []UserRole
	if err := db.Where("signature = ''").Find(&records).Error; err != nil {
		return err
	}

	for _, record := range records {
		signature, err := signUserRole(record.UserId, record.RoleId)
		if err != nil {
			return err
		}

		if err := db.Model(&UserRole{}).
			Where("user_id = ? AND role_id = ?", record.UserId, record.RoleId).
			Update("signature", signature).Error; err != nil {
			return err
		}
	}

	return nil
}
