package database

// Use this module like this:
// db := NewDB("path/to/database.db")
// InsertConversation(db, "prompt", "response", "model_name", model.temp)

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type ChatDB struct {
	db      *sql.DB
	dbTable string
}

// Retun errors to the caller in case we want to ignore them. That is, just
// because we can't store the conversations in the database doesn't mean we
// should stop the program.
func NewDB(dbPath string, dbTable string) (*ChatDB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	err = createTable(db, dbTable)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	sqlDB := ChatDB{}
	sqlDB.db = db
	sqlDB.dbTable = dbTable
	return &sqlDB, nil
}

// TODO: when I add the ability to get the number of tokens used in the
// response, store that
func createTable(db *sql.DB, dbTable string) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS ` + dbTable + ` (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			prompt TEXT NOT NULL,
			response TEXT NOT NULL,
			model_name TEXT NOT NULL,
			temperature REAL NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("error creating %s table: %v", dbTable, err)
	}

	return nil
}

func (sqlDB *ChatDB) InsertConversation(
	prompt,
	response,
	modelName string,
	temperature float32,
) error {
	_, err := sqlDB.db.Exec(`
		INSERT INTO `+sqlDB.dbTable+` (prompt, response, model_name, temperature)
		VALUES (?, ?, ?, ?);
	`, prompt, response, modelName, temperature)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}

func (sqlDB *ChatDB) Close() {
	err := sqlDB.db.Close()
	if err != nil {
		log.Fatalf("error closing database: %v", err)
	}
}
