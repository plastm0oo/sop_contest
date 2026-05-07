package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
