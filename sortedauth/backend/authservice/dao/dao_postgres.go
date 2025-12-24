package dao

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// PostgresDAO implements the DAO interface using PostgreSQL and sqlx
type UserPostgresDAO struct {
	db *sqlx.DB
}

// NewPostgresDAO creates a new PostgreSQL DAO instance
func NewUserPostgresDAO(config *PostgresConfig) (*UserPostgresDAO, error) {
	dsn := config.GetPostgresDSN()

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.Pool.MaxOpenConnections)
	db.SetMaxIdleConns(config.Pool.MaxIdleConnections)
	db.SetConnMaxLifetime(config.Pool.ConnectionMaxLifetime)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	slog.Info("PostgreSQL DAO created successfully",
		"host", config.Host,
		"port", config.Port,
		"database", config.Database,
		"max_open_conns", config.Pool.MaxOpenConnections)

	return &UserPostgresDAO{db: db}, nil
}

func (u *UserPostgresDAO) DoesUserExist(userID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM sortedauth_users WHERE user_id = $1`
	err := u.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

func (u *UserPostgresDAO) CreateUserIfNotExists(userID, email, name string) (string, error) {
	// First check if user exists by email
	existingUserID, err := u.GetUserIDByEmail(email)
	if err == nil && existingUserID != "" {
		// User already exists, return existing user ID
		return existingUserID, nil
	}

	// Check if user already exists by userID
	exists, err := u.DoesUserExist(userID)
	if err != nil {
		slog.Error("Error checking if user already exists by userID", "error", err)
		return "", err
	}
	if exists {
		slog.Debug("User already exists by userID", "userID", userID)
		return userID, nil
	}

	// User doesn't exist, insert new user
	insertQuery := `
		INSERT INTO sortedauth_users 
		(user_id, email, name)
		VALUES ($1, $2, $3)
	`

	slog.Debug("Inserting new user", "userID", userID, "email", email, "name", name)
	_, err = u.db.Exec(insertQuery, userID, email, name)
	if err != nil {
		slog.Error("Error inserting new user", "error", err)
		return "", err
	}

	return userID, nil
}

func (u *UserPostgresDAO) GetUserIDByEmail(email string) (string, error) {
	var userID string
	query := `SELECT user_id FROM sortedauth_users WHERE email = $1`
	err := u.db.QueryRow(query, email).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil // User not found, but not an error
		}
		slog.Error("Error getting user ID by email", "error", err)
		return "", err
	}
	return userID, nil
}

func (u *UserPostgresDAO) GetUsersList(page, pageSize int64) ([]*User, error) {
	var users []*User
	query := `SELECT user_id, email, name, '' as roles FROM sortedauth_users LIMIT $1 OFFSET $2`
	err := u.db.Select(&users, query, pageSize, (page-1)*pageSize)
	if err != nil {
		slog.Error("authservice:dao:GetUsersList", "error", err)
		return nil, err
	}
	return users, nil
}

func (u *UserPostgresDAO) CreateAccountIfNotExists(userID string, provider string, providerAccountID string) (string, error) {
	// Check if account already exists
	var existingID string
	checkQuery := `SELECT id FROM sortedauth_accounts WHERE provider = $1 AND provider_account_id = $2`
	err := u.db.QueryRow(checkQuery, provider, providerAccountID).Scan(&existingID)
	if err == nil && existingID != "" {
		return existingID, nil
	}

	// Generate new account ID
	accountID := uuid.New().String()

	insertQuery := `
		INSERT INTO sortedauth_accounts (id, user_id, provider, provider_account_id)
		VALUES ($1, $2, $3, $4)
	`
	_, err = u.db.Exec(insertQuery, accountID, userID, provider, providerAccountID)
	if err != nil {
		slog.Error("Error creating account", "error", err)
		return "", err
	}
	return accountID, nil
}

// TenantPostgresDAO implements the TenantDAO interface using PostgreSQL
type TenantPostgresDAO struct {
	db *sqlx.DB
}

// NewTenantPostgresDAO creates a new PostgreSQL Tenant DAO instance
func NewTenantPostgresDAO(config *PostgresConfig) (*TenantPostgresDAO, error) {
	dsn := config.GetPostgresDSN()

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.Pool.MaxOpenConnections)
	db.SetMaxIdleConns(config.Pool.MaxIdleConnections)
	db.SetConnMaxLifetime(config.Pool.ConnectionMaxLifetime)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	return &TenantPostgresDAO{db: db}, nil
}

func (t *TenantPostgresDAO) CreateTenant(name string, description string, tenantType int64, createdBy string) (string, error) {
	tenantID := uuid.New().String()

	insertQuery := `
		INSERT INTO sortedauth_tenants (id, name, description, type, created_by)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := t.db.Exec(insertQuery, tenantID, name, description, tenantType, createdBy)
	if err != nil {
		slog.Error("Error creating tenant", "error", err)
		return "", err
	}
	return tenantID, nil
}

func (t *TenantPostgresDAO) GetTenantUsers(tenantID string, page int64, pageSize int64) ([]*TenantUser, error) {
	var tenantUsers []*TenantUser
	query := `
		SELECT u.user_id, u.name, u.email, tu.tenant_id, tu.role
		FROM sortedauth_tenant_users tu
		JOIN sortedauth_users u ON tu.user_id = u.user_id
		WHERE tu.tenant_id = $1
		LIMIT $2 OFFSET $3
	`
	err := t.db.Select(&tenantUsers, query, tenantID, pageSize, (page-1)*pageSize)
	if err != nil {
		slog.Error("Error getting tenant users", "error", err)
		return nil, err
	}
	return tenantUsers, nil
}

func (t *TenantPostgresDAO) GetTenantsList(page, pageSize int64) ([]*Tenant, error) {
	var tenants []*Tenant
	query := `
		SELECT id, name, description, type, created_by
		FROM sortedauth_tenants
		LIMIT $1 OFFSET $2
	`
	err := t.db.Select(&tenants, query, pageSize, (page-1)*pageSize)
	if err != nil {
		slog.Error("Error getting tenants list", "error", err)
		return nil, err
	}
	return tenants, nil
}

func (t *TenantPostgresDAO) AddUserToTenant(tenantID string, userID string, role string) (string, error) {
	// Check if user is already in tenant
	var existingID string
	checkQuery := `SELECT id FROM sortedauth_tenant_users WHERE tenant_id = $1 AND user_id = $2`
	err := t.db.QueryRow(checkQuery, tenantID, userID).Scan(&existingID)
	if err == nil && existingID != "" {
		return existingID, nil
	}

	relationID := uuid.New().String()
	insertQuery := `
		INSERT INTO sortedauth_tenant_users (id, tenant_id, user_id, role)
		VALUES ($1, $2, $3, $4)
	`
	_, err = t.db.Exec(insertQuery, relationID, tenantID, userID, role)
	if err != nil {
		slog.Error("Error adding user to tenant", "error", err)
		return "", err
	}
	return relationID, nil
}

func (t *TenantPostgresDAO) RemoveUserFromTenant(tenantID string, userID string) (string, error) {
	deleteQuery := `DELETE FROM sortedauth_tenant_users WHERE tenant_id = $1 AND user_id = $2`
	result, err := t.db.Exec(deleteQuery, tenantID, userID)
	if err != nil {
		slog.Error("Error removing user from tenant", "error", err)
		return "", err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return "", err
	}

	if rowsAffected == 0 {
		return "", fmt.Errorf("user %s not found in tenant %s", userID, tenantID)
	}

	return fmt.Sprintf("Removed user %s from tenant %s", userID, tenantID), nil
}

// ProviderPostgresDAO implements the ProviderDAO interface using PostgreSQL
type ProviderPostgresDAO struct {
	db *sqlx.DB
}

// NewProviderPostgresDAO creates a new PostgreSQL Provider DAO instance
func NewProviderPostgresDAO(config *PostgresConfig) (*ProviderPostgresDAO, error) {
	dsn := config.GetPostgresDSN()

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.Pool.MaxOpenConnections)
	db.SetMaxIdleConns(config.Pool.MaxIdleConnections)
	db.SetConnMaxLifetime(config.Pool.ConnectionMaxLifetime)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	return &ProviderPostgresDAO{db: db}, nil
}

func (p *ProviderPostgresDAO) CreateProvider(name string, enabled bool, oauthURL string, tokenURL string, redirectURL string, scope string) (string, error) {
	providerID := uuid.New().String()

	insertQuery := `
		INSERT INTO sortedauth_providers (id, name, enabled, oauth_url, token_url, redirecturl, scope)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := p.db.Exec(insertQuery, providerID, name, enabled, oauthURL, tokenURL, redirectURL, scope)
	if err != nil {
		slog.Error("Error creating provider", "error", err)
		return "", err
	}
	return providerID, nil
}

func (p *ProviderPostgresDAO) GetProvider(providerID string) (*Provider, error) {
	var provider Provider
	query := `
		SELECT id, name, enabled, oauth_url, token_url, redirecturl as redirect_url, scope
		FROM sortedauth_providers
		WHERE id = $1
	`
	err := p.db.Get(&provider, query, providerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("provider not found: %s", providerID)
		}
		slog.Error("Error getting provider", "error", err)
		return nil, err
	}
	return &provider, nil
}

func (p *ProviderPostgresDAO) GetProvidersList(page, pageSize int64) ([]*Provider, error) {
	var providers []*Provider
	query := `
		SELECT id, name, enabled, oauth_url, token_url, redirecturl as redirect_url, scope
		FROM sortedauth_providers
		LIMIT $1 OFFSET $2
	`
	err := p.db.Select(&providers, query, pageSize, (page-1)*pageSize)
	if err != nil {
		slog.Error("Error getting providers list", "error", err)
		return nil, err
	}
	return providers, nil
}
