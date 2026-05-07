package usecase

import (
	"context"
	//"errors"
	"fmt"

	"github.com/plastm0oo/sop_contest/internal/service"
)

//var ErrEmptyMessage = errors.New("message must not be empty")

type useCase struct {
	repo service.Repository
}

func New(repo service.Repository) service.UseCase {
	return &useCase{repo: repo}
}

func (uc *useCase) Health(ctx context.Context) service.HealthResponse {
	return service.HealthResponse{Status: "ok"}
}

func (uc *useCase) ListTeachers(ctx context.Context, params service.TeacherListParams) (service.TeacherListResponse, error) {
	items, total, err := uc.repo.ListTeachers(ctx, params)
	if err != nil {
		return service.TeacherListResponse{}, fmt.Errorf("list teachers: %w", err)
	}

	return service.TeacherListResponse{
		Items:  items,
		Total:  total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}
