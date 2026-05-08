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
	Refresh(ctx context.Context, req RefreshRequest) (AuthResponse, error)
	Logout(ctx context.Context, req LogoutRequest) error

	CreateFeedback(ctx context.Context, userID int64, req FeedbackCreateRequest) (FeedbackResponse, error)
	ListMyFeedbacks(ctx context.Context, userID int64) (MyFeedbacksResponse, error)

	GetTeacherByID(ctx context.Context, id int64) (TeacherDetailsResponse, error)

	ListAdminFeedbacks(ctx context.Context, params AdminFeedbackListParams) (AdminFeedbacksResponse, error)
	BlockUser(ctx context.Context, userID int64) error
}

type Repository interface {
	ListTeachers(ctx context.Context, params TeacherListParams) ([]TeacherListItem, int64, error)
	GetTeacherByID(ctx context.Context, id int64) (TeacherDetailsResponse, error)

	CreateUser(ctx context.Context, email, passwordHash, role string) (AuthUser, error)
	GetUserByEmail(ctx context.Context, email string) (AuthUser, error)

	CreateRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (RefreshTokenRecord, error)
	RotateRefreshToken(ctx context.Context, oldTokenID int64, userID int64, newTokenHash string, expiresAt time.Time) error
	RevokeRefreshToken(ctx context.Context, tokenHash string) error

	CreateFeedback(ctx context.Context, userID int64, teacherID int64, rating int, comment string) (FeedbackResponse, error)
	ListFeedbacksByUser(ctx context.Context, userID int64) ([]MyFeedbackItem, error)

	ListAdminFeedbacks(ctx context.Context, params AdminFeedbackListParams) ([]AdminFeedbackItem, int64, error)
	BlockUserAndRevokeTokens(ctx context.Context, userID int64) error
}
