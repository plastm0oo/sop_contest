package service

import (
	"context"
	"net/http"
	"time"
)

type Handler interface {
	RegisterRoutes(mux *http.ServeMux)
}

type UseCase interface {
	Health(ctx context.Context) HealthResponse
	ListTeachers(ctx context.Context, params TeacherListParams) (TeacherListResponse, error)
	Register(ctx context.Context, req RegisterRequest) (AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (AuthResponse, error)
	CreateFeedback(ctx context.Context, userID int64, req FeedbackCreateRequest) (FeedbackResponse, error)
	ListMyFeedbacks(ctx context.Context, userID int64) (MyFeedbacksResponse, error)
}

type Repository interface {
	ListTeachers(ctx context.Context, params TeacherListParams) ([]TeacherListItem, int64, error)
	CreateUser(ctx context.Context, email, passwordHash, role string) (AuthUser, error)
	GetUserByEmail(ctx context.Context, email string) (AuthUser, error)
	CreateRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	CreateFeedback(ctx context.Context, userID int64, teacherID int64, rating int, comment string) (FeedbackResponse, error)
	ListFeedbacksByUser(ctx context.Context, userID int64) ([]MyFeedbackItem, error)
}
