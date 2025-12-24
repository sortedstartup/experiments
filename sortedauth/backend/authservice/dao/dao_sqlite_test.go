package dao

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates an in-memory SQLite database and runs migrations
// Returns the sqlx.DB connection that can be used to create any DAO type
func setupTestDB(t *testing.T) *sqlx.DB {
	// Create in-memory database
	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}

	// Run migrations using the embedded migration system
	err = MigrateDB_UsingConnection_SQLite(
		db.DB,
		sqliteMigrationFiles,
		"db/sqlite/scripts/migrations",
		MIGRATION_TABLE,
	)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

func TestCreateUserIfNotExists_NewUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	dao := &UserSqliteDAO{db: db}

	userID := "user-123"
	email := "test@example.com"
	name := "Test User"

	// Create a new user
	returnedUserID, err := dao.CreateUserIfNotExists(userID, email, name)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if returnedUserID != userID {
		t.Errorf("Expected userID %s, got %s", userID, returnedUserID)
	}

	// Verify user was created
	exists, err := dao.DoesUserExist(userID)
	if err != nil {
		t.Fatalf("Error checking if user exists: %v", err)
	}
	if !exists {
		t.Error("Expected user to exist, but it doesn't")
	}

	// Verify user details
	retrievedUserID, err := dao.GetUserIDByEmail(email)
	if err != nil {
		t.Fatalf("Error getting user by email: %v", err)
	}
	if retrievedUserID != userID {
		t.Errorf("Expected userID %s, got %s", userID, retrievedUserID)
	}
}

func TestCreateUserIfNotExists_ExistingUserByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	dao := &UserSqliteDAO{db: db}

	userID1 := "user-123"
	email := "test@example.com"
	name1 := "Test User"

	// Create initial user
	returnedUserID1, err := dao.CreateUserIfNotExists(userID1, email, name1)
	if err != nil {
		t.Fatalf("Expected no error on first create, got: %v", err)
	}
	if returnedUserID1 != userID1 {
		t.Errorf("Expected userID %s, got %s", userID1, returnedUserID1)
	}

	// Try to create user with same email but different userID
	userID2 := "user-456"
	name2 := "Different User"
	returnedUserID2, err := dao.CreateUserIfNotExists(userID2, email, name2)
	if err != nil {
		t.Fatalf("Expected no error on duplicate email, got: %v", err)
	}

	// Should return the original user ID
	if returnedUserID2 != userID1 {
		t.Errorf("Expected original userID %s, got %s", userID1, returnedUserID2)
	}

	// Verify that the second userID was NOT created
	exists, err := dao.DoesUserExist(userID2)
	if err != nil {
		t.Fatalf("Error checking if user exists: %v", err)
	}
	if exists {
		t.Error("Expected second userID to not be created")
	}
}

func TestCreateUserIfNotExists_ExistingUserByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	dao := &UserSqliteDAO{db: db}

	userID := "user-123"
	email1 := "test1@example.com"
	name := "Test User"

	// Create initial user
	returnedUserID1, err := dao.CreateUserIfNotExists(userID, email1, name)
	if err != nil {
		t.Fatalf("Expected no error on first create, got: %v", err)
	}
	if returnedUserID1 != userID {
		t.Errorf("Expected userID %s, got %s", userID, returnedUserID1)
	}

	// Try to create user with same userID but different email
	email2 := "test2@example.com"
	returnedUserID2, err := dao.CreateUserIfNotExists(userID, email2, name)
	if err != nil {
		t.Fatalf("Expected no error on duplicate userID, got: %v", err)
	}

	// Should return the existing user ID
	if returnedUserID2 != userID {
		t.Errorf("Expected existing userID %s, got %s", userID, returnedUserID2)
	}

	// Verify that the original email is still associated
	retrievedUserID, err := dao.GetUserIDByEmail(email1)
	if err != nil {
		t.Fatalf("Error getting user by original email: %v", err)
	}
	if retrievedUserID != userID {
		t.Errorf("Expected original email to still be associated with userID %s, got %s", userID, retrievedUserID)
	}

	// Verify that the new email is NOT associated
	retrievedUserID2, err := dao.GetUserIDByEmail(email2)
	if err != nil {
		t.Fatalf("Error getting user by new email: %v", err)
	}
	if retrievedUserID2 != "" {
		t.Errorf("Expected new email to not be associated, but got userID %s", retrievedUserID2)
	}
}

func TestCreateUserIfNotExists_MultipleUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	dao := &UserSqliteDAO{db: db}

	// Create multiple different users
	users := []struct {
		userID string
		email  string
		name   string
	}{
		{"user-1", "user1@example.com", "User One"},
		{"user-2", "user2@example.com", "User Two"},
		{"user-3", "user3@example.com", "User Three"},
	}

	for _, user := range users {
		returnedUserID, err := dao.CreateUserIfNotExists(user.userID, user.email, user.name)
		if err != nil {
			t.Fatalf("Error creating user %s: %v", user.userID, err)
		}
		if returnedUserID != user.userID {
			t.Errorf("Expected userID %s, got %s", user.userID, returnedUserID)
		}
	}

	// Verify all users exist
	for _, user := range users {
		exists, err := dao.DoesUserExist(user.userID)
		if err != nil {
			t.Fatalf("Error checking if user %s exists: %v", user.userID, err)
		}
		if !exists {
			t.Errorf("Expected user %s to exist", user.userID)
		}

		retrievedUserID, err := dao.GetUserIDByEmail(user.email)
		if err != nil {
			t.Fatalf("Error getting user by email %s: %v", user.email, err)
		}
		if retrievedUserID != user.userID {
			t.Errorf("Expected userID %s for email %s, got %s", user.userID, user.email, retrievedUserID)
		}
	}
}

func TestCreateUserIfNotExists_EmptyFields(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	dao := &UserSqliteDAO{db: db}

	// Test with empty name (should still work)
	userID := "user-empty-name"
	email := "empty@example.com"
	returnedUserID, err := dao.CreateUserIfNotExists(userID, email, "")
	if err != nil {
		t.Fatalf("Expected no error with empty name, got: %v", err)
	}
	if returnedUserID != userID {
		t.Errorf("Expected userID %s, got %s", userID, returnedUserID)
	}
}

// ========== AddUserToTenant Tests ==========

func TestAddUserToTenant_BasicCase(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	dao := &TenantSqliteDAO{db: db}
	userDAO := &UserSqliteDAO{db: db}

	// Create test user and tenant
	userID := "user-1"
	tenantID := "tenant-1"
	_, err := userDAO.CreateUserIfNotExists(userID, "user1@example.com", "User One")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create tenant with specific ID for test
	_, err = db.Exec(`INSERT INTO sortedauth_tenants (id, name, description, type, created_by) VALUES (?, ?, ?, ?, ?)`,
		tenantID, "Tenant One", "Test tenant", 1, userID)
	if err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}

	// Add user to tenant
	relationID, err := dao.AddUserToTenant(tenantID, userID, "admin")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if relationID == "" {
		t.Error("Expected non-empty relation ID")
	}

	// Verify the relationship was created
	users, err := dao.GetTenantUsers(tenantID, 1, 100)
	if err != nil {
		t.Fatalf("Error getting tenant users: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("Expected 1 user in tenant, got %d", len(users))
	}
	if users[0].UserId != userID {
		t.Errorf("Expected userID %s, got %s", userID, users[0].UserId)
	}
	if users[0].Role != "admin" {
		t.Errorf("Expected role 'admin', got %s", users[0].Role)
	}
}

func TestAddUserToTenant_DuplicateRelationship(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	dao := &TenantSqliteDAO{db: db}
	userDAO := &UserSqliteDAO{db: db}

	// Create test user and tenant
	userID := "user-1"
	tenantID := "tenant-1"
	_, err := userDAO.CreateUserIfNotExists(userID, "user1@example.com", "User One")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err = db.Exec(`INSERT INTO sortedauth_tenants (id, name, description, type, created_by) VALUES (?, ?, ?, ?, ?)`,
		tenantID, "Tenant One", "Test tenant", 1, userID)
	if err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}

	// Add user to tenant first time
	relationID1, err := dao.AddUserToTenant(tenantID, userID, "admin")
	if err != nil {
		t.Fatalf("Expected no error on first add, got: %v", err)
	}

	// Try to add the same user to the same tenant again
	relationID2, err := dao.AddUserToTenant(tenantID, userID, "user")
	if err != nil {
		t.Fatalf("Expected no error on duplicate add, got: %v", err)
	}

	// Should return the existing relation ID
	if relationID2 != relationID1 {
		t.Errorf("Expected existing relationID %s, got %s", relationID1, relationID2)
	}

	// Verify only one relationship exists
	users, err := dao.GetTenantUsers(tenantID, 1, 100)
	if err != nil {
		t.Fatalf("Error getting tenant users: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Expected 1 user in tenant, got %d", len(users))
	}
}

func TestAddUserToTenant_MultipleUsersOneTenant(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	dao := &TenantSqliteDAO{db: db}
	userDAO := &UserSqliteDAO{db: db}

	// Create test users
	tenantID := "tenant-1"
	users := []struct {
		userID string
		email  string
		name   string
		role   string
	}{
		{"user-1", "user1@example.com", "User One", "admin"},
		{"user-2", "user2@example.com", "User Two", "user"},
		{"user-3", "user3@example.com", "User Three", "viewer"},
	}

	// Create all users
	for _, user := range users {
		_, err := userDAO.CreateUserIfNotExists(user.userID, user.email, user.name)
		if err != nil {
			t.Fatalf("Failed to create user %s: %v", user.userID, err)
		}
	}

	// Create tenant (created by first user)
	_, err := db.Exec(`INSERT INTO sortedauth_tenants (id, name, description, type, created_by) VALUES (?, ?, ?, ?, ?)`,
		tenantID, "Tenant One", "Test tenant", 1, users[0].userID)
	if err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}

	// Add all users to the tenant
	for _, user := range users {
		relationID, err := dao.AddUserToTenant(tenantID, user.userID, user.role)
		if err != nil {
			t.Fatalf("Error adding user %s to tenant: %v", user.userID, err)
		}
		if relationID == "" {
			t.Errorf("Expected non-empty relation ID for user %s", user.userID)
		}
	}

	// Verify all users are in the tenant
	tenantUsers, err := dao.GetTenantUsers(tenantID, 1, 100)
	if err != nil {
		t.Fatalf("Error getting tenant users: %v", err)
	}
	if len(tenantUsers) != len(users) {
		t.Fatalf("Expected %d users in tenant, got %d", len(users), len(tenantUsers))
	}

	// Verify each user has correct role
	userRoleMap := make(map[string]string)
	for _, tu := range tenantUsers {
		userRoleMap[tu.UserId] = tu.Role
	}

	for _, user := range users {
		role, exists := userRoleMap[user.userID]
		if !exists {
			t.Errorf("User %s not found in tenant", user.userID)
		}
		if role != user.role {
			t.Errorf("Expected role %s for user %s, got %s", user.role, user.userID, role)
		}
	}
}

func TestAddUserToTenant_OneUserMultipleTenants(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	dao := &TenantSqliteDAO{db: db}
	userDAO := &UserSqliteDAO{db: db}

	// Create one user
	userID := "user-1"
	_, err := userDAO.CreateUserIfNotExists(userID, "user1@example.com", "User One")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create multiple tenants
	tenants := []struct {
		tenantID    string
		name        string
		description string
		role        string
	}{
		{"tenant-1", "Tenant One", "First tenant", "admin"},
		{"tenant-2", "Tenant Two", "Second tenant", "user"},
		{"tenant-3", "Tenant Three", "Third tenant", "viewer"},
	}

	for _, tenant := range tenants {
		_, err := db.Exec(`INSERT INTO sortedauth_tenants (id, name, description, type, created_by) VALUES (?, ?, ?, ?, ?)`,
			tenant.tenantID, tenant.name, tenant.description, 1, userID)
		if err != nil {
			t.Fatalf("Failed to create tenant %s: %v", tenant.tenantID, err)
		}
	}

	// Add the user to all tenants with different roles
	for _, tenant := range tenants {
		relationID, err := dao.AddUserToTenant(tenant.tenantID, userID, tenant.role)
		if err != nil {
			t.Fatalf("Error adding user to tenant %s: %v", tenant.tenantID, err)
		}
		if relationID == "" {
			t.Errorf("Expected non-empty relation ID for tenant %s", tenant.tenantID)
		}
	}

	// Verify user is in all tenants with correct roles
	for _, tenant := range tenants {
		tenantUsers, err := dao.GetTenantUsers(tenant.tenantID, 1, 100)
		if err != nil {
			t.Fatalf("Error getting users for tenant %s: %v", tenant.tenantID, err)
		}
		if len(tenantUsers) != 1 {
			t.Fatalf("Expected 1 user in tenant %s, got %d", tenant.tenantID, len(tenantUsers))
		}
		if tenantUsers[0].UserId != userID {
			t.Errorf("Expected userID %s in tenant %s, got %s", userID, tenant.tenantID, tenantUsers[0].UserId)
		}
		if tenantUsers[0].Role != tenant.role {
			t.Errorf("Expected role %s in tenant %s, got %s", tenant.role, tenant.tenantID, tenantUsers[0].Role)
		}
	}
}

func TestAddUserToTenant_ManyToMany(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	dao := &TenantSqliteDAO{db: db}
	userDAO := &UserSqliteDAO{db: db}

	// Create multiple users
	users := []struct {
		userID string
		email  string
		name   string
	}{
		{"user-1", "user1@example.com", "User One"},
		{"user-2", "user2@example.com", "User Two"},
		{"user-3", "user3@example.com", "User Three"},
	}

	for _, user := range users {
		_, err := userDAO.CreateUserIfNotExists(user.userID, user.email, user.name)
		if err != nil {
			t.Fatalf("Failed to create user %s: %v", user.userID, err)
		}
	}

	// Create multiple tenants
	tenants := []struct {
		tenantID    string
		name        string
		description string
	}{
		{"tenant-1", "Tenant One", "First tenant"},
		{"tenant-2", "Tenant Two", "Second tenant"},
	}

	for _, tenant := range tenants {
		_, err := db.Exec(`INSERT INTO sortedauth_tenants (id, name, description, type, created_by) VALUES (?, ?, ?, ?, ?)`,
			tenant.tenantID, tenant.name, tenant.description, 1, users[0].userID)
		if err != nil {
			t.Fatalf("Failed to create tenant %s: %v", tenant.tenantID, err)
		}
	}

	// Create many-to-many relationships
	// User 1 -> Tenant 1 (admin), Tenant 2 (admin)
	// User 2 -> Tenant 1 (user)
	// User 3 -> Tenant 2 (user)
	relationships := []struct {
		tenantID string
		userID   string
		role     string
	}{
		{"tenant-1", "user-1", "admin"},
		{"tenant-1", "user-2", "user"},
		{"tenant-2", "user-1", "admin"},
		{"tenant-2", "user-3", "user"},
	}

	for _, rel := range relationships {
		relationID, err := dao.AddUserToTenant(rel.tenantID, rel.userID, rel.role)
		if err != nil {
			t.Fatalf("Error creating relationship (%s, %s): %v", rel.tenantID, rel.userID, err)
		}
		if relationID == "" {
			t.Errorf("Expected non-empty relation ID for (%s, %s)", rel.tenantID, rel.userID)
		}
	}

	// Verify Tenant 1 has 2 users
	tenant1Users, err := dao.GetTenantUsers("tenant-1", 1, 100)
	if err != nil {
		t.Fatalf("Error getting users for tenant-1: %v", err)
	}
	if len(tenant1Users) != 2 {
		t.Errorf("Expected 2 users in tenant-1, got %d", len(tenant1Users))
	}

	// Verify Tenant 2 has 2 users
	tenant2Users, err := dao.GetTenantUsers("tenant-2", 1, 100)
	if err != nil {
		t.Fatalf("Error getting users for tenant-2: %v", err)
	}
	if len(tenant2Users) != 2 {
		t.Errorf("Expected 2 users in tenant-2, got %d", len(tenant2Users))
	}

	// Verify specific relationships
	// Check if user-1 is admin in both tenants
	tenant1Roles := make(map[string]string)
	for _, u := range tenant1Users {
		tenant1Roles[u.UserId] = u.Role
	}
	if tenant1Roles["user-1"] != "admin" {
		t.Errorf("Expected user-1 to be admin in tenant-1, got %s", tenant1Roles["user-1"])
	}
	if tenant1Roles["user-2"] != "user" {
		t.Errorf("Expected user-2 to be user in tenant-1, got %s", tenant1Roles["user-2"])
	}

	tenant2Roles := make(map[string]string)
	for _, u := range tenant2Users {
		tenant2Roles[u.UserId] = u.Role
	}
	if tenant2Roles["user-1"] != "admin" {
		t.Errorf("Expected user-1 to be admin in tenant-2, got %s", tenant2Roles["user-1"])
	}
	if tenant2Roles["user-3"] != "user" {
		t.Errorf("Expected user-3 to be user in tenant-2, got %s", tenant2Roles["user-3"])
	}
}

func TestAddUserToTenant_RemoveUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	dao := &TenantSqliteDAO{db: db}
	userDAO := &UserSqliteDAO{db: db}

	// Create test user and tenant
	userID := "user-1"
	tenantID := "tenant-1"
	_, err := userDAO.CreateUserIfNotExists(userID, "user1@example.com", "User One")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err = db.Exec(`INSERT INTO sortedauth_tenants (id, name, description, type, created_by) VALUES (?, ?, ?, ?, ?)`,
		tenantID, "Tenant One", "Test tenant", 1, userID)
	if err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}

	// Add user to tenant
	_, err = dao.AddUserToTenant(tenantID, userID, "admin")
	if err != nil {
		t.Fatalf("Error adding user to tenant: %v", err)
	}

	// Verify user is in tenant
	users, err := dao.GetTenantUsers(tenantID, 1, 100)
	if err != nil {
		t.Fatalf("Error getting tenant users: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("Expected 1 user in tenant, got %d", len(users))
	}

	// Remove user from tenant
	result, err := dao.RemoveUserFromTenant(tenantID, userID)
	if err != nil {
		t.Fatalf("Error removing user from tenant: %v", err)
	}
	if result == "" {
		t.Error("Expected non-empty result message")
	}

	// Verify user is no longer in tenant
	users, err = dao.GetTenantUsers(tenantID, 1, 100)
	if err != nil {
		t.Fatalf("Error getting tenant users after removal: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("Expected 0 users in tenant after removal, got %d", len(users))
	}
}

func TestAddUserToTenant_RemoveNonExistentUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	dao := &TenantSqliteDAO{db: db}
	userDAO := &UserSqliteDAO{db: db}

	// Create test user and tenant
	userID := "user-1"
	tenantID := "tenant-1"
	_, err := userDAO.CreateUserIfNotExists(userID, "user1@example.com", "User One")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err = db.Exec(`INSERT INTO sortedauth_tenants (id, name, description, type, created_by) VALUES (?, ?, ?, ?, ?)`,
		tenantID, "Tenant One", "Test tenant", 1, userID)
	if err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}

	// Try to remove a user that's not in the tenant
	_, err = dao.RemoveUserFromTenant(tenantID, "non-existent-user")
	if err == nil {
		t.Error("Expected error when removing non-existent user, got nil")
	}
}
