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
	"go-starter/internal/videos"
	videosapp "go-starter/internal/videos/application"
	videosinfra "go-starter/internal/videos/infrastructure"
	videospres "go-starter/internal/videos/presentation"
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
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Requested-With", "Accept", "Origin"},
	}))

	presentation.AuthErrorHandler = authpresentation.AuthErrorHandler
	presentation.UserErrorHandler = userspresentation.UserErrorHandler
	presentation.VideoErrorHandler = videospres.VideoErrorHandler

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

	sharedinfra.SeedAdmin(context.Background(), userRepo, passwordAdapter, cfg.AdminEmail(), cfg.AdminPassword())

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
		Storage:         usersinfra.NewUserStorageAdapter(s3Adapter),
		IDGenerator:     idGen,
	})
	usersModule.RegisterRoutes(v1.Group("/users"), cfg.JWTAccessTokenSecret())

	videoRepo := videosinfra.NewVideoRepository(client)
	videoStorage := videosinfra.NewVideoStorageAdapter(s3Adapter)
	videoQueue := videosinfra.NewInMemoryQueueAdapter()

	videosModule := videos.NewModule(videos.Dependencies{
		VideoRepo:    videoRepo,
		Storage:      videoStorage,
		MessageQueue: videoQueue,
		IDGenerator:  idGen,
	})
	videosModule.RegisterRoutes(v1, cfg.JWTAccessTokenSecret())

	ffmpegAdapter := videosinfra.NewFfmpegAdapter()
	processVideoUseCase := videosapp.NewProcessVideo(videoRepo, videoStorage, ffmpegAdapter)
	consumer := videospres.NewConsumerQueueWorker(videoQueue, processVideoUseCase)
	go consumer.Start(context.Background())

	pollPendingVideos := videosapp.NewPollPendingVideos(videoRepo, videoQueue)
	producerWorker := videospres.NewProducerWorker(pollPendingVideos)
	go producerWorker.Start(context.Background())

	return e
}
