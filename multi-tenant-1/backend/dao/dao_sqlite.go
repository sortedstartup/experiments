package dao

import (
	"context"
	"database/sql"
	"sync"
)

// This package will contain database access logic.
// Implement your DB models and methods here.

// SuperDAO is for operations on the super DB (app.db)
type SuperDAO interface {
	CreateTenant(ctx context.Context, id, name string) error
}

type superDAO struct {
	db *sql.DB
}

func NewSuperDAO(db *sql.DB) SuperDAO {
	return &superDAO{db: db}
}

func (d *superDAO) CreateTenant(ctx context.Context, id, name string) error {
	_, err := d.db.Exec("INSERT INTO tenants (id, name) VALUES (?, ?)", id, name)
	return err
}

type TenantDAO interface {
	CreateProject(ctx context.Context, tenantID, id, name string) error
	CreateTask(ctx context.Context, tenantID, id, projectId, name string) error
}

type tenantDAO struct{}

func NewTenantDAO() TenantDAO {
	return &tenantDAO{}
}

func (d *tenantDAO) CreateProject(ctx context.Context, tenantID, id, name string) error {
	db, err := GetOrCreateTenantDB(tenantID)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "INSERT INTO project (id, name) VALUES (?, ?)", id, name)
	return err
}

func (d *tenantDAO) CreateTask(ctx context.Context, tenantID, id, projectId, name string) error {
	db, err := GetOrCreateTenantDB(tenantID)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "INSERT INTO task (id, project_id, name) VALUES (?, ?, ?)", id, projectId, name)
	return err
}

var (
	tenantDBs   = make(map[string]*sql.DB)
	tenantDBsMu sync.Mutex
)

func GetOrCreateTenantDB(tenantID string) (*sql.DB, error) {
	tenantDBsMu.Lock()
	defer tenantDBsMu.Unlock()
	if db, ok := tenantDBs[tenantID]; ok {
		return db, nil
	}
	db, err := sql.Open("sqlite3", tenantID+".db")
	if err != nil {
		return nil, err
	}
	tenantDBs[tenantID] = db
	return db, nil
}
