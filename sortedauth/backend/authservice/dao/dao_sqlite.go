package dao

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

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
	query := `SELECT COUNT(*) FROM userservice_users WHERE user_id = ?`
	err := u.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

func (u *UserSqliteDAO) CreateUserIfNotExists(userID, email, name, roles, oAuthProvider, oAuthUserID string, isFederated bool) (string, error) {
	// First check if user exists by email
	existingUserID, err := u.GetUserIDByEmail(email)
	if err == nil && existingUserID != "" {
		// User already exists, return existing user ID
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
		INSERT INTO userservice_users 
		(user_id, email, name, roles, oauth_provider, oauth_user_id, is_federated)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	slog.Debug("Inserting new user", "userID", userID, "email", email, "name", name, "roles", roles, "oAuthProvider", oAuthProvider, "oAuthUserID", oAuthUserID, "isFederated", isFederated)
	_, err = u.db.Exec(insertQuery, userID, email, name, roles, oAuthProvider, oAuthUserID, isFederated)
	if err != nil {
		slog.Error("Error inserting new user", "error", err)
		return "", err
	}
	return userID, nil
}

func (u *UserSqliteDAO) GetUserIDByEmail(email string) (string, error) {
	var userID string
	query := `SELECT user_id FROM userservice_users WHERE email = ?`
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
	query := `SELECT user_id, email, name, roles FROM userservice_users LIMIT ? OFFSET ?`
	err := u.db.Select(&users, query, pageSize, (page-1)*pageSize)
	if err != nil {
		slog.Error("authservice:dao:GetUsersList", "error", err)
		return nil, err
	}
	return users, nil
}
