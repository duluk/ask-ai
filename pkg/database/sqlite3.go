package database

// Use an sqlite3 database to store the data from the LLM output
// The database will have the following tables:
// 1. interactions: store the prompt, response, and model name
// 2. FUTURE: models: store the model name and the model's parameters

// Use this module like this:
// db := NewDB("path/to/database.db")
// InsertInteraction(db, "prompt", "response", "model_name")

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

func NewDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("Error opening database: %v", err)
	}

	createInteractionsTable(db)

	return db, nil
}

func createInteractionsTable(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS interactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			prompt TEXT NOT NULL,
			response TEXT NOT NULL,
			model_name TEXT
		);
	`)
	if err != nil {
		log.Fatalf("Error creating interactions table: %v", err)
	}
}

func InsertInteraction(db *sql.DB, prompt, response, modelName string) {
	_, err := db.Exec(`
		INSERT INTO interactions (prompt, response, model_name)
		VALUES (?, ?, ?);
	`, prompt, response, modelName)
	if err != nil {
		log.Fatalf("Error inserting interaction into database: %v", err)
	}
}

// TODO: some serious refactoring can be done with these query functions
func QueryAllInteractions(db *sql.DB) {
	rows, err := db.Query(`
		SELECT * FROM interactions;
	`)
	if err != nil {
		log.Fatalf("Error querying database for interactions: %v", err)
	}

	for rows.Next() {
		var id int
		var timestamp string
		var prompt string
		var response string
		var modelName string
		err := rows.Scan(&id, &timestamp, &prompt, &response, &modelName)
		if err != nil {
			log.Fatalf("Error scanning row: %v", err)
		}

		fmt.Printf("id: %d, timestamp: %s, prompt: %s, response: %s, model_name: %s\n", id, timestamp, prompt, response, modelName)
	}
}

func QueryInteractionsByModel(db *sql.DB, modelName string) {
	rows, err := db.Query(`
		SELECT * FROM interactions WHERE model_name = ?;
	`, modelName)
	if err != nil {
		log.Fatalf("Error querying database for interactions: %v", err)
	}

	for rows.Next() {
		var id int
		var timestamp string
		var prompt string
		var response string
		var modelName string
		err := rows.Scan(&id, &timestamp, &prompt, &response, &modelName)
		if err != nil {
			log.Fatalf("Error scanning row: %v", err)
		}

		fmt.Printf("id: %d, timestamp: %s, prompt: %s, response: %s, model_name: %s\n", id, timestamp, prompt, response, modelName)
	}
}

func QueryInteractionsForResponseStr(db *sql.DB, response string) {
	rows, err := db.Query(`
		SELECT * FROM interactions WHERE response LIKE ?;
	`, response)
	if err != nil {
		log.Fatalf("Error querying database for interactions: %v", err)
	}

	for rows.Next() {
		var id int
		var timestamp string
		var prompt string
		var response string
		var modelName string
		err := rows.Scan(&id, &timestamp, &prompt, &response, &modelName)
		if err != nil {
			log.Fatalf("Error scanning row: %v", err)
		}

		fmt.Printf("id: %d, timestamp: %s, prompt: %s, response: %s, model_name: %s\n", id, timestamp, prompt, response, modelName)
	}
}

func QueryInteractionsForPromptStr(db *sql.DB, prompt string) {
	rows, err := db.Query(`
		SELECT * FROM interactions WHERE prompt LIKE ?;
	`, prompt)
	if err != nil {
		log.Fatalf("Error querying database for interactions: %v", err)
	}

	for rows.Next() {
		var id int
		var timestamp string
		var prompt string
		var response string
		var modelName string
		err := rows.Scan(&id, &timestamp, &prompt, &response, &modelName)
		if err != nil {
			log.Fatalf("Error scanning row: %v", err)
		}

		fmt.Printf("id: %d, timestamp: %s, prompt: %s, response: %s, model_name: %s\n", id, timestamp, prompt, response, modelName)
	}
}

func QueryInteractionsBySearchStr(db *sql.DB, search string) {
	rows, err := db.Query(`
		SELECT * FROM interactions WHERE prompt LIKE ? OR response LIKE ? OR model_name LIKE ?;
	`, search, search, search)
	if err != nil {
		log.Fatalf("Error querying database for interactions: %v", err)
	}

	for rows.Next() {
		var id int
		var timestamp string
		var prompt string
		var response string
		var modelName string
		err := rows.Scan(&id, &timestamp, &prompt, &response, &modelName)
		if err != nil {
			log.Fatalf("Error scanning row: %v", err)
		}

		fmt.Printf("id: %d, timestamp: %s, prompt: %s, response: %s, model_name: %s\n", id, timestamp, prompt, response, modelName)
	}
}

func QueryInteractionsByPrompt(db *sql.DB, prompt string) {
	rows, err := db.Query(`
		SELECT * FROM interactions WHERE prompt = ?;
	`, prompt)
	if err != nil {
		log.Fatalf("Error querying database for interactions: %v", err)
	}

	for rows.Next() {
		var id int
		var timestamp string
		var prompt string
		var response string
		var modelName string
		err := rows.Scan(&id, &timestamp, &prompt, &response, &modelName)
		if err != nil {
			log.Fatalf("Error scanning row: %v", err)
		}

		fmt.Printf("id: %d, timestamp: %s, prompt: %s, response: %s, model_name: %s\n", id, timestamp, prompt, response, modelName)
	}
}

func CloseDB(db *sql.DB) {
	err := db.Close()
	if err != nil {
		log.Fatalf("Error closing database: %v", err)
	}
}

func DeleteInteractions(db *sql.DB) {
	_, err := db.Exec(`
		DELETE FROM interactions;
	`)
	if err != nil {
		log.Fatalf("Error deleting interactions from database: %v", err)
	}
}

func DeleteInteractionsTable(db *sql.DB) {
	_, err := db.Exec(`
		DROP TABLE interactions;
	`)
	if err != nil {
		log.Fatalf("Error deleting interactions table from database: %v", err)
	}
}

func DeleteDB(dbPath string) {
	err := deleteFile(dbPath)
	if err != nil {
		log.Fatalf("Error deleting database: %v", err)
	}
}

func deleteFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("Error deleting file: %v", err)
	}
	return nil
}
