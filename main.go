package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
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

	storageProvider, err := storage.NewLocalProvider(cfg.Media.StoragePath, cfg.Media.PublicBaseURL)
	if err != nil {
		_ = closeRepo()
		return nil, nil, err
	}

	app := fiber.New(fiber.Config{
		BodyLimit:    int(cfg.Media.MaxUploadBytes + 1024*1024),
		ErrorHandler: media.JSONErrorHandler,
	})

	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)
		logger.Printf("%s %s %d %s", c.Method(), c.Path(), c.Response().StatusCode(), duration)

		return err
	})

	registerBaseRoutes(app, cfg)

	mediaService := media.NewService(repo, storageProvider, cfg.Media.MaxUploadBytes, cfg.Media.PublicBaseURL)
	mediaHandler := media.NewHandler(mediaService, logger)
	mediaHandler.RegisterRoutes(app)

	return app, closeRepo, nil
}

func registerBaseRoutes(app *fiber.App, cfg config.Config) {
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	app.Get("/", func(c *fiber.Ctx) error {
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
