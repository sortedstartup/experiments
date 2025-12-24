package dao

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserDaoFactory struct {
}

// SQLiteDAO implements the DAO interface using SQLite and sqlx
type UserSqliteDAO struct {
	db *sqlx.DB
}

// NewSQLiteDAO creates a new SQLite DAO instance
func NewUserSqliteDAO(sqliteUrl string) (*UserSqliteDAO, error) {
	// sqlite_vec.Auto()

	db, err := sqlx.Open("sqlite3", sqliteUrl)
	if err != nil {
		return nil, err
	}

	return &UserSqliteDAO{db: db}, nil
}

func (u *UserSqliteDAO) DoesUserExist(userID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM sortedauth_users WHERE user_id = ?`
	err := u.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

func (u *UserSqliteDAO) CreateUserIfNotExists(userID, email, name string) (string, error) {
	// First check if user exists by email
	existingUserID, err := u.GetUserIDByEmail(email)
	if err == nil && existingUserID != "" {
		// User already exists, return existing user ID
		slog.Debug("User already exists by email", "email", email)
		return existingUserID, nil
	}

	// Check if user already exists by userID
	slog.Debug("Checking if user already exists by userID", "userID", userID)
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
		VALUES (?, ?, ?)
	`

	slog.Debug("Inserting new user", "userID", userID, "email", email, "name", name)
	_, err = u.db.Exec(insertQuery, userID, email, name)
	if err != nil {
		slog.Error("Error inserting new user", "error", err)
		return "", err
	}

	return userID, nil
}

func (u *UserSqliteDAO) GetUserIDByEmail(email string) (string, error) {
	var userID string
	query := `SELECT user_id FROM sortedauth_users WHERE email = ?`
	fmt.Println("SQLITE db variable in GetUserIDByEmail", u.db)
	err := u.db.QueryRow(query, email).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil // User not found, but not an error
		}
		return "", err
	}
	return userID, nil
}

func (u *UserSqliteDAO) GetUsersList(page, pageSize int64) ([]*User, error) {
	var users []*User
	query := `SELECT user_id, email, name, '' as roles FROM sortedauth_users LIMIT ? OFFSET ?`
	err := u.db.Select(&users, query, pageSize, (page-1)*pageSize)
	if err != nil {
		slog.Error("authservice:dao:GetUsersList", "error", err)
		return nil, err
	}
	return users, nil
}

func (u *UserSqliteDAO) CreateAccountIfNotExists(userID string, provider string, providerAccountID string) (string, error) {
	// Check if account already exists
	var existingID string
	checkQuery := `SELECT id FROM sortedauth_accounts WHERE provider = ? AND provider_account_id = ?`
	err := u.db.QueryRow(checkQuery, provider, providerAccountID).Scan(&existingID)
	if err == nil && existingID != "" {
		return existingID, nil
	}

	// Generate new account ID
	accountID := uuid.New().String()

	insertQuery := `
		INSERT INTO sortedauth_accounts (id, user_id, provider, provider_account_id)
		VALUES (?, ?, ?, ?)
	`
	_, err = u.db.Exec(insertQuery, accountID, userID, provider, providerAccountID)
	if err != nil {
		slog.Error("Error creating account", "error", err)
		return "", err
	}
	return accountID, nil
}

// TenantSqliteDAO implements the TenantDAO interface using SQLite
type TenantSqliteDAO struct {
	db *sqlx.DB
}

// NewTenantSqliteDAO creates a new SQLite Tenant DAO instance
func NewTenantSqliteDAO(sqliteUrl string) (*TenantSqliteDAO, error) {
	db, err := sqlx.Open("sqlite3", sqliteUrl)
	if err != nil {
		return nil, err
	}
	return &TenantSqliteDAO{db: db}, nil
}

func (t *TenantSqliteDAO) CreateTenant(name string, description string, tenantType int64, createdBy string) (string, error) {
	tenantID := uuid.New().String()

	insertQuery := `
		INSERT INTO sortedauth_tenants (id, name, description, type, created_by)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := t.db.Exec(insertQuery, tenantID, name, description, tenantType, createdBy)
	if err != nil {
		slog.Error("Error creating tenant", "error", err)
		return "", err
	}
	return tenantID, nil
}

func (t *TenantSqliteDAO) GetTenantUsers(tenantID string, page int64, pageSize int64) ([]*TenantUser, error) {
	var tenantUsers []*TenantUser
	query := `
		SELECT u.user_id, u.name, u.email, tu.tenant_id, tu.role
		FROM sortedauth_tenant_users tu
		JOIN sortedauth_users u ON tu.user_id = u.user_id
		WHERE tu.tenant_id = ?
		LIMIT ? OFFSET ?
	`
	err := t.db.Select(&tenantUsers, query, tenantID, pageSize, (page-1)*pageSize)
	if err != nil {
		slog.Error("Error getting tenant users", "error", err)
		return nil, err
	}
	return tenantUsers, nil
}

func (t *TenantSqliteDAO) GetTenantsList(page, pageSize int64) ([]*Tenant, error) {
	var tenants []*Tenant
	query := `
		SELECT id, name, description, type, created_by
		FROM sortedauth_tenants
		LIMIT ? OFFSET ?
	`
	err := t.db.Select(&tenants, query, pageSize, (page-1)*pageSize)
	if err != nil {
		slog.Error("Error getting tenants list", "error", err)
		return nil, err
	}
	return tenants, nil
}

func (t *TenantSqliteDAO) AddUserToTenant(tenantID string, userID string, role string) (string, error) {
	// Check if user is already in tenant
	var existingID string
	checkQuery := `SELECT id FROM sortedauth_tenant_users WHERE tenant_id = ? AND user_id = ?`
	err := t.db.QueryRow(checkQuery, tenantID, userID).Scan(&existingID)
	if err == nil && existingID != "" {
		return existingID, nil
	}

	relationID := uuid.New().String()
	insertQuery := `
		INSERT INTO sortedauth_tenant_users (id, tenant_id, user_id, role)
		VALUES (?, ?, ?, ?)
	`
	_, err = t.db.Exec(insertQuery, relationID, tenantID, userID, role)
	if err != nil {
		slog.Error("Error adding user to tenant", "error", err)
		return "", err
	}
	return relationID, nil
}

func (t *TenantSqliteDAO) RemoveUserFromTenant(tenantID string, userID string) (string, error) {
	deleteQuery := `DELETE FROM sortedauth_tenant_users WHERE tenant_id = ? AND user_id = ?`
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

// ProviderSqliteDAO implements the ProviderDAO interface using SQLite
type ProviderSqliteDAO struct {
	db *sqlx.DB
}

// NewProviderSqliteDAO creates a new SQLite Provider DAO instance
func NewProviderSqliteDAO(sqliteUrl string) (*ProviderSqliteDAO, error) {
	db, err := sqlx.Open("sqlite3", sqliteUrl)
	if err != nil {
		return nil, err
	}
	return &ProviderSqliteDAO{db: db}, nil
}

func (p *ProviderSqliteDAO) CreateProvider(name string, enabled bool, oauthURL string, tokenURL string, redirectURL string, scope string) (string, error) {
	providerID := uuid.New().String()

	enabledInt := 0
	if enabled {
		enabledInt = 1
	}

	insertQuery := `
		INSERT INTO sortedauth_providers (id, name, enabled, oauth_url, token_url, redirect_url, scope)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := p.db.Exec(insertQuery, providerID, name, enabledInt, oauthURL, tokenURL, redirectURL, scope)
	if err != nil {
		slog.Error("Error creating provider", "error", err)
		return "", err
	}
	return providerID, nil
}

func (p *ProviderSqliteDAO) GetProvider(providerID string) (*Provider, error) {
	var provider Provider
	query := `
		SELECT id, name, enabled, oauth_url, token_url, redirect_url, scope
		FROM sortedauth_providers
		WHERE id = ?
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

func (p *ProviderSqliteDAO) GetProvidersList(page, pageSize int64) ([]*Provider, error) {
	var providers []*Provider
	query := `
		SELECT id, name, enabled, oauth_url, token_url, redirect_url, scope
		FROM sortedauth_providers
		LIMIT ? OFFSET ?
	`
	err := p.db.Select(&providers, query, pageSize, (page-1)*pageSize)
	if err != nil {
		slog.Error("Error getting providers list", "error", err)
		return nil, err
	}
	return providers, nil
}
