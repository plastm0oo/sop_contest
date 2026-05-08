package migrator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

func Run(ctx context.Context, db *sqlx.DB, dir string) error {
	if err := ensureSchemaMigrationsTable(ctx, db); err != nil {
		return err
	}

	files, err := migrationFiles(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		name := filepath.Base(file)

		applied, err := isApplied(ctx, db, name)
		if err != nil {
			return err
		}

		if applied {
			continue
		}

		if err := applyMigration(ctx, db, file, name); err != nil {
			return err
		}
	}

	return nil
}

func ensureSchemaMigrationsTable(ctx context.Context, db *sqlx.DB) error {
	const query = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
	`

	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	return nil
}

func migrationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations directory %q: %w", dir, err)
	}

	files := make([]string, 0)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		if strings.HasSuffix(name, "_down.sql") {
			continue
		}

		files = append(files, filepath.Join(dir, name))
	}

	sort.Strings(files)

	return files, nil
}

func isApplied(ctx context.Context, db *sqlx.DB, version string) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM schema_migrations
			WHERE version = $1
		);
	`

	var exists bool

	if err := db.GetContext(ctx, &exists, query, version); err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}

	return exists, nil
}

func applyMigration(ctx context.Context, db *sqlx.DB, file string, version string) error {
	sqlBytes, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", version, err)
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", version, err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if _, err := tx.ExecContext(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("execute migration %s: %w", version, err)
	}

	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(version) VALUES ($1);`, version); err != nil {
		return fmt.Errorf("record migration %s: %w", version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", version, err)
	}

	committed = true
	return nil
}
