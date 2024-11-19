package database

// Use an sqlite3 database to store the data from the LLM output
// The database will have the following tables:
// 1. conversations: store the prompt, response, and model name
// 2. FUTURE: models: store the model name and the model's parameters

// Use this module like this:
// db := NewDB("path/to/database.db")
// InsertConversation(db, "prompt", "response", "model_name")

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type SQLite3DB struct {
	db *sql.DB
}

// Retun errors to the caller in case we want to ignore them. That is, just
// because we can't store the conversations in the database doesn't mean we
// should stop the program.
func NewDB(dbPath string) (*SQLite3DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	err = createConversationsTable(db)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	sqlDB := SQLite3DB{}
	sqlDB.db = db
	return &sqlDB, nil
}

func createConversationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS conversations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			prompt TEXT NOT NULL,
			response TEXT NOT NULL,
			model_name TEXT
		);
	`)
	if err != nil {
		return fmt.Errorf("error creating conversations table: %v", err)
	}

	return nil
}

func (sqlDB *SQLite3DB) InsertConversation(prompt, response, modelName string) error {
	_, err := sqlDB.db.Exec(`
		INSERT INTO conversations (prompt, response, model_name)
		VALUES (?, ?, ?);
	`, prompt, response, modelName)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}

// TODO: some serious refactoring can be done with these query functions
func QueryAllConversations(db *sql.DB) error {
	rows, err := db.Query(`
		SELECT * FROM conversations;
	`)
	if err != nil {
		return fmt.Errorf("error querying database for conversations: %v", err)
	}

	for rows.Next() {
		var id int
		var timestamp string
		var prompt string
		var response string
		var modelName string
		err := rows.Scan(&id, &timestamp, &prompt, &response, &modelName)
		if err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}

		fmt.Printf("id: %d, timestamp: %s, prompt: %s, response: %s, model_name: %s\n", id, timestamp, prompt, response, modelName)
	}

	return nil
}

func QueryConversationsByModel(db *sql.DB, modelName string) error {
	rows, err := db.Query(`
		SELECT * FROM conversations WHERE model_name = ?;
	`, modelName)
	if err != nil {
		return fmt.Errorf("error querying database for conversations: %v", err)
	}

	for rows.Next() {
		var id int
		var timestamp string
		var prompt string
		var response string
		var modelName string
		err := rows.Scan(&id, &timestamp, &prompt, &response, &modelName)
		if err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}

		fmt.Printf("id: %d, timestamp: %s, prompt: %s, response: %s, model_name: %s\n", id, timestamp, prompt, response, modelName)
	}

	return nil
}

func QueryConversationsForResponseStr(db *sql.DB, response string) error {
	rows, err := db.Query(`
		SELECT * FROM conversations WHERE response LIKE ?;
	`, response)
	if err != nil {
		return fmt.Errorf("error querying database for conversations: %v", err)
	}

	for rows.Next() {
		var id int
		var timestamp string
		var prompt string
		var response string
		var modelName string
		err := rows.Scan(&id, &timestamp, &prompt, &response, &modelName)
		if err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}

		fmt.Printf("id: %d, timestamp: %s, prompt: %s, response: %s, model_name: %s\n", id, timestamp, prompt, response, modelName)
	}

	return nil
}

func QueryConversationsForPromptStr(db *sql.DB, prompt string) error {
	rows, err := db.Query(`
		SELECT * FROM conversations WHERE prompt LIKE ?;
	`, prompt)
	if err != nil {
		return fmt.Errorf("error querying database for conversations: %v", err)
	}

	for rows.Next() {
		var id int
		var timestamp string
		var prompt string
		var response string
		var modelName string
		err := rows.Scan(&id, &timestamp, &prompt, &response, &modelName)
		if err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}

		fmt.Printf("id: %d, timestamp: %s, prompt: %s, response: %s, model_name: %s\n", id, timestamp, prompt, response, modelName)
	}

	return nil
}

func QueryConversationsBySearchStr(db *sql.DB, search string) error {
	rows, err := db.Query(`
		SELECT * FROM conversations WHERE prompt LIKE ? OR response LIKE ? OR model_name LIKE ?;
	`, search, search, search)
	if err != nil {
		return fmt.Errorf("error querying database for conversations: %v", err)
	}

	for rows.Next() {
		var id int
		var timestamp string
		var prompt string
		var response string
		var modelName string
		err := rows.Scan(&id, &timestamp, &prompt, &response, &modelName)
		if err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}

		fmt.Printf("id: %d, timestamp: %s, prompt: %s, response: %s, model_name: %s\n", id, timestamp, prompt, response, modelName)
	}

	return nil
}

func QueryConversationsByPrompt(db *sql.DB, prompt string) error {
	rows, err := db.Query(`
		SELECT * FROM conversations WHERE prompt = ?;
	`, prompt)
	if err != nil {
		return fmt.Errorf("error querying database for conversations: %v", err)
	}

	for rows.Next() {
		var id int
		var timestamp string
		var prompt string
		var response string
		var modelName string
		err := rows.Scan(&id, &timestamp, &prompt, &response, &modelName)
		if err != nil {
			return fmt.Errorf("Error scanning row: %v", err)
		}

		fmt.Printf("id: %d, timestamp: %s, prompt: %s, response: %s, model_name: %s\n", id, timestamp, prompt, response, modelName)
	}

	return nil
}

func CloseDB(db *sql.DB) {
	err := db.Close()
	if err != nil {
		log.Fatalf("error closing database: %v", err)
	}
}

func DeleteConversations(db *sql.DB) {
	_, err := db.Exec(`
		DELETE FROM conversations;
	`)
	if err != nil {
		log.Fatalf("error deleting conversations from database: %v", err)
	}
}

func DeleteConversationsTable(db *sql.DB) {
	_, err := db.Exec(`
		DROP TABLE conversations;
	`)
	if err != nil {
		log.Fatalf("error deleting conversations table from database: %v", err)
	}
}

func DeleteDB(dbPath string) {
	err := deleteFile(dbPath)
	if err != nil {
		log.Fatalf("error deleting database: %v", err)
	}
}

func deleteFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("error deleting file: %v", err)
	}
	return nil
}
