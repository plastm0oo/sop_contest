package repository

import (
	"context"
	"fmt"

	//"log"

	"github.com/plastm0oo/sop_contest/internal/service"

	"github.com/jmoiron/sqlx"
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
