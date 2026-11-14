//go:build wireinject
// +build wireinject

package main

import (
	"github.com/alkuinvito/ai-assistant/internal/auth"
	"github.com/alkuinvito/ai-assistant/internal/invoices"
	"github.com/alkuinvito/ai-assistant/internal/middlewares"
	"github.com/alkuinvito/ai-assistant/internal/roles"
	"github.com/alkuinvito/ai-assistant/internal/routers"
	"github.com/alkuinvito/ai-assistant/internal/uploads"
	"github.com/alkuinvito/ai-assistant/internal/users"
	"github.com/alkuinvito/ai-assistant/pkg/cache"
	"github.com/alkuinvito/ai-assistant/pkg/database"
	"github.com/alkuinvito/ai-assistant/pkg/mailer"
	"github.com/alkuinvito/ai-assistant/pkg/storage"
	"github.com/gofiber/fiber/v3"
	"github.com/google/wire"
	"github.com/sirupsen/logrus"
)

var authSet = wire.NewSet(
	auth.NewAuthRouter,
	auth.NewAuthHandler,
	auth.NewAuthService,
)

var invoiceSet = wire.NewSet(
	invoices.NewInvoiceHandler,
	invoices.NewInvoiceRepository,
	invoices.NewInvoiceRouter,
	invoices.NewInvoiceService,
)

var roleSet = wire.NewSet(
	roles.NewRoleHandler,
	roles.NewRoleRepository,
	roles.NewRoleRouter,
	roles.NewRoleService,
)

var uploadSet = wire.NewSet(
	uploads.NewUploadRepository,
	uploads.NewUploadService,
)

var userSet = wire.NewSet(
	users.NewUserRepository,
	users.NewUserService,
	users.NewUserHandler,
	users.NewUserRouter,
)

func NewHttpServer(log *logrus.Logger) (*fiber.App, func(), error) {
	panic(wire.Build(
		cache.NewRedisCache,
		database.NewDatabase,
		mailer.NewSmtpMail,
		storage.NewMinioStorage,
		middlewares.NewMiddleware,
		authSet,
		invoiceSet,
		roleSet,
		uploadSet,
		userSet,
		routers.NewRouter,
		NewApp,
	))
}
