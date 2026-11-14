package main

import (
	"context"
	"os"

	"github.com/alkuinvito/ai-assistant/internal/invoices"
	"github.com/alkuinvito/ai-assistant/internal/roles"
	"github.com/alkuinvito/ai-assistant/internal/uploads"
	"github.com/alkuinvito/ai-assistant/internal/users"
	"github.com/alkuinvito/ai-assistant/pkg/database"
	"github.com/alkuinvito/ai-assistant/pkg/hash"
	"github.com/alkuinvito/ai-assistant/pkg/logger"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func init() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
}

func main() {
	log := logger.NewLogger()

	log.Info("Migrating database...")

	db, cleanup, err := database.NewDatabase(log)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	if err := db.AutoMigrate(
		&uploads.File{},
		&uploads.UserQuota{},
		&invoices.Invoice{},
		&invoices.InvoiceItem{},
		&users.User{},
		&roles.Role{},
		&roles.Permission{},
		&roles.RolePermission{},
		&roles.UserRole{},
	); err != nil {
		log.Fatal(err)
	}

	if err := roles.Seed(db); err != nil {
		log.Fatal(err)
	}

	bootstrapAdmin(db, log)

	log.Info("Database migrated successfully")
}

// bootstrapAdmin assigns the "admin" role to the user identified by ADMIN_EMAIL.
// This is how the first administrator is created, since role management requires
// an existing admin.
func bootstrapAdmin(db *gorm.DB, log *logrus.Logger) {
	email := os.Getenv("ADMIN_EMAIL")
	if email == "" {
		return
	}

	ctx := context.Background()

	userService := users.NewUserService(db, users.NewUserRepository())
	admin, err := userService.GetUserByEmail(db, hash.HashString(email))
	if err != nil {
		log.Warnf("bootstrap admin skipped: %v", err)
		return
	}

	roleService := roles.NewRoleService(db, roles.NewRoleRepository())
	roleList, err := roleService.ListRoles(ctx)
	if err != nil {
		log.Warnf("bootstrap admin skipped: %v", err)
		return
	}

	for _, role := range roleList {
		if role.Name != "admin" {
			continue
		}

		err := roleService.AssignRolesToUser(ctx, admin.Id.String(), &roles.DTOAssignRoleRequest{
			RoleIds: []string{role.Id.String()},
		})
		if err != nil {
			log.Warnf("failed to assign admin role: %v", err)
			return
		}

		log.Infof("assigned admin role to %s", email)
		return
	}

	log.Warn("bootstrap admin skipped: admin role not found")
}
