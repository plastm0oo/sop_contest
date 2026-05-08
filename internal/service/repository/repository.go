package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	//"log"

	"github.com/plastm0oo/sop_contest/internal/service"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type repo struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) service.Repository {
	return &repo{db: db}
}

func (r *repo) ListTeachers(ctx context.Context, params service.TeacherListParams) ([]service.TeacherListItem, int64, error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM teachers t
		WHERE ($1 = '' OR t.full_name ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR t.faculty = $2);
	`

	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, params.Q, params.Faculty); err != nil {
		return nil, 0, fmt.Errorf("count teachers: %w", err)
	}

	const listQuery = `
		SELECT
			t.id,
			t.full_name,
			t.faculty,
			COUNT(f.id) AS reviews_count,
			COALESCE(ROUND(AVG(f.rating)::numeric, 1)::float8, 0) AS avg_rating
		FROM teachers t
		LEFT JOIN feedbacks f ON f.teacher_id = t.id
		WHERE ($1 = '' OR t.full_name ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR t.faculty = $2)
		GROUP BY t.id, t.full_name, t.faculty
		ORDER BY t.full_name ASC
		LIMIT $3 OFFSET $4;
	`

	var items []service.TeacherListItem
	if err := r.db.SelectContext(ctx, &items, listQuery, params.Q, params.Faculty, params.Limit, params.Offset); err != nil {
		return nil, 0, fmt.Errorf("list teachers: %w", err)
	}

	return items, total, nil
}

func (r *repo) CreateUser(ctx context.Context, email, passwordHash, role string) (service.AuthUser, error) {
	const query = `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, email, password_hash, role, is_blocked;
	`

	var user service.AuthUser

	err := r.db.GetContext(ctx, &user, query, email, passwordHash, role)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return service.AuthUser{}, service.ErrEmailAlreadyExists
		}

		return service.AuthUser{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (r *repo) GetUserByEmail(ctx context.Context, email string) (service.AuthUser, error) {
	const query = `
		SELECT id, email, password_hash, role, is_blocked
		FROM users
		WHERE email = $1;
	`

	var user service.AuthUser

	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return service.AuthUser{}, service.ErrNotFound
		}

		return service.AuthUser{}, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (r *repo) CreateRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	const query = `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3);
	`

	if _, err := r.db.ExecContext(ctx, query, userID, tokenHash, expiresAt); err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}

	return nil
}

func (r *repo) CreateFeedback(
	ctx context.Context,
	userID int64,
	teacherID int64,
	rating int,
	comment string,
) (service.FeedbackResponse, error) {
	const query = `
		INSERT INTO feedbacks (teacher_id, user_id, rating, comment)
		VALUES ($1, $2, $3, $4)
		RETURNING id, teacher_id, rating, comment, created_at;
	`

	var feedback service.FeedbackResponse

	err := r.db.GetContext(ctx, &feedback, query, teacherID, userID, rating, comment)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505":
				return service.FeedbackResponse{}, service.ErrFeedbackAlreadyExists
			case "23503":
				return service.FeedbackResponse{}, service.ErrTeacherNotFound
			}
		}

		return service.FeedbackResponse{}, fmt.Errorf("create feedback: %w", err)
	}

	return feedback, nil
}

func (r *repo) ListFeedbacksByUser(ctx context.Context, userID int64) ([]service.MyFeedbackItem, error) {
	const query = `
		SELECT
			f.id,
			f.teacher_id,
			t.full_name AS teacher_name,
			f.rating,
			f.comment,
			f.created_at
		FROM feedbacks f
		JOIN teachers t ON t.id = f.teacher_id
		WHERE f.user_id = $1
		ORDER BY f.created_at DESC;
	`

	items := make([]service.MyFeedbackItem, 0)

	if err := r.db.SelectContext(ctx, &items, query, userID); err != nil {
		return nil, fmt.Errorf("list feedbacks by user: %w", err)
	}

	return items, nil
}

func (r *repo) GetTeacherByID(ctx context.Context, id int64) (service.TeacherDetailsResponse, error) {
	const teacherQuery = `
		SELECT
			t.id,
			t.full_name,
			t.faculty,
			COALESCE(t.email, '') AS email,
			COUNT(f.id) AS reviews_count,
			COALESCE(ROUND(AVG(f.rating)::numeric, 1)::float8, 0) AS avg_rating
		FROM teachers t
		LEFT JOIN feedbacks f ON f.teacher_id = t.id
		WHERE t.id = $1
		GROUP BY t.id, t.full_name, t.faculty, t.email;
	`

	var row service.TeacherDetailsRow

	if err := r.db.GetContext(ctx, &row, teacherQuery, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return service.TeacherDetailsResponse{}, service.ErrTeacherNotFound
		}

		return service.TeacherDetailsResponse{}, fmt.Errorf("get teacher by id: %w", err)
	}

	const distributionQuery = `
		SELECT rating, COUNT(*) AS count
		FROM feedbacks
		WHERE teacher_id = $1
		GROUP BY rating;
	`

	var distributionRows []service.RatingDistributionRow

	if err := r.db.SelectContext(ctx, &distributionRows, distributionQuery, id); err != nil {
		return service.TeacherDetailsResponse{}, fmt.Errorf("get teacher rating distribution: %w", err)
	}

	distribution := map[string]int64{
		"1": 0,
		"2": 0,
		"3": 0,
		"4": 0,
		"5": 0,
	}

	for _, item := range distributionRows {
		distribution[strconv.Itoa(item.Rating)] = item.Count
	}

	return service.TeacherDetailsResponse{
		ID:                 row.ID,
		FullName:           row.FullName,
		Faculty:            row.Faculty,
		Email:              row.Email,
		ReviewsCount:       row.ReviewsCount,
		AvgRating:          row.AvgRating,
		RatingDistribution: distribution,
	}, nil
}

func (r *repo) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (service.RefreshTokenRecord, error) {
	const query = `
		SELECT
			rt.id,
			rt.user_id,
			u.email,
			u.role,
			u.is_blocked,
			rt.expires_at,
			rt.revoked_at
		FROM refresh_tokens rt
		JOIN users u ON u.id = rt.user_id
		WHERE rt.token_hash = $1;
	`

	var record service.RefreshTokenRecord

	if err := r.db.GetContext(ctx, &record, query, tokenHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return service.RefreshTokenRecord{}, service.ErrNotFound
		}

		return service.RefreshTokenRecord{}, fmt.Errorf("get refresh token by hash: %w", err)
	}

	return record, nil
}

func (r *repo) RotateRefreshToken(
	ctx context.Context,
	oldTokenID int64,
	userID int64,
	newTokenHash string,
	expiresAt time.Time,
) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin rotate refresh token tx: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	updateResult, err := tx.ExecContext(
		ctx,
		`
			UPDATE refresh_tokens
			SET revoked_at = now()
			WHERE id = $1
			  AND user_id = $2
			  AND revoked_at IS NULL;
		`,
		oldTokenID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("revoke old refresh token: %w", err)
	}

	rowsAffected, err := updateResult.RowsAffected()
	if err != nil {
		return fmt.Errorf("check revoked refresh token rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return service.ErrInvalidRefreshToken
	}

	_, err = tx.ExecContext(
		ctx,
		`
			INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
			VALUES ($1, $2, $3);
		`,
		userID,
		newTokenHash,
		expiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert rotated refresh token: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit rotate refresh token tx: %w", err)
	}

	return nil
}

func (r *repo) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(
		ctx,
		`
			UPDATE refresh_tokens
			SET revoked_at = now()
			WHERE token_hash = $1
			  AND revoked_at IS NULL;
		`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}

	return nil
}

func (r *repo) ListAdminFeedbacks(
	ctx context.Context,
	params service.AdminFeedbackListParams,
) ([]service.AdminFeedbackItem, int64, error) {
	where := " WHERE 1 = 1"
	args := make([]any, 0)

	if params.UserID != nil {
		args = append(args, *params.UserID)
		where += fmt.Sprintf(" AND f.user_id = $%d", len(args))
	}

	if params.TeacherID != nil {
		args = append(args, *params.TeacherID)
		where += fmt.Sprintf(" AND f.teacher_id = $%d", len(args))
	}

	countQuery := `
		SELECT COUNT(*)
		FROM feedbacks f
		JOIN users u ON u.id = f.user_id
		JOIN teachers t ON t.id = f.teacher_id
	` + where

	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("count admin feedbacks: %w", err)
	}

	args = append(args, params.Limit)
	limitPlaceholder := len(args)

	args = append(args, params.Offset)
	offsetPlaceholder := len(args)

	listQuery := `
		SELECT
			f.id,
			f.teacher_id,
			t.full_name AS teacher_name,
			f.user_id,
			u.email AS user_email,
			f.rating,
			f.comment,
			f.created_at
		FROM feedbacks f
		JOIN users u ON u.id = f.user_id
		JOIN teachers t ON t.id = f.teacher_id
	` + where + fmt.Sprintf(`
		ORDER BY f.created_at DESC
		LIMIT $%d OFFSET $%d;
	`, limitPlaceholder, offsetPlaceholder)

	items := make([]service.AdminFeedbackItem, 0)

	if err := r.db.SelectContext(ctx, &items, listQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("list admin feedbacks: %w", err)
	}

	return items, total, nil
}

func (r *repo) BlockUserAndRevokeTokens(ctx context.Context, userID int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin block user tx: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(
		ctx,
		`
			UPDATE users
			SET is_blocked = true
			WHERE id = $1;
		`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("block user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check block user rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return service.ErrUserNotFound
	}

	_, err = tx.ExecContext(
		ctx,
		`
			UPDATE refresh_tokens
			SET revoked_at = now()
			WHERE user_id = $1
			  AND revoked_at IS NULL;
		`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("revoke blocked user refresh tokens: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit block user tx: %w", err)
	}

	return nil
}
