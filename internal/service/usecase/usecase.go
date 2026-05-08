package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/plastm0oo/sop_contest/internal/auth"
	"github.com/plastm0oo/sop_contest/internal/service"
)

//var ErrEmptyMessage = errors.New("message must not be empty")

type useCase struct {
	repo    service.Repository
	authCfg service.AuthConfig
}

func New(repo service.Repository, authCfg service.AuthConfig) service.UseCase {
	return &useCase{
		repo:    repo,
		authCfg: authCfg,
	}
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

func (uc *useCase) Register(ctx context.Context, req service.RegisterRequest) (service.AuthResponse, error) {
	email := normalizeEmail(req.Email)

	if details := validateRegisterInput(email, req.Password); len(details) > 0 {
		return service.AuthResponse{}, service.ValidationError{Details: details}
	}

	role := "user"
	if uc.authCfg.AdminEmail != "" && email == normalizeEmail(uc.authCfg.AdminEmail) {
		role = "admin"
	}

	passwordHash, err := auth.HashPassword(req.Password, uc.authCfg.BcryptCost)
	if err != nil {
		return service.AuthResponse{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := uc.repo.CreateUser(ctx, email, passwordHash, role)
	if err != nil {
		return service.AuthResponse{}, err
	}

	return uc.issueTokenPair(ctx, user)
}

func (uc *useCase) Login(ctx context.Context, req service.LoginRequest) (service.AuthResponse, error) {
	email := normalizeEmail(req.Email)

	if email == "" || req.Password == "" {
		return service.AuthResponse{}, service.ErrInvalidCredentials
	}

	user, err := uc.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return service.AuthResponse{}, service.ErrInvalidCredentials
		}

		return service.AuthResponse{}, err
	}

	if user.IsBlocked {
		return service.AuthResponse{}, service.ErrAccountBlocked
	}

	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		return service.AuthResponse{}, service.ErrInvalidCredentials
	}

	return uc.issueTokenPair(ctx, user)
}

func (uc *useCase) issueTokenPair(ctx context.Context, user service.AuthUser) (service.AuthResponse, error) {
	accessToken, err := auth.GenerateAccessToken(
		user.ID,
		user.Email,
		user.Role,
		uc.authCfg.JWTSecret,
		uc.authCfg.AccessTokenDuration,
	)
	if err != nil {
		return service.AuthResponse{}, err
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return service.AuthResponse{}, err
	}

	refreshTokenHash := auth.HashRefreshToken(refreshToken)
	refreshExpiresAt := now().Add(uc.authCfg.RefreshTokenDuration)

	if err := uc.repo.CreateRefreshToken(ctx, user.ID, refreshTokenHash, refreshExpiresAt); err != nil {
		return service.AuthResponse{}, err
	}

	return service.AuthResponse{
		User: service.PublicUser{
			ID:    user.ID,
			Email: user.Email,
			Role:  user.Role,
		},
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresIn:  int64(uc.authCfg.AccessTokenDuration.Seconds()),
		RefreshExpiresIn: int64(uc.authCfg.RefreshTokenDuration.Seconds()),
	}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validateRegisterInput(email, password string) map[string]string {
	details := make(map[string]string)

	if email == "" || !strings.Contains(email, "@") {
		details["email"] = "must be a valid email"
	}

	if err := validatePassword(password); err != "" {
		details["password"] = err
	}

	return details
}

func validatePassword(password string) string {
	if len(password) < 8 {
		return "минимум 8 символов, одна буква и одна цифра"
	}

	hasLetter := false
	hasDigit := false

	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}

		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}

	if !hasLetter || !hasDigit {
		return "минимум 8 символов, одна буква и одна цифра"
	}

	return ""
}

var now = func() time.Time {
	return time.Now()
}

func (uc *useCase) CreateFeedback(
	ctx context.Context,
	userID int64,
	req service.FeedbackCreateRequest,
) (service.FeedbackResponse, error) {
	req.Comment = strings.TrimSpace(req.Comment)

	if details := validateFeedbackInput(req); len(details) > 0 {
		return service.FeedbackResponse{}, service.ValidationError{Details: details}
	}

	feedback, err := uc.repo.CreateFeedback(ctx, userID, req.TeacherID, req.Rating, req.Comment)
	if err != nil {
		return service.FeedbackResponse{}, err
	}

	return feedback, nil
}

func (uc *useCase) ListMyFeedbacks(ctx context.Context, userID int64) (service.MyFeedbacksResponse, error) {
	items, err := uc.repo.ListFeedbacksByUser(ctx, userID)
	if err != nil {
		return service.MyFeedbacksResponse{}, err
	}

	if items == nil {
		items = make([]service.MyFeedbackItem, 0)
	}

	return service.MyFeedbacksResponse{Items: items}, nil
}

func validateFeedbackInput(req service.FeedbackCreateRequest) map[string]string {
	details := make(map[string]string)

	if req.TeacherID <= 0 {
		details["teacher_id"] = "must be a positive integer"
	}

	if req.Rating < 1 || req.Rating > 5 {
		details["rating"] = "must be between 1 and 5"
	}

	commentLen := utf8.RuneCountInString(req.Comment)
	if commentLen < 10 || commentLen > 2000 {
		details["comment"] = "length must be between 10 and 2000 characters"
	}

	return details
}
