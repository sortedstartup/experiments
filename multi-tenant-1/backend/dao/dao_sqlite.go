package dao

import (
	"context"
	"database/sql"
)

// This package will contain database access logic.
// Implement your DB models and methods here.

// DAO is the interface for database operations.
type DAO interface {
	CreateTenant(ctx context.Context, id string, name string) error
	CreateProject(ctx context.Context, id string, name string) error
	CreateTask(ctx context.Context, id string, projectId string, name string) error
}

// NewDAO returns a new DAO implementation.
func NewDAO(db *sql.DB) DAO {
	return &sqliteDAO{db: db}
}

type sqliteDAO struct {
	db *sql.DB
	// ctx context.Context
}

func (d *sqliteDAO) CreateTenant(ctx context.Context, id string, name string) error {
	_, err := d.db.Exec("INSERT INTO tenants (id, name) VALUES (?, ?)", id, name)
	return err
}

func (d *sqliteDAO) CreateProject(ctx context.Context, id string, name string) error {
	_, err := d.db.ExecContext(ctx, "INSERT INTO project (id, name) VALUES (?, ?)", id, name)
	return err
}

func (d *sqliteDAO) CreateTask(ctx context.Context, id string, projectId string, name string) error {
	_, err := d.db.ExecContext(ctx, "INSERT INTO task (id, project_id, name) VALUES (?, ?, ?)", id, projectId, name)
	return err
}
