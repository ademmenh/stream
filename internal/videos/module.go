package videos

import (
	"github.com/labstack/echo/v4"

	sharedpres "go-starter/internal/shared/presentation"
	"go-starter/internal/videos/application"
	"go-starter/internal/videos/domain"
	"go-starter/internal/videos/presentation"
)

type Module struct {
	handlers *presentation.VideosHandlers
}

type Dependencies struct {
	VideoRepo    domain.IVideoRepository
	Storage      domain.IStorageAdapter
	MessageQueue domain.IMessageQueue
	IDGenerator  application.IDGenerator
}

const (
	defaultRawUploadExpiry   = 60
	defaultPhotoUploadExpiry = 15
)

func NewModule(deps Dependencies) *Module {
	createVideoUseCase := application.NewCreateVideo(
		deps.VideoRepo,
		deps.Storage,
		deps.IDGenerator,
		defaultRawUploadExpiry,
		defaultPhotoUploadExpiry,
	)
	triggerProcessingUseCase := application.NewTriggerProcessing(
		deps.VideoRepo,
		deps.MessageQueue,
	)
	replaceVideoUseCase := application.NewReplaceVideo(
		deps.VideoRepo,
		deps.Storage,
		defaultRawUploadExpiry,
	)
	regenerateQualityUseCase := application.NewRegenerateQuality(
		deps.VideoRepo,
		deps.Storage,
		deps.MessageQueue,
	)
	deleteVideoUseCase := application.NewDeleteVideo(
		deps.VideoRepo,
		deps.Storage,
	)
	listVideosUseCase := application.NewListVideos(
		deps.VideoRepo,
		deps.Storage,
	)
	listCatalogUseCase := application.NewListCatalog(
		deps.VideoRepo,
		deps.Storage,
	)
	getVideoStreamUseCase := application.NewGetVideoStream(
		deps.VideoRepo,
		deps.Storage,
	)

	handlers := presentation.NewVideosHandlers(
		createVideoUseCase,
		triggerProcessingUseCase,
		replaceVideoUseCase,
		regenerateQualityUseCase,
		deleteVideoUseCase,
		listVideosUseCase,
		listCatalogUseCase,
		getVideoStreamUseCase,
	)

	return &Module{handlers: handlers}
}

func (m *Module) RegisterRoutes(group *echo.Group, jwtSecret string) {
	admin := group.Group("/admin", sharedpres.AuthMiddleware(jwtSecret), sharedpres.RoleGuard("admin"))
	admin.POST("/videos", m.handlers.CreateVideo)
	admin.POST("/videos/:id/process", m.handlers.TriggerProcessing)
	admin.PUT("/videos/:id/replace", m.handlers.ReplaceVideo)
	admin.POST("/videos/:id/regenerate", m.handlers.RegenerateQuality)
	admin.DELETE("/videos/:id", m.handlers.DeleteVideo)
	admin.GET("/videos", m.handlers.ListVideos)

	videos := group.Group("/videos")
	videos.GET("", m.handlers.ListCatalog)
	videos.GET("/:id", m.handlers.GetVideoStream)
}
