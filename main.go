package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/tomasacebal/go-media-api/internal/auth"
	"github.com/tomasacebal/go-media-api/internal/config"
	"github.com/tomasacebal/go-media-api/internal/media"
	"github.com/tomasacebal/go-media-api/internal/storage"
	"github.com/tomasacebal/go-media-api/internal/web"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config invalida: %v", err)
	}

	app, cleanup, err := buildApp(cfg, log.Default())
	if err != nil {
		log.Fatalf("inicializar app: %v", err)
	}
	defer func() {
		if err := cleanup(); err != nil {
			log.Printf("cerrar recursos: %v", err)
		}
	}()

	if err := app.Listen(":" + cfg.Port); err != nil {
		panic(err)
	}
}

func buildApp(cfg config.Config, logger *log.Logger) (*fiber.App, func() error, error) {
	repo, closeRepo, err := media.NewSQLiteRepository(cfg.Media.SQLitePath)
	if err != nil {
		return nil, nil, err
	}

	keyRepo, closeKeyRepo, err := auth.NewSQLiteKeyRepository(cfg.Media.SQLitePath)
	if err != nil {
		_ = closeRepo()
		return nil, nil, err
	}

	storageProvider, err := storage.NewLocalProvider(cfg.Media.StoragePath, cfg.Media.PublicBaseURL)
	if err != nil {
		_ = closeKeyRepo()
		_ = closeRepo()
		return nil, nil, err
	}

	app := fiber.New(fiber.Config{
		BodyLimit:    requestBodyLimit(cfg.Media.MaxUploadBytes),
		ErrorHandler: media.JSONErrorHandler,
	})

	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)
		logger.Printf("%s %s %d %s", c.Method(), c.Path(), c.Response().StatusCode(), duration)

		return err
	})

	keyService := auth.NewAPIKeyService(keyRepo)
	sessionManager := auth.NewSessionManager(cfg.Auth.AdminUsername, cfg.Auth.AdminPassword, cfg.Auth.SessionSecret)
	authMiddleware := auth.NewMiddleware(sessionManager, keyService)
	authHandler := auth.NewHandler(sessionManager, keyService, authMiddleware)
	authHandler.RegisterRoutes(app)
	registerBaseRoutes(app, cfg, authMiddleware)
	mediaService := media.NewService(repo, storageProvider, cfg.Media.MaxUploadBytes, cfg.Media.PublicBaseURL)
	mediaHandler := media.NewHandler(mediaService, logger, authMiddleware)
	mediaHandler.RegisterRoutes(app)

	cleanup := func() error {
		keyErr := closeKeyRepo()
		mediaErr := closeRepo()
		if keyErr != nil {
			return keyErr
		}
		return mediaErr
	}

	return app, cleanup, nil
}

func registerBaseRoutes(app *fiber.App, cfg config.Config, authMiddleware *auth.Middleware) {
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	app.Get("/", authMiddleware.RequireSessionPage, func(c *fiber.Ctx) error {
		return c.Type("html").SendString(web.GalleryHTML())
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})

	app.Get("/info", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"name":    "go-media-api",
			"version": "dev",
			"port":    cfg.Port,
		})
	})
}

func requestBodyLimit(maxUploadBytes int64) int {
	if maxUploadBytes > 0 {
		limit := maxUploadBytes + 1024*1024
		maxInt := int64(int(^uint(0) >> 1))
		if limit > maxInt {
			return int(maxInt)
		}
		return int(limit)
	}
	return int(^uint(0) >> 1)
}
