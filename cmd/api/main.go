package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/alkuinvito/ai-assistant/internal/routers"
	"github.com/alkuinvito/ai-assistant/pkg/logger"
	"github.com/gofiber/fiber/v3"
)

func NewApp(router *routers.Router) *fiber.App {
	return router.Handle()
}

func main() {
	log := logger.NewLogger()

	app, cleanup, err := NewHttpServer(log)
	if err != nil {
		log.Fatal("failed to create server instance", err)
		return
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8000"
	}

	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Info("server stopped", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	// Run cleanup function for all services
	cleanup()

	// Graceful shutdown — waits for in-flight requests to finish
	if err := app.Shutdown(); err != nil {
		log.Fatal("forced shutdown", err)
	}

	log.Info("server exited cleanly")
}
