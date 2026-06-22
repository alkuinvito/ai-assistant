package database

import (
	"os"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase(logger *logrus.Logger) (*gorm.DB, func(), error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  os.Getenv("DATABASE_URL"),
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		TranslateError: true,
	})

	return db, cleanup(db, logger), err
}

func cleanup(db *gorm.DB, logger *logrus.Logger) func() {
	return func() {
		logger.Info("cleaning up database")

		sqlDB, err := db.DB()
		if err != nil {
			logger.Fatal("failed to get sql db", err)
		}

		err = sqlDB.Close()
		if err != nil {
			logger.Fatal("failed to close sql db", err)
		}
	}
}
