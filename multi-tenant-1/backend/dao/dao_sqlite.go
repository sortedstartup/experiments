package dao

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel"
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
	_, span := otel.Tracer("go_manual").Start(ctx, "db")
	defer span.End()
	_, err := d.db.Exec("INSERT INTO tenants (id, name) VALUES (?, ?)", id, name)
	return err
}

type TenantDAO interface {
	CreateProject(ctx context.Context, tenantID, id, name string) error
	CreateTask(ctx context.Context, tenantID, id, projectId, name string) error
	GetProjects(ctx context.Context, tenantID string) ([]Project, error)
	GetTasks(ctx context.Context, tenantID, projectId string) ([]Task, error)
}

type tenantDAO struct{}

func NewTenantDAO() TenantDAO {
	return &tenantDAO{}
}

func InitTenantDBs(superDB *sql.DB, _ string) error {
	rows, err := superDB.Query("SELECT id FROM tenants")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var tid string
		if err := rows.Scan(&tid); err != nil {
			return err
		}
		db, err := sql.Open("sqlite3", tid+".db")
		if err != nil {
			return err
		}
		tenantDBsMu.Lock()
		tenantDBs[tid] = db
		tenantDBsMu.Unlock()
	}
	return nil
}

func RegisterTenantDB(ctx context.Context, tenantID, _ string) error {
	_, span := otel.Tracer("go_manual").Start(ctx, "register db")
	defer span.End()
	db, err := sql.Open("sqlite3", tenantID+".db")
	if err != nil {
		return err
	}
	// tenantDBsMu.Lock()
	// defer tenantDBsMu.Unlock()
	tenantDBs[tenantID] = db
	return nil
}

var (
	tenantDBs   = make(map[string]*sql.DB)
	tenantDBsMu sync.Mutex

	// Export tenantDBs and tenantDBsMu for direct access from api package
	TenantDBs   = tenantDBs
	TenantDBsMu = &tenantDBsMu
)

func (d *tenantDAO) CreateProject(ctx context.Context, tenantID, id, name string) error {
	_, span := otel.Tracer("go_manual").Start(ctx, "create project dao")
	defer span.End()
	db, ok := tenantDBs[tenantID]
	if !ok {
		return fmt.Errorf("tenant DB not found  %s", tenantID)
	}
	_, err := db.ExecContext(ctx, "INSERT INTO project (id, name) VALUES (?, ?)", id, name)
	return err
}

func (d *tenantDAO) CreateTask(ctx context.Context, tenantID, id, projectId, name string) error {
	_, span := otel.Tracer("go_manual").Start(ctx, "create task dao")
	defer span.End()
	db, ok := tenantDBs[tenantID]
	if !ok {
		return fmt.Errorf("tenant DB not found  %s", tenantID)
	}
	_, err := db.ExecContext(ctx, "INSERT INTO task (id, project_id, name) VALUES (?, ?, ?)", id, projectId, name)
	return err
}

type Project struct {
	ID   string
	Name string
}

type Task struct {
	ID        string
	Name      string
	ProjectID string
}

func (d *tenantDAO) GetProjects(ctx context.Context, tenantID string) ([]Project, error) {
	db, ok := tenantDBs[tenantID]
	if !ok {
		return nil, fmt.Errorf("tenant DB not found  %s", tenantID)
	}
	rows, err := db.QueryContext(ctx, "SELECT id, name FROM project")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (d *tenantDAO) GetTasks(ctx context.Context, tenantID, projectId string) ([]Task, error) {
	db, ok := tenantDBs[tenantID]
	if !ok {
		return nil, fmt.Errorf("tenant DB not found  %s", tenantID)
	}
	rows, err := db.QueryContext(ctx, "SELECT id, name, project_id FROM task WHERE project_id = ?", projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Name, &t.ProjectID); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// Add this helper to access tenantDBs safely from outside the package
func TenantDBExists(tenantID string) (*sql.DB, bool) {
	tenantDBsMu.Lock()
	defer tenantDBsMu.Unlock()
	db, ok := tenantDBs[tenantID]
	return db, ok
}
