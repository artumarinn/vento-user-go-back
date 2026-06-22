package usecase

import (
	"context"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type TagUsecases struct {
	repo port.TagRepository
}

func NewTagUsecases(repo port.TagRepository) *TagUsecases {
	return &TagUsecases{repo: repo}
}

func (uc *TagUsecases) ListTags(ctx context.Context, userID string) ([]dto.TagResponse, error) {
	tags, err := uc.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := make([]dto.TagResponse, len(tags))
	for i, tag := range tags {
		res[i] = mapTagEntityToDTO(tag)
	}
	return res, nil
}

func (uc *TagUsecases) CreateTag(ctx context.Context, userID string, req dto.CreateTagRequest) (dto.TagResponse, error) {
	color := req.Color
	if color == "" {
		color = "#9CA3AF"
	}

	t := entity.NewTag(userID, req.Label, color)

	if err := uc.repo.Save(ctx, t); err != nil {
		return dto.TagResponse{}, err
	}

	return mapTagEntityToDTO(t), nil
}

func (uc *TagUsecases) DeleteTag(ctx context.Context, userID string, tagID string) error {
	return uc.repo.Delete(ctx, tagID, userID)
}

func mapTagEntityToDTO(t *entity.Tag) dto.TagResponse {
	return dto.TagResponse{
		ID:        t.ID,
		Label:     t.Label,
		Color:     t.Color,
		CreatedAt: t.CreatedAt,
	}
}
