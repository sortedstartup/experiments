package api

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"sortedstartup/multi-tenant/dao"
	"sortedstartup/multi-tenant/test/proto"

	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	otellog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc/metadata"
)

type Server struct {
	proto.UnimplementedSortedtestServer
	SuperDAO  dao.SuperDAO
	Log       *slog.Logger
	TenantDAO dao.TenantDAO
}

func NewServer(superDB *sql.DB, loggerProvider *otellog.LoggerProvider) *Server {
	log := otelslog.NewLogger("my/pkg/name", otelslog.WithLoggerProvider(loggerProvider))
	superDao := dao.NewSuperDAO(superDB)
	tenantDao := dao.NewTenantDAO()
	return &Server{SuperDAO: superDao, TenantDAO: tenantDao, Log: log}
}

func (s *Server) CreateTenant(ctx context.Context, req *proto.CreateTenantRequest) (*proto.CreateTenantResponse, error) {
	s.Log.Info("In create tenant api")

	ctx, span := otel.Tracer("go_manual").Start(ctx, "create tenant api layer")
	defer span.End()

	id := uuid.New().String()
	err := s.SuperDAO.CreateTenant(ctx, id, req.Name)
	if err != nil {
		return &proto.CreateTenantResponse{Message: "Failed to create tenant: " + err.Error()}, err
	}

	// Create the tenant DB file and register it
	dbPath := fmt.Sprintf("../mono/%s.db", id)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return &proto.CreateTenantResponse{Message: "Failed to create tenant DB: " + err.Error()}, err
	}
	defer db.Close()
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS project (id TEXT PRIMARY KEY, name TEXT)`)
	if err != nil {
		return &proto.CreateTenantResponse{Message: "db created, failed project table " + err.Error()}, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS task (id TEXT PRIMARY KEY,project_id TEXT, name TEXT)`)
	if err != nil {
		return &proto.CreateTenantResponse{Message: "db created, failed task table " + err.Error()}, err
	}
	s.Log.Info("about to register in create tenant api")
	if err := dao.RegisterTenantDB(ctx, id, dbPath); err != nil {
		return &proto.CreateTenantResponse{Message: "Failed to register tenant DB: " + err.Error()}, err
	}

	return &proto.CreateTenantResponse{Message: id}, nil
}

func ExtractTenantID(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("missing metadata")
	}
	if vals := md.Get("tenant-id"); len(vals) > 0 {
		return vals[0], nil
	}
	return "", fmt.Errorf("missing tenant_id in header")
}

func (s *Server) CreateProject(ctx context.Context, req *proto.CreateProjectRequest) (*proto.CreateProjectResponse, error) {
	s.Log.Info("In create project api")

	ctx, span := otel.Tracer("go_manual").Start(ctx, "create project api layer")
	defer span.End()

	tenantID, err := ExtractTenantID(ctx)
	if err != nil {
		return &proto.CreateProjectResponse{Message: err.Error()}, nil
	}
	if tenantID == "" {
		return &proto.CreateProjectResponse{Message: "Missing tenant ID in header"}, nil
	}
	_, ok := dao.TenantDBs[tenantID]
	if !ok {
		return &proto.CreateProjectResponse{Message: fmt.Sprintf("tenant not found: %s", tenantID)}, nil
	}
	projectID := uuid.New().String()

	err = s.TenantDAO.CreateProject(ctx, tenantID, projectID, req.Name)
	if err != nil {
		return &proto.CreateProjectResponse{Message: "Failed to create project: " + err.Error()}, err
	}

	return &proto.CreateProjectResponse{Message: projectID}, nil
}

func (s *Server) CreateTask(ctx context.Context, req *proto.CreateTaskRequest) (*proto.CreateTaskResponse, error) {
	s.Log.Info("In create task api")

	ctx, span := otel.Tracer("go_manual").Start(ctx, "create task api layer")
	defer span.End()

	tenantID, err := ExtractTenantID(ctx)
	if err != nil {
		return &proto.CreateTaskResponse{Message: err.Error()}, nil
	}
	if tenantID == "" {
		return &proto.CreateTaskResponse{Message: "Missing tenant ID in header"}, nil
	}
	_, ok := dao.TenantDBs[tenantID]
	if !ok {
		return &proto.CreateTaskResponse{Message: fmt.Sprintf("tenant not found: %s", tenantID)}, nil
	}
	taskID := uuid.New().String()

	err = s.TenantDAO.CreateTask(ctx, tenantID, taskID, req.ProjectId, req.Name)
	if err != nil {
		return &proto.CreateTaskResponse{Message: "Failed to create task: " + err.Error()}, err
	}

	return &proto.CreateTaskResponse{Message: taskID}, nil
}

func (s *Server) GetProjects(ctx context.Context, req *proto.GetProjectsRequest) (*proto.GetProjectsResponse, error) {
	tenantID, err := ExtractTenantID(ctx)
	if err != nil {
		return nil, fmt.Errorf("missing or invalid tenant ID: %w", err)
	}
	if tenantID == "" {
		return nil, fmt.Errorf("missing tenant ID in header")
	}

	_, ok := dao.TenantDBs[tenantID]
	if !ok {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}
	projects, err := s.TenantDAO.GetProjects(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	var protoProjects []*proto.Project
	for _, p := range projects {
		protoProjects = append(protoProjects, &proto.Project{Id: p.ID, Name: p.Name})
	}
	return &proto.GetProjectsResponse{Projects: protoProjects}, nil
}

func (s *Server) GetTasks(ctx context.Context, req *proto.GetTasksRequest) (*proto.GetTasksResponse, error) {
	tenantID, err := ExtractTenantID(ctx)
	if err != nil {
		return nil, fmt.Errorf("missing or invalid tenant ID: %w", err)
	}
	if tenantID == "" {
		return nil, fmt.Errorf("missing tenant ID in header")
	}
	_, ok := dao.TenantDBs[tenantID]
	if !ok {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}
	tasks, err := s.TenantDAO.GetTasks(ctx, tenantID, req.ProjectId)
	if err != nil {
		return nil, err
	}
	var protoTasks []*proto.Task
	for _, t := range tasks {
		protoTasks = append(protoTasks, &proto.Task{Id: t.ID, Name: t.Name, ProjectId: t.ProjectID})
	}
	return &proto.GetTasksResponse{Tasks: protoTasks}, nil
}
