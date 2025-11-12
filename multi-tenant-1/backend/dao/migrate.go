package dao

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed db/scripts/migrations/*
var sqliteMigrationFiles embed.FS

const migrationTable = "app_migrations"

// MigrateSQLite runs migrations using embedded migration files.
func MigrateSQLite(dbURL string) error {
	entries, err := fs.ReadDir(sqliteMigrationFiles, "db/scripts/migrations")
	if err != nil {
		log.Printf("DEBUG: failed to list embedded migration files: %v", err)
	} else {
		for _, entry := range entries {
			log.Printf("DEBUG: embedded migration file: %s", entry.Name())
		}
	}

	sqlDB, err := sql.Open("sqlite3", dbURL)
	if err != nil {
		return fmt.Errorf("failed to open sqlite db: %w", err)
	}

	files, err := iofs.New(sqliteMigrationFiles, "db/scripts/migrations")
	if err != nil {
		return fmt.Errorf("failed to load migration files: %w", err)
	}

	dbInstance, err := sqlite.WithInstance(sqlDB, &sqlite.Config{MigrationsTable: migrationTable})
	if err != nil {
		return fmt.Errorf("failed to create sqlite instance: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", files, "sqlite3", dbInstance)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}
	log.Println("Migrations applied successfully")
	return nil
}
