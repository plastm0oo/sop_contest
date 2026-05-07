package service

import (
	"context"
	"net/http"
)

type Handler interface {
	RegisterRoutes(mux *http.ServeMux)
}

type UseCase interface {
	Health(ctx context.Context) HealthResponse
	ListTeachers(ctx context.Context, params TeacherListParams) (TeacherListResponse, error)
}

type Repository interface {
	ListTeachers(ctx context.Context, params TeacherListParams) ([]TeacherListItem, int64, error)
}
