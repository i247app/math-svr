package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"math-ai.com/math-ai/internal/infrastructure/logger"
)

const migrationTable = `CREATE TABLE IF NOT EXISTS schema_migrations (
	version VARCHAR(255) NOT NULL PRIMARY KEY,
	applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`

// Migrate reads .sql files from migrationsDir and applies any that haven't been run yet.
// Files are sorted lexicographically; use a naming convention like 001_create_users.sql.
func Migrate(ctx context.Context, db *sql.DB, migrationsDir string) error {
	lg := logger.From(ctx)

	// Ensure migrations tracking table exists
	if _, err := db.ExecContext(ctx, migrationTable); err != nil {
		return fmt.Errorf("migration: create tracking table: %w", err)
	}

	// Read migration files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("migration: read dir %s: %w", migrationsDir, err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	// Get already-applied versions
	applied, err := getAppliedMigrations(ctx, db)
	if err != nil {
		return err
	}

	// Apply pending migrations
	for _, f := range files {
		version := strings.TrimSuffix(f, ".sql")
		if applied[version] {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			return fmt.Errorf("migration: read file %s: %w", f, err)
		}

		lg.Info("[MIGRATE] Applying", "version", version)

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("migration: begin tx for %s: %w", version, err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration: execute %s: %w", version, err)
		}

		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration: record %s: %w", version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("migration: commit %s: %w", version, err)
		}

		lg.Info("[MIGRATE] Applied", "version", version)
	}

	return nil
}

func getAppliedMigrations(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("migration: query applied: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("migration: scan version: %w", err)
		}
		applied[v] = true
	}
	return applied, rows.Err()
}
