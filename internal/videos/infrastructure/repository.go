package infrastructure

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	ent "go-starter/internal/shared/infrastructure/ent/generated"
	entvideo "go-starter/internal/shared/infrastructure/ent/generated/video"
	shareddomain "go-starter/internal/shared/domain"
	"go-starter/internal/videos/domain"
)

type VideoRepository struct {
	client *ent.Client
}

func NewVideoRepository(client *ent.Client) *VideoRepository {
	return &VideoRepository{client: client}
}

func (r *VideoRepository) Create(ctx context.Context, v *domain.Video) (*domain.Video, error) {
	created, err := r.client.Video.Create().
		SetID(uuid.MustParse(v.ID.String())).
		SetTitle(v.Title).
		SetDescription(v.Description).
		SetType(entvideo.Type(v.Type.String())).
		SetStatus(entvideo.Status(v.Status.String())).
		SetQualities(qualitiesToStrings(v.Qualities)).
		SetUploadedAt(v.UploadedAt).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create video: %w", err)
	}
	return toDomain(created), nil
}

func (r *VideoRepository) FindByID(ctx context.Context, id string) (*domain.Video, error) {
	v, err := r.client.Video.Query().
		Where(entvideo.IDEQ(uuid.MustParse(id))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("find video by id: %w", err)
	}
	return toDomain(v), nil
}

func (r *VideoRepository) List(ctx context.Context, filter domain.VideoListFilter) ([]*domain.Video, int, error) {
	query := r.client.Video.Query()

	if filter.Search != "" {
		query = query.Where(
			entvideo.Or(
				entvideo.TitleContainsFold(filter.Search),
				entvideo.DescriptionContainsFold(filter.Search),
			),
		)
	}

	if filter.Type != nil {
		query = query.Where(entvideo.TypeEQ(entvideo.Type(filter.Type.String())))
	}

	if filter.Status != nil {
		query = query.Where(entvideo.StatusEQ(entvideo.Status(filter.Status.String())))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count videos: %w", err)
	}

	var fieldToSort string
	switch filter.SortBy {
	case "title":
		fieldToSort = entvideo.FieldTitle
	case "type":
		fieldToSort = entvideo.FieldType
	case "status":
		fieldToSort = entvideo.FieldStatus
	default:
		fieldToSort = entvideo.FieldUploadedAt
	}
	if filter.Order == "asc" {
		query = query.Order(ent.Asc(fieldToSort))
	} else {
		query = query.Order(ent.Desc(fieldToSort))
	}

	offset := (filter.Page - 1) * filter.Limit
	videos, err := query.
		Limit(filter.Limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list videos: %w", err)
	}

	result := make([]*domain.Video, len(videos))
	for i, v := range videos {
		result[i] = toDomain(v)
	}
	return result, total, nil
}

func (r *VideoRepository) Update(ctx context.Context, v *domain.Video) (*domain.Video, error) {
	updated, err := r.client.Video.UpdateOneID(uuid.MustParse(v.ID.String())).
		SetTitle(v.Title).
		SetDescription(v.Description).
		SetType(entvideo.Type(v.Type.String())).
		SetStatus(entvideo.Status(v.Status.String())).
		SetQualities(qualitiesToStrings(v.Qualities)).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("update video: %w", err)
	}
	return toDomain(updated), nil
}

func (r *VideoRepository) Delete(ctx context.Context, id string) error {
	err := r.client.Video.DeleteOneID(uuid.MustParse(id)).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("delete video: %w", err)
	}
	return nil
}

func toDomain(v *ent.Video) *domain.Video {
	return &domain.Video{
		ID:          shareddomain.IdFromStr(v.ID.String()),
		Title:       v.Title,
		Description: v.Description,
		Type:        domain.VideoType(v.Type),
		Status:      domain.VideoStatus(v.Status),
		Qualities:   stringsToQualities(v.Qualities),
		UploadedAt:  v.UploadedAt,
	}
}

func qualitiesToStrings(qualities []domain.VideoQuality) []string {
	res := make([]string, len(qualities))
	for i, q := range qualities {
		res[i] = string(q)
	}
	return res
}

func stringsToQualities(s []string) []domain.VideoQuality {
	res := make([]domain.VideoQuality, len(s))
	for i, v := range s {
		res[i] = domain.VideoQuality(v)
	}
	return res
}
