package presentation

import (
	"net/http"

	sharedpres "go-starter/internal/shared/presentation"
	"go-starter/internal/videos/application"

	"github.com/labstack/echo/v4"
)

type VideosHandlers struct {
	createVideoUseCase       *application.CreateVideo
	triggerProcessingUseCase *application.TriggerProcessing
	replaceVideoUseCase      *application.ReplaceVideo
	regenerateQualityUseCase *application.RegenerateQuality
	deleteVideoUseCase       *application.DeleteVideo
	listVideosUseCase        *application.ListVideos
	listCatalogUseCase       *application.ListCatalog
	getVideoStreamUseCase    *application.GetVideoStream
}

func NewVideosHandlers(
	createVideo *application.CreateVideo,
	triggerProcessing *application.TriggerProcessing,
	replaceVideo *application.ReplaceVideo,
	regenerateQuality *application.RegenerateQuality,
	deleteVideo *application.DeleteVideo,
	listVideos *application.ListVideos,
	listCatalog *application.ListCatalog,
	getVideoStream *application.GetVideoStream,
) *VideosHandlers {
	return &VideosHandlers{
		createVideoUseCase:       createVideo,
		triggerProcessingUseCase: triggerProcessing,
		replaceVideoUseCase:      replaceVideo,
		regenerateQualityUseCase: regenerateQuality,
		deleteVideoUseCase:       deleteVideo,
		listVideosUseCase:        listVideos,
		listCatalogUseCase:       listCatalog,
		getVideoStreamUseCase:    getVideoStream,
	}
}

func (h *VideosHandlers) CreateVideo(c echo.Context) error {
	var dto CreateVideoDto
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	result, err := h.createVideoUseCase.Execute(c.Request().Context(), application.CreateVideoInput{
		Title:       dto.Title,
		Description: dto.Description,
		Type:        dto.Type,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, sharedpres.Response[*application.CreateVideoOutput]{
		Message:    "Video created",
		StatusCode: 201,
		Data:       result,
	})
}

func (h *VideosHandlers) TriggerProcessing(c echo.Context) error {
	videoID := c.Param("id")

	var dto TriggerProcessingDto
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}
	if len(dto.RequestedQualities) == 0 {
		dto.RequestedQualities = []string{"480p", "1080p"}
	}

	err := h.triggerProcessingUseCase.Execute(c.Request().Context(), application.TriggerProcessingInput{
		VideoID:            videoID,
		RequestedQualities: dto.RequestedQualities,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusAccepted, sharedpres.Response[any]{
		Message:    "Processing triggered",
		StatusCode: 202,
		Data:       nil,
	})
}

func (h *VideosHandlers) ReplaceVideo(c echo.Context) error {
	videoID := c.Param("id")

	result, err := h.replaceVideoUseCase.Execute(c.Request().Context(), application.ReplaceVideoInput{
		VideoID: videoID,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, sharedpres.Response[*application.ReplaceVideoOutput]{
		Message:    "Video replacement initiated",
		StatusCode: 200,
		Data:       result,
	})
}

func (h *VideosHandlers) RegenerateQuality(c echo.Context) error {
	videoID := c.Param("id")

	var dto RegenerateQualityDto
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	err := h.regenerateQualityUseCase.Execute(c.Request().Context(), application.RegenerateQualityInput{
		VideoID: videoID,
		Quality: dto.Quality,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusAccepted, sharedpres.Response[any]{
		Message:    "Quality regeneration triggered",
		StatusCode: 202,
		Data:       nil,
	})
}

func (h *VideosHandlers) DeleteVideo(c echo.Context) error {
	videoID := c.Param("id")

	err := h.deleteVideoUseCase.Execute(c.Request().Context(), videoID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, sharedpres.Response[any]{
		Message:    "Video deleted",
		StatusCode: 200,
		Data:       nil,
	})
}

func (h *VideosHandlers) ListVideos(c echo.Context) error {
	query := ListVideosQuery{Page: 1, Limit: 20, SortBy: "uploaded_at", Order: "desc"}
	_ = c.Bind(&query)
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > 100 {
		query.Limit = 20
	}

	result, err := h.listVideosUseCase.Execute(c.Request().Context(), application.ListVideosInput{
		Search: query.Search,
		Type:   query.Type,
		Status: query.Status,
		Page:   query.Page,
		Limit:  query.Limit,
		SortBy: query.SortBy,
		Order:  query.Order,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, sharedpres.PaginatedResponse[application.VideoOutput]{
		Message:    "Videos retrieved",
		StatusCode: 200,
		Data:       result.Videos,
		Pagination: sharedpres.PaginationMeta{
			Total: result.Total,
			Page:  result.Page,
			Limit: result.Limit,
		},
	})
}

func (h *VideosHandlers) ListCatalog(c echo.Context) error {
	query := ListCatalogQuery{Page: 1, Limit: 20}
	_ = c.Bind(&query)
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > 100 {
		query.Limit = 20
	}

	result, err := h.listCatalogUseCase.Execute(c.Request().Context(), application.ListCatalogInput{
		Page:  query.Page,
		Limit: query.Limit,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, sharedpres.PaginatedResponse[application.VideoOutput]{
		Message:    "Catalog retrieved",
		StatusCode: 200,
		Data:       result.Videos,
		Pagination: sharedpres.PaginationMeta{
			Total: result.Total,
			Page:  result.Page,
			Limit: result.Limit,
		},
	})
}

func (h *VideosHandlers) GetVideoStream(c echo.Context) error {
	videoID := c.Param("id")

	result, err := h.getVideoStreamUseCase.Execute(c.Request().Context(), videoID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, sharedpres.Response[*application.VideoStreamOutput]{
		Message:    "Video stream retrieved",
		StatusCode: 200,
		Data:       result,
	})
}
