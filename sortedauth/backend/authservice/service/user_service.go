package service

import (
	"log"
	"log/slog"
	"os"
	"sortedstartup/authservice/dao"
	"strings"

	"github.com/google/uuid"
)

type UserService struct {
	dao dao.UserDAO
}

func NewUserService(dao dao.UserDAO) *UserService {
	slog.Debug("authservice:service:NewUserService")
	return &UserService{dao: dao}
}

func (u *UserService) Init(config *dao.Config) {
	slog.Debug("authservice:service:Init")
	switch config.Database.Type {
	case dao.DatabaseTypeSQLite:
		slog.Info("UserService: Running SQLite migrations")
		if err := dao.MigrateSQLite(config.Database.SQLite.URL); err != nil {
			log.Fatalf("UserService: Failed to migrate SQLite database: %v", err)
		}
		if err := dao.SeedSqlite(config.Database.SQLite.URL); err != nil {
			log.Fatalf("UserService: Failed to seed SQLite database: %v", err)
		}
	case dao.DatabaseTypePostgres:
		slog.Info("UserService: Running PostgreSQL migrations")
		dsn := config.Database.Postgres.GetPostgresDSN()
		if err := dao.MigratePostgres(dsn); err != nil {
			log.Fatalf("UserService: Failed to migrate PostgreSQL database: %v", err)
		}
		if err := dao.SeedPostgres(dsn); err != nil {
			log.Fatalf("UserService: Failed to seed PostgreSQL database: %v", err)
		}
	default:
		log.Fatalf("UserService: Unsupported database type: %s", config.Database.Type)
	}
}

// GenerateNewUserID generates a new unique user ID using UUID
func GenerateNewUserID() string {
	slog.Info("authservice:service:GenerateNewUserID")
	return uuid.New().String()
}

func (u *UserService) DoesUserExist(userID string) (bool, error) {
	slog.Info("authservice:service:DoesUserExist", "userID", userID)
	return u.dao.DoesUserExist(userID)
}

func (u *UserService) CreateUserIfNotExists(email, name, roles, oAuthProvider, oAuthUserID string, isFederated bool) (string, error) {
	slog.Info("authservice:service:CreateUserIfNotExists", "email", email, "roles", roles, "oAuthProvider", oAuthProvider, "oAuthUserID", oAuthUserID, "isFederated", isFederated)
	// Generate a new user ID
	userID := GenerateNewUserID()

	return u.dao.CreateUserIfNotExists(userID, email, name)
}

// getEnvOrDefault returns the environment variable value or the default value
func getEnvOrDefault(key, defaultValue string) string {
	slog.Info("authservice:service:getEnvOrDefault", "key", key, "defaultValue", defaultValue)
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isEmailAllowed checks if the given email is in the allowlist from ALLOWED_LOGIN_EMAILS env variable
// The environment variable should contain comma-separated email addresses
// If ALLOWED_LOGIN_EMAILS is not set or empty, all emails are allowed
func isEmailAllowed(email string) bool {
	slog.Info("authservice:service:isEmailAllowed", "email", email)
	allowedEmails := os.Getenv("ALLOWED_LOGIN_EMAILS")

	// If no allowlist is configured, allow all emails
	if allowedEmails == "" {
		return true
	}

	// Split the comma-separated list and check each email
	emails := strings.Split(allowedEmails, ",")
	for _, allowedEmail := range emails {
		// Trim whitespace and compare (case-insensitive)
		if strings.TrimSpace(strings.ToLower(allowedEmail)) == strings.ToLower(email) {
			return true
		}
	}

	return false
}

func (u *UserService) GetUsersList(page int64, pageSize int64) ([]*dao.User, error) {
	slog.Info("authservice:service:GetUsersList", "page", page, "pageSize", pageSize)
	users, err := u.dao.GetUsersList(page, pageSize)
	if err != nil {
		slog.Error("authservice:service:GetUsersList", "error", err)
		return nil, err
	}
	return users, nil
}
