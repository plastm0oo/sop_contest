package service

import (
	"database/sql"
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

type FeedbackCreateRequest struct {
	TeacherID int64  `json:"teacher_id"`
	Rating    int    `json:"rating"`
	Comment   string `json:"comment"`
}

type FeedbackResponse struct {
	ID        int64     `db:"id" json:"id"`
	TeacherID int64     `db:"teacher_id" json:"teacher_id"`
	Rating    int       `db:"rating" json:"rating"`
	Comment   string    `db:"comment" json:"comment"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type MyFeedbackItem struct {
	ID          int64     `db:"id" json:"id"`
	TeacherID   int64     `db:"teacher_id" json:"teacher_id"`
	TeacherName string    `db:"teacher_name" json:"teacher_name"`
	Rating      int       `db:"rating" json:"rating"`
	Comment     string    `db:"comment" json:"comment"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type MyFeedbacksResponse struct {
	Items []MyFeedbackItem `json:"items"`
}

type TeacherDetailsResponse struct {
	ID                 int64            `json:"id"`
	FullName           string           `json:"full_name"`
	Faculty            string           `json:"faculty"`
	Email              string           `json:"email"`
	ReviewsCount       int64            `json:"reviews_count"`
	AvgRating          float64          `json:"avg_rating"`
	RatingDistribution map[string]int64 `json:"rating_distribution"`
}

type TeacherDetailsRow struct {
	ID           int64   `db:"id"`
	FullName     string  `db:"full_name"`
	Faculty      string  `db:"faculty"`
	Email        string  `db:"email"`
	ReviewsCount int64   `db:"reviews_count"`
	AvgRating    float64 `db:"avg_rating"`
}

type RatingDistributionRow struct {
	Rating int   `db:"rating"`
	Count  int64 `db:"count"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRecord struct {
	ID        int64        `db:"id"`
	UserID    int64        `db:"user_id"`
	Email     string       `db:"email"`
	Role      string       `db:"role"`
	IsBlocked bool         `db:"is_blocked"`
	ExpiresAt time.Time    `db:"expires_at"`
	RevokedAt sql.NullTime `db:"revoked_at"`
}

type AdminFeedbackListParams struct {
	UserID    *int64
	TeacherID *int64
	Limit     int
	Offset    int
}

type AdminFeedbackItem struct {
	ID          int64     `db:"id" json:"id"`
	TeacherID   int64     `db:"teacher_id" json:"teacher_id"`
	TeacherName string    `db:"teacher_name" json:"teacher_name"`
	UserID      int64     `db:"user_id" json:"user_id"`
	UserEmail   string    `db:"user_email" json:"user_email"`
	Rating      int       `db:"rating" json:"rating"`
	Comment     string    `db:"comment" json:"comment"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type AdminFeedbacksResponse struct {
	Items  []AdminFeedbackItem `json:"items"`
	Total  int64               `json:"total"`
	Limit  int                 `json:"limit"`
	Offset int                 `json:"offset"`
}
