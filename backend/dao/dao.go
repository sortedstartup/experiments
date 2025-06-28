package dao

import (
	"database/sql"
)

// This package will contain database access logic.
// Implement your DB models and methods here.

// DAO is the interface for database operations.
type DAO interface {
	SaveMessage(chatID, message string) error
	GetMessages(chatID string) ([]string, error)
}

// NewDAO returns a new DAO implementation.
func NewDAO(db *sql.DB) DAO {
	return &sqliteDAO{db: db}
}

type sqliteDAO struct {
	db *sql.DB
}

// Implement the DAO interface methods for sqliteDAO here.
func (d *sqliteDAO) SaveMessage(chatID, message string) error {
	_, err := d.db.Exec("INSERT INTO messages (chat_id, message) VALUES (?, ?)", chatID, message)
	return err
}

func (d *sqliteDAO) GetMessages(chatID string) ([]string, error) {
	rows, err := d.db.Query("SELECT message FROM messages WHERE chat_id = ?", chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []string
	for rows.Next() {
		var msg string
		if err := rows.Scan(&msg); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
