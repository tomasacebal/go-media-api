package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/tomasacebal/go-media-api/internal/auth"
	"github.com/tomasacebal/go-media-api/internal/config"
	"github.com/tomasacebal/go-media-api/internal/database"
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
	defaultQuotaBytes := cfg.Product.DefaultQuotaBytes
	if defaultQuotaBytes <= 0 {
		defaultQuotaBytes = int64(10) * 1024 * 1024 * 1024
	}
	shareTTLDays := cfg.Product.ShareTTLDays
	if shareTTLDays <= 0 {
		shareTTLDays = 30
	}

	db, closeDB, err := database.OpenSQLite(cfg.Media.SQLitePath)
	if err != nil {
		return nil, nil, err
	}

	userRepo := auth.NewSQLiteUserRepository(db)
	userService := auth.NewUserService(userRepo, defaultQuotaBytes, shareTTLDays)
	admin, err := userService.EnsureBootstrapAdmin(
		context.Background(),
		cfg.Auth.AdminEmail,
		cfg.Auth.AdminName,
		cfg.Auth.AdminPassword,
		defaultQuotaBytes,
		shareTTLDays,
	)
	if err != nil {
		_ = closeDB()
		return nil, nil, err
	}
	if err := database.AssignLegacyOwnership(context.Background(), db, admin.ID); err != nil {
		_ = closeDB()
		return nil, nil, err
	}

	repo := media.NewSQLiteRepository(db)
	keyRepo := auth.NewSQLiteKeyRepository(db)

	storageProvider, err := storage.NewLocalProvider(cfg.Media.StoragePath, cfg.Media.PublicBaseURL)
	if err != nil {
		_ = closeDB()
		return nil, nil, err
	}

	app := fiber.New(fiber.Config{
		BodyLimit:    requestBodyLimit(cfg.Media.MaxUploadBytes),
		ErrorHandler: media.JSONErrorHandler,
		ReadTimeout:  15 * time.Minute,
		WriteTimeout: 15 * time.Minute,
		IdleTimeout:  2 * time.Minute,
	})

	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "no-referrer")
		c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://static.cloudflareinsights.com; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data:; connect-src 'self'")
		return c.Next()
	})

	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)
		logger.Printf("%s %s %d %s", c.Method(), c.Path(), c.Response().StatusCode(), duration)

		return err
	})

	keyService := auth.NewAPIKeyService(keyRepo)
	sessionManager := auth.NewSessionManager(userService, cfg.Auth.SessionSecret, cfg.Auth.CookieSecure, time.Duration(cfg.Auth.SessionTTLHours)*time.Hour)
	authMiddleware := auth.NewMiddleware(sessionManager, keyService, userService)
	authHandler := auth.NewHandler(sessionManager, userService, keyService, authMiddleware)
	authHandler.RegisterRoutes(app)
	registerBaseRoutes(app, cfg, authMiddleware)
	mediaService := media.NewService(repo, storageProvider, userService, cfg.Media.MaxUploadBytes, cfg.Media.PublicBaseURL, shareTTLDays, cfg.Media.StoragePath)
	mediaHandler := media.NewHandler(mediaService, logger, authMiddleware)
	mediaHandler.RegisterRoutes(app)

	cleanup := func() error {
		return closeDB()
	}

	return app, cleanup, nil
}

func registerBaseRoutes(app *fiber.App, cfg config.Config, authMiddleware *auth.Middleware) {
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "image/x-icon")
		c.Set(fiber.HeaderCacheControl, "public, max-age=31536000")
		return c.Send(web.SDCardIcon)
	})
	app.Get("/favicon.svg", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "image/svg+xml")
		c.Set(fiber.HeaderCacheControl, "public, max-age=31536000")
		return c.SendString(web.SDCardSVG)
	})

	app.Get("/", authMiddleware.RequireSessionPage, func(c *fiber.Ctx) error {
		return c.Type("html").SendString(web.AppHTML())
	})

	app.Get("/assets/app.css", func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderCacheControl, "public, max-age=300")
		return c.Type("css").SendString(web.AppCSS())
	})

	app.Get("/assets/app.js", func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderCacheControl, "public, max-age=300")
		return c.Type("js").SendString(web.AppJS())
	})

	app.Get("/assets/share.js", func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderCacheControl, "public, max-age=300")
		return c.Type("js").SendString(web.ShareJS())
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})

	app.Get("/info", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"name":    "AceTransfer",
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
