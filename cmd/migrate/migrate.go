package main

import (
	"github.com/alkuinvito/ai-assistant/internal/users"
	"github.com/alkuinvito/ai-assistant/pkg/database"
	"github.com/alkuinvito/ai-assistant/pkg/logger"
	"github.com/joho/godotenv"
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
		&users.User{},
	); err != nil {
		log.Fatal(err)
	}

	log.Info("Database migrated successfully")
}
