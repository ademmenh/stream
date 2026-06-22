package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"go-starter/internal/auth"
	authinfra "go-starter/internal/auth/infrastructure"
	authpresentation "go-starter/internal/auth/presentation"
	config "go-starter/internal/config/domain"

	sharedinfra "go-starter/internal/shared/infrastructure"
	presentation "go-starter/internal/shared/presentation"
	"go-starter/internal/users"
	usersinfra "go-starter/internal/users/infrastructure"
	userspresentation "go-starter/internal/users/presentation"
)

func CreateApp(cfg config.IConfig) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.HTTPErrorHandler = presentation.CustomHTTPErrorHandler

	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.CORSOrigins(),
		AllowCredentials: cfg.CORSCredentials(),
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"*"},
	}))

	presentation.AuthErrorHandler = authpresentation.AuthErrorHandler
	presentation.UserErrorHandler = userspresentation.UserErrorHandler

	sharedinfra.InitLogger(cfg.LogsDirname())

	client, err := sharedinfra.NewDB(cfg.DatabaseURL())
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return nil
	}

	s3Adapter, err := sharedinfra.NewS3Adapter(cfg)
	if err != nil {
		slog.Warn("S3 adapter unavailable, continuing without it", "error", err)
	} else {
		if err := s3Adapter.Init(context.Background()); err != nil {
			slog.Warn("S3 adapter init failed, continuing without it", "error", err)
		}
	}
	_ = s3Adapter

	jwtAdapter := authinfra.NewJwtAdapter(cfg)
	passwordAdapter := authinfra.NewPasswordAdapter()
	userRepo := usersinfra.NewUserRepository(client)
	idGen := usersinfra.NewIDGenerator()

	apiPrefix := fmt.Sprintf("/api/v%s", cfg.APIVersion())
	v1 := e.Group(apiPrefix)

	health := v1.Group("/health")
	health.GET("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{
			"message": "Service Healthy",
			"data":    "ok",
		})
	})

	authModule := auth.NewModule(auth.Dependencies{
		JwtAdapter:      jwtAdapter,
		PasswordAdapter: passwordAdapter,
		IDGenerator:     idGen,
		UserRepo:        userRepo,
		Config:          cfg,
	})
	authModule.RegisterRoutes(v1.Group("/auth"))

	usersModule := users.NewModule(users.Dependencies{
		UserRepo:        userRepo,
		PasswordAdapter: passwordAdapter,
	})
	usersModule.RegisterRoutes(v1.Group("/users"), cfg.JWTAccessTokenSecret())

	return e
}
