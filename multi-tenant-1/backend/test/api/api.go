package api

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc/metadata"

	"sortedstartup/multi-tenant/dao"
	"sortedstartup/multi-tenant/test/proto"
)

type Server struct {
	proto.UnimplementedSortedtestServer
	DAO dao.DAO
}

func NewServer(db *sql.DB) *Server {
	myDao := dao.NewDAO(db)
	return &Server{DAO: myDao}
}

func (s *Server) CreateTenant(ctx context.Context, req *proto.CreateTenantRequest) (*proto.CreateTenantResponse, error) {
	id := uuid.New().String()
	err := s.DAO.CreateTenant(ctx, id, req.Name)
	if err != nil {
		return &proto.CreateTenantResponse{Message: "Failed to create tenant: " + err.Error()}, err
	}

	tenantDBPath := id + ".db"
	db, err := sql.Open("sqlite3", tenantDBPath)
	if err != nil {
		return &proto.CreateTenantResponse{Message: "Tenant created, but failed to create DB: " + err.Error()}, err
	}
	defer db.Close()

	// yaha kisi trah se sari tables like (project,task....) tables initialize krni pdegi
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
	db, err := sql.Open("sqlite3", tenantID+".db")
	if err != nil {
		return &proto.CreateProjectResponse{Message: "db open fail " + err.Error()}, err
	}
	defer db.Close()

	projectID := uuid.New().String()
	fmt.Println(projectID)
	fmt.Println(req.Name)
	daoInstance := dao.NewDAO(db)
	err = daoInstance.CreateProject(ctx, projectID, req.Name)
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
	db, err := sql.Open("sqlite3", tenantID+".db")
	if err != nil {
		return &proto.CreateTaskResponse{Message: "db open fail " + err.Error()}, err
	}
	defer db.Close()

	taskID := uuid.New().String()
	fmt.Println(taskID)
	fmt.Println(req.Name)
	fmt.Println(req.ProjectId)
	daoInstance := dao.NewDAO(db)
	err = daoInstance.CreateTask(ctx, taskID, req.ProjectId, req.Name)
	if err != nil {
		return &proto.CreateTaskResponse{Message: "task fail " + err.Error()}, err
	}

	return &proto.CreateTaskResponse{Message: taskID}, nil
}
