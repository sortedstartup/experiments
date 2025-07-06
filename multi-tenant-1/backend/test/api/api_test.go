package api

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"sortedstartup/multi-tenant/test/proto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

// Mock structs for testing
type MockSuperDAO struct {
	mock.Mock
}

func (m *MockSuperDAO) CreateTenant(ctx context.Context, id, name string) error {
	args := m.Called(ctx, id, name)
	return args.Error(0)
}

type MockTenantDAO struct {
	mock.Mock
}

func (m *MockTenantDAO) CreateProject(ctx context.Context, tenantID, projectID, name string) error {
	args := m.Called(ctx, tenantID, projectID, name)
	return args.Error(0)
}

func (m *MockTenantDAO) CreateTask(ctx context.Context, tenantID, taskID, projectID, name string) error {
	args := m.Called(ctx, tenantID, taskID, projectID, name)
	return args.Error(0)
}

// Test helper function to create a test server
func createTestServer() *Server {
	// Ensure the mono directory exists for DB creation
	monoDir := filepath.Join("..", "mono")
	_ = os.MkdirAll(monoDir, 0755)

	// Create a simple logger for testing
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	return &Server{
		SuperDAO:  &MockSuperDAO{},
		TenantDAO: &MockTenantDAO{},
		Log:       logger,
	}
}

// Test helper to create context with tenant ID
func createContextWithTenantID(tenantID string) context.Context {
	md := metadata.Pairs("tenant-id", tenantID)
	return metadata.NewIncomingContext(context.Background(), md)
}

// Tests for CreateTenant
func TestServer_CreateTenant_Success(t *testing.T) {
	server := createTestServer()
	mockSuperDAO := server.SuperDAO.(*MockSuperDAO)

	// Setup mock expectations
	mockSuperDAO.On("CreateTenant", mock.Anything, mock.AnythingOfType("string"), "test-tenant").Return(nil)

	// Create request
	req := &proto.CreateTenantRequest{
		Name: "test-tenant",
	}

	// Call the method
	resp, err := server.CreateTenant(context.Background(), req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Message) // Message should contain the tenant ID

	// Verify that the tenant ID is a valid UUID
	_, err = uuid.Parse(resp.Message)
	assert.NoError(t, err, "Response message should be a valid UUID")

	// Verify mock was called
	mockSuperDAO.AssertExpectations(t)
}

func TestServer_CreateTenant_DAOError(t *testing.T) {
	server := createTestServer()
	mockSuperDAO := server.SuperDAO.(*MockSuperDAO)

	// Setup mock to return error
	expectedError := fmt.Errorf("database connection failed")
	mockSuperDAO.On("CreateTenant", mock.Anything, mock.AnythingOfType("string"), "test-tenant").Return(expectedError)

	req := &proto.CreateTenantRequest{
		Name: "test-tenant",
	}

	resp, err := server.CreateTenant(context.Background(), req)

	// Assertions
	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.Message, "Failed to create tenant:")
	assert.Contains(t, resp.Message, "database connection failed")

	mockSuperDAO.AssertExpectations(t)
}

// Tests for CreateProject
func TestServer_CreateProject_Success(t *testing.T) {
	server := createTestServer()
	mockTenantDAO := server.TenantDAO.(*MockTenantDAO)

	tenantID := "test-tenant-123"
	projectName := "test-project"

	// Setup mock expectations
	mockTenantDAO.On("CreateProject", mock.Anything, tenantID, mock.AnythingOfType("string"), projectName).Return(nil)

	// Create context with tenant ID
	ctx := createContextWithTenantID(tenantID)

	req := &proto.CreateProjectRequest{
		Name: projectName,
	}

	resp, err := server.CreateProject(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Message)

	// Verify that the project ID is a valid UUID
	_, err = uuid.Parse(resp.Message)
	assert.NoError(t, err, "Response message should be a valid UUID")

	mockTenantDAO.AssertExpectations(t)
}

func TestServer_CreateProject_MissingTenantID(t *testing.T) {
	server := createTestServer()

	req := &proto.CreateProjectRequest{
		Name: "test-project",
	}

	// Call without tenant ID in context
	resp, err := server.CreateProject(context.Background(), req)

	// Assertions
	assert.NoError(t, err) // The method returns nil error but error message in response
	assert.NotNil(t, resp)
	assert.Contains(t, resp.Message, "Missing")
}

func TestServer_CreateProject_InvalidMetadata(t *testing.T) {
	server := createTestServer()

	// Create context with empty metadata
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs())

	req := &proto.CreateProjectRequest{
		Name: "test-project",
	}

	resp, err := server.CreateProject(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.Message, "Missing tenant_id in header")
}

func TestServer_CreateProject_DAOError(t *testing.T) {
	server := createTestServer()
	mockTenantDAO := server.TenantDAO.(*MockTenantDAO)

	tenantID := "test-tenant-123"
	projectName := "test-project"

	// Setup mock to return error
	expectedError := fmt.Errorf("project creation failed")
	mockTenantDAO.On("CreateProject", mock.Anything, tenantID, mock.AnythingOfType("string"), projectName).Return(expectedError)

	ctx := createContextWithTenantID(tenantID)
	req := &proto.CreateProjectRequest{
		Name: projectName,
	}

	resp, err := server.CreateProject(ctx, req)

	// Assertions
	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.Message, "Failed to create project:")
	assert.Contains(t, resp.Message, "project creation failed")

	mockTenantDAO.AssertExpectations(t)
}

// Tests for CreateTask
func TestServer_CreateTask_Success(t *testing.T) {
	server := createTestServer()
	mockTenantDAO := server.TenantDAO.(*MockTenantDAO)

	tenantID := "test-tenant-123"
	taskName := "test-task"
	projectID := "project-456"

	// Setup mock expectations
	mockTenantDAO.On("CreateTask", mock.Anything, tenantID, mock.AnythingOfType("string"), projectID, taskName).Return(nil)

	ctx := createContextWithTenantID(tenantID)
	req := &proto.CreateTaskRequest{
		Name:      taskName,
		ProjectId: projectID,
	}

	resp, err := server.CreateTask(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Message)

	// Verify that the task ID is a valid UUID
	_, err = uuid.Parse(resp.Message)
	assert.NoError(t, err, "Response message should be a valid UUID")

	mockTenantDAO.AssertExpectations(t)
}

func TestServer_CreateTask_MissingTenantID(t *testing.T) {
	server := createTestServer()

	req := &proto.CreateTaskRequest{
		Name:      "test-task",
		ProjectId: "project-456",
	}

	resp, err := server.CreateTask(context.Background(), req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.Message, "Missing")
}

func TestServer_CreateTask_DAOError(t *testing.T) {
	server := createTestServer()
	mockTenantDAO := server.TenantDAO.(*MockTenantDAO)

	tenantID := "test-tenant-123"
	taskName := "test-task"
	projectID := "project-456"

	// Setup mock to return error
	expectedError := fmt.Errorf("task creation failed")
	mockTenantDAO.On("CreateTask", mock.Anything, tenantID, mock.AnythingOfType("string"), projectID, taskName).Return(expectedError)

	ctx := createContextWithTenantID(tenantID)
	req := &proto.CreateTaskRequest{
		Name:      taskName,
		ProjectId: projectID,
	}

	resp, err := server.CreateTask(ctx, req)

	// Assertions
	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.Message, "Failed to create task:")
	assert.Contains(t, resp.Message, "task creation failed")

	mockTenantDAO.AssertExpectations(t)
}

// Tests for ExtractTenantID helper function
func TestExtractTenantID_Success(t *testing.T) {
	expectedTenantID := "tenant-123"
	ctx := createContextWithTenantID(expectedTenantID)

	tenantID, err := ExtractTenantID(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedTenantID, tenantID)
}

func TestExtractTenantID_NoMetadata(t *testing.T) {
	ctx := context.Background()

	tenantID, err := ExtractTenantID(ctx)

	assert.Error(t, err)
	assert.Empty(t, tenantID)
	assert.Contains(t, err.Error(), "Missing metadata")
}

func TestExtractTenantID_NoTenantHeader(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("other-header", "value"))

	tenantID, err := ExtractTenantID(ctx)

	assert.Error(t, err)
	assert.Empty(t, tenantID)
	assert.Contains(t, err.Error(), "Missing tenant_id in header")
}
