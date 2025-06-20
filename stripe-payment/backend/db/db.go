package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite" // SQLite driver

)

var DB *sql.DB

// InitDB sets up the connection and creates tables
func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open db: %w", err)
	}

	// test connection
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("db ping error: %w", err)
	}

	// schema
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		is_sub BOOLEAN NOT NULL,
		session_id TEXT
	);
	`

	_, err = DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

func SaveUserAfterPayment(email string, isSub bool, sessionID string) error {
	// Insert or update user
	query := `
		INSERT INTO users (email, is_sub, session_id)
		VALUES (?, ?, ?)
		ON CONFLICT(email) DO UPDATE SET
			is_sub = excluded.is_sub,
			session_id = excluded.session_id;
	`

	_, err := DB.Exec(query, email, isSub, sessionID)
	if err != nil {
		return fmt.Errorf("db insert error: %w", err)
	}
	return nil
}
