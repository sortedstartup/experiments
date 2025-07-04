package api

import (
	"context"
	"database/sql"
	"fmt"

	"sortedstartup/multi-tenant/dao"
	"sortedstartup/multi-tenant/test/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

type Server struct {
	proto.UnimplementedSortedtestServer
	SuperDAO  dao.SuperDAO
	TenantDAO dao.TenantDAO
}

func NewServer(superDB *sql.DB) *Server {
	superDao := dao.NewSuperDAO(superDB)
	tenantDao := dao.NewTenantDAO()
	return &Server{SuperDAO: superDao, TenantDAO: tenantDao}
}

func (s *Server) CreateTenant(ctx context.Context, req *proto.CreateTenantRequest) (*proto.CreateTenantResponse, error) {
	id := uuid.New().String()
	err := s.SuperDAO.CreateTenant(ctx, id, req.Name)
	if err != nil {
		return &proto.CreateTenantResponse{Message: "Failed to create tenant: " + err.Error()}, err
	}

	// Create and initialize the tenant DB tables
	db, err := dao.GetOrCreateTenantDB(id)
	if err != nil {
		return &proto.CreateTenantResponse{Message: "Tenant created, but failed to create DB: " + err.Error()}, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS project (id TEXT PRIMARY KEY, name TEXT)`)
	if err != nil {
		return &proto.CreateTenantResponse{Message: "db created, failed project table " + err.Error()}, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS task (id TEXT PRIMARY KEY,project_id TEXT, name TEXT)`)
	if err != nil {
		return &proto.CreateTenantResponse{Message: "db created, failed task table " + err.Error()}, err
	}

	return &proto.CreateTenantResponse{Message: id}, nil
}

func ExtractTenantID(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("Missing metadata")
	}
	if vals := md.Get("tenant-id"); len(vals) > 0 {
		return vals[0], nil
	} else if vals := md.Get("tenant_id"); len(vals) > 0 {
		return vals[0], nil
	} else if vals := md.Get("x-tenant-id"); len(vals) > 0 {
		return vals[0], nil
	}
	return "", fmt.Errorf("Missing tenant_id in header")
}

func (s *Server) CreateProject(ctx context.Context, req *proto.CreateProjectRequest) (*proto.CreateProjectResponse, error) {
	tenantID, err := ExtractTenantID(ctx)
	if err != nil {
		return &proto.CreateProjectResponse{Message: err.Error()}, nil
	}
	projectID := uuid.New().String()
	fmt.Println(projectID)
	fmt.Println(req.Name)
	err = s.TenantDAO.CreateProject(ctx, tenantID, projectID, req.Name)
	if err != nil {
		return &proto.CreateProjectResponse{Message: "Failed project create : " + err.Error()}, err
	}

	return &proto.CreateProjectResponse{Message: projectID}, nil
}

func (s *Server) CreateTask(ctx context.Context, req *proto.CreateTaskRequest) (*proto.CreateTaskResponse, error) {
	tenantID, err := ExtractTenantID(ctx)
	if err != nil {
		return &proto.CreateTaskResponse{Message: err.Error()}, nil
	}
	taskID := uuid.New().String()
	fmt.Println(taskID)
	fmt.Println(req.Name)
	fmt.Println(req.ProjectId)
	err = s.TenantDAO.CreateTask(ctx, tenantID, taskID, req.ProjectId, req.Name)
	if err != nil {
		return &proto.CreateTaskResponse{Message: "task fail " + err.Error()}, err
	}

	return &proto.CreateTaskResponse{Message: taskID}, nil
}
