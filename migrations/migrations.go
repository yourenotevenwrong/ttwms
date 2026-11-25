// Package migrations applies the embedded PostgreSQL schema in filename order.
package migrations

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"fmt"
)

//go:embed *.sql
var files embed.FS

// Check verifies schema compatibility without changing the database. Staging
// and production run migrations as a separate deployment step, allowing the
// application itself to use credentials without schema-changing privileges.
func Check(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, `SELECT name,checksum FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("read schema history; run migrations first: %w", err)
	}
	applied := make(map[string]string)
	for rows.Next() {
		var name, checksum string
		if err := rows.Scan(&name, &checksum); err != nil {
			rows.Close()
			return err
		}
		applied[name] = checksum
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	entries, err := files.ReadDir(".")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		body, err := files.ReadFile(entry.Name())
		if err != nil {
			return err
		}
		checksum, exists := applied[entry.Name()]
		if !exists {
			return fmt.Errorf("migration %s is pending; run -migrate first", entry.Name())
		}
		if checksum != fmt.Sprintf("%x", sha256.Sum256(body)) {
			return fmt.Errorf("applied migration %s has changed", entry.Name())
		}
		delete(applied, entry.Name())
	}
	if len(applied) != 0 {
		return fmt.Errorf("database contains migrations unavailable in this build")
	}
	return nil
}

// Apply serializes concurrent migrators and commits all pending migrations
// atomically. Applied SQL files must never be edited; add a new migration instead.
func Apply(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(816223492)`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		name text PRIMARY KEY, checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now()
	)`); err != nil {
		return err
	}
	entries, err := files.ReadDir(".")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		body, err := files.ReadFile(entry.Name())
		if err != nil {
			return err
		}
		checksum := fmt.Sprintf("%x", sha256.Sum256(body))
		var applied string
		err = tx.QueryRowContext(ctx, `SELECT checksum FROM schema_migrations WHERE name=$1`, entry.Name()).Scan(&applied)
		if err == nil {
			if applied != checksum {
				return fmt.Errorf("applied migration %s has changed", entry.Name())
			}
			continue
		}
		if err != sql.ErrNoRows {
			return err
		}
		if _, err := tx.ExecContext(ctx, string(body)); err != nil {
			return fmt.Errorf("apply %s: %w", entry.Name(), err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(name, checksum) VALUES ($1,$2)`, entry.Name(), checksum); err != nil {
			return err
		}
	}
	return tx.Commit()
}
