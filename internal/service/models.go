package service

import (
	"time"
)

type HealthResponse struct {
	Status string `json:"status"`
}

type TeacherListParams struct {
	Q       string
	Faculty string
	Limit   int
	Offset  int
}

type TeacherListItem struct {
	ID           int64   `db:"id" json:"id"`
	FullName     string  `db:"full_name" json:"full_name"`
	Faculty      string  `db:"faculty" json:"faculty"`
	ReviewsCount int64   `db:"reviews_count" json:"reviews_count"`
	AvgRating    float64 `db:"avg_rating" json:"avg_rating"`
}

type TeacherListResponse struct {
	Items  []TeacherListItem `json:"items"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

type AuthConfig struct {
	JWTSecret            string
	AdminEmail           string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	BcryptCost           int
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthUser struct {
	ID           int64  `db:"id" json:"id"`
	Email        string `db:"email" json:"email"`
	PasswordHash string `db:"password_hash" json:"-"`
	Role         string `db:"role" json:"role"`
	IsBlocked    bool   `db:"is_blocked" json:"-"`
}

type PublicUser struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type AuthResponse struct {
	User             PublicUser `json:"user"`
	AccessToken      string     `json:"access_token"`
	RefreshToken     string     `json:"refresh_token"`
	AccessExpiresIn  int64      `json:"access_expires_in"`
	RefreshExpiresIn int64      `json:"refresh_expires_in"`
}
