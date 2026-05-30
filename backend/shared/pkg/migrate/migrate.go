// Package migrate — общая логика применения SQL-миграций.
//
// Используется:
//   - cmd/server/main.go (migrate subcommand)
//   - internal/testutil/testcontainers.go (интеграционные тесты)
//
// Поддерживает два режима:
//  1. Tracked — с таблицей schema_migrations (production)
//  2. Blind — без отслеживания (test containers)
package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	schemaName    = "lkfl_platform"
	trackingTable = schemaName + ".schema_migrations"
)

// execFunc — функция для выполнения SQL.
type execFunc func(string, ...interface{}) error

// queryRowFunc — функция для QueryRow.
type queryRowFunc func(string, ...interface{}) pgx.Row

// Apply применяет все SQL-миграции из директории.
//
// conn — *pgx.Conn для production (tracked режим).
// Если conn = nil, используется pool режим (test).
//
// tracked — если true, используется таблица schema_migrations
// для отслеживания применённых миграций.
func Apply(ctx context.Context, conn *pgx.Conn, pool *pgxpool.Pool, tracked bool) error {
	var exec execFunc
	var qrow queryRowFunc

	if conn != nil {
		exec = func(sql string, args ...interface{}) error {
			_, err := conn.Exec(ctx, sql, args...)
			return err
		}
		qrow = func(sql string, args ...interface{}) pgx.Row {
			return conn.QueryRow(ctx, sql, args...)
		}
	} else if pool != nil {
		exec = func(sql string, args ...interface{}) error {
			_, err := pool.Exec(ctx, sql, args...)
			return err
		}
		qrow = func(sql string, args ...interface{}) pgx.Row {
			return pool.QueryRow(ctx, sql, args...)
		}
	} else {
		return fmt.Errorf("migrate: no connection provided")
	}

	if err := createSchema(exec); err != nil {
		return err
	}

	if tracked {
		if err := createTrackingTable(exec); err != nil {
			return err
		}
	}

	migrationDir := FindMigrationsDir()
	if migrationDir == "" {
		return nil // миграции не найдены — OK (пустой проект)
	}

	files, err := readMigrationFiles(migrationDir)
	if err != nil {
		return err
	}

	for _, name := range files {
		if tracked {
			exists, err := isMigrationApplied(qrow, name)
			if err != nil {
				return err
			}
			if exists {
				continue
			}
		}

		data, err := os.ReadFile(filepath.Join(migrationDir, name))
		if err != nil {
			return wrapError("read migration %s", name, err)
		}

		if err := exec(string(data)); err != nil {
			return wrapError("apply migration %s", name, err)
		}

		if tracked {
			if err := recordMigration(exec, name); err != nil {
				return err
			}
		}
	}

	return nil
}

// createSchema создаёт схему lkfl_platform если она не существует.
func createSchema(exec execFunc) error {
	return exec("CREATE SCHEMA IF NOT EXISTS " + schemaName)
}

// createTrackingTable создаёт таблицу отслеживания миграций.
func createTrackingTable(exec execFunc) error {
	return exec(`
		CREATE TABLE IF NOT EXISTS ` + trackingTable + ` (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
}

// isMigrationApplied проверяет, применена ли миграция.
func isMigrationApplied(qrow queryRowFunc, name string) (bool, error) {
	var exists bool
	err := qrow(`
		SELECT EXISTS(SELECT 1 FROM `+trackingTable+` WHERE filename = $1)
	`, name).Scan(&exists)
	return exists, err
}

// recordMigration записывает факт применения миграции.
func recordMigration(exec execFunc, name string) error {
	return exec(`
		INSERT INTO `+trackingTable+` (filename) VALUES ($1)
	`, name)
}

// readMigrationFiles читает список .sql файлов (без .sql.down).
func readMigrationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, wrapError("read migrations dir %s", dir, err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") && !strings.HasSuffix(e.Name(), ".sql.down") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

// FindMigrationsDir ищет директорию миграций.
func FindMigrationsDir() string {
	candidates := []string{
		"migrations",
		"../migrations",
		"../../migrations",
		"../../../migrations",
		"../../../../migrations",
		"/app/migrations",
	}
	for _, c := range candidates {
		info, err := os.Stat(c)
		if err == nil && info.IsDir() {
			return c
		}
	}
	return ""
}

func wrapError(format string, name string, err error) error {
	return &MigrationError{Action: format, Name: name, Err: err}
}

// MigrationError — ошибка применения миграции.
type MigrationError struct {
	Action string
	Name   string
	Err    error
}

func (e *MigrationError) Error() string {
	return e.Err.Error()
}

func (e *MigrationError) Unwrap() error {
	return e.Err
}
