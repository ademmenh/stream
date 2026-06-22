package application

import (
	"context"
	"time"

	shareddomain "go-starter/internal/shared/domain"
	"go-starter/internal/videos/domain"
)

type VideoOutput struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Type         string   `json:"type"`
	Status       string   `json:"status"`
	Qualities    []string `json:"qualities"`
	ThumbnailUrl string   `json:"thumbnail_url"`
	UploadedAt   string   `json:"uploaded_at"`
}

type ListVideosInput struct {
	Search string
	Type   *string
	Status *string
	Page   int
	Limit  int
	SortBy string
	Order  string
}

type ListVideosOutput struct {
	Videos []VideoOutput `json:"videos"`
	Total  int           `json:"total"`
	Page   int           `json:"page"`
	Limit  int           `json:"limit"`
}

type ListVideos struct {
	videoRepo domain.IVideoRepository
	storage   domain.IStorageAdapter
}

func NewListVideos(
	videoRepo domain.IVideoRepository,
	storage domain.IStorageAdapter,
) *ListVideos {
	return &ListVideos{
		videoRepo: videoRepo,
		storage:   storage,
	}
}

func (uc *ListVideos) Execute(ctx context.Context, input ListVideosInput) (*ListVideosOutput, error) {
	pagination := shareddomain.NewPaginationParams(input.Page, input.Limit)

	filter := domain.VideoListFilter{
		Search: input.Search,
		Page:   pagination.Page,
		Limit:  pagination.Limit,
		SortBy: input.SortBy,
		Order:  input.Order,
	}

	if input.Type != nil {
		t := domain.VideoType(*input.Type)
		filter.Type = &t
	}
	if input.Status != nil {
		s := domain.VideoStatus(*input.Status)
		filter.Status = &s
	}

	videos, total, err := uc.videoRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	output := make([]VideoOutput, len(videos))
	for i, v := range videos {
		thumbnailKey := "photos/" + v.GetID() + "/thumbnail.jpg"
		thumbnailUrl, _ := uc.storage.GeneratePresignedGetUrl(ctx, thumbnailKey, 1*time.Hour)

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

	return &ListVideosOutput{
		Videos: output,
		Total:  total,
		Page:   pagination.Page,
		Limit:  pagination.Limit,
	}, nil
}

func qualitiesToStrings(qualities []domain.VideoQuality) []string {
	res := make([]string, len(qualities))
	for i, q := range qualities {
		res[i] = string(q)
	}
	return res
}
