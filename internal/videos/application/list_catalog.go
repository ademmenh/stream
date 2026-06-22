package application

import (
	"context"
	"time"

	shareddomain "go-starter/internal/shared/domain"
	"go-starter/internal/videos/domain"
)

type ListCatalogInput struct {
	Page  int
	Limit int
}

type ListCatalogOutput struct {
	Videos []VideoOutput `json:"videos"`
	Total  int           `json:"total"`
	Page   int           `json:"page"`
	Limit  int           `json:"limit"`
}

type ListCatalog struct {
	videoRepo domain.IVideoRepository
	storage   domain.IStorageAdapter
}

func NewListCatalog(
	videoRepo domain.IVideoRepository,
	storage domain.IStorageAdapter,
) *ListCatalog {
	return &ListCatalog{
		videoRepo: videoRepo,
		storage:   storage,
	}
}

func (uc *ListCatalog) Execute(ctx context.Context, input ListCatalogInput) (*ListCatalogOutput, error) {
	pagination := shareddomain.NewPaginationParams(input.Page, input.Limit)

	readyStatus := domain.StatusReady
	filter := domain.VideoListFilter{
		Page:   pagination.Page,
		Limit:  pagination.Limit,
		Status: &readyStatus,
	}

	videos, total, err := uc.videoRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	output := make([]VideoOutput, len(videos))
	thumbExpiry := 1 * time.Hour
	for i, v := range videos {
		thumbnailKey := "photos/" + v.GetID() + "/thumbnail.jpg"
		thumbnailUrl, _ := uc.storage.GeneratePresignedGetUrl(ctx, thumbnailKey, thumbExpiry)

		output[i] = VideoOutput{
			ID:           v.GetID(),
			Title:        v.GetTitle(),
			Description:  v.GetDescription(),
			Type:         v.GetType().String(),
			Status:       v.GetStatus().String(),
			Qualities:    qualitiesToStrings(v.GetQualities()),
			ThumbnailUrl: thumbnailUrl,
			UploadedAt:   v.GetUploadedAt().Format(time.RFC3339),
		}
	}

	return &ListCatalogOutput{
		Videos: output,
		Total:  total,
		Page:   pagination.Page,
		Limit:  pagination.Limit,
	}, nil
}
