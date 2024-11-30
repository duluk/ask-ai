package database

// Use this module like this:
// db := NewDB("path/to/database.db")
// InsertConversation(db, "prompt", "response", "model_name", model.temp)

import (
	"database/sql"
	"fmt"
	"log"

	// "github.com/duluk/ask-ai/pkg/LLM"
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

	_, err = db.Exec(DBSchema(dbTable))
	if err != nil {
		return nil, fmt.Errorf("error creating %s table: %v", dbTable, err)
	}

	sqlDB := ChatDB{}
	sqlDB.db = db
	sqlDB.dbTable = dbTable
	return &sqlDB, nil
}

func (sqlDB *ChatDB) InsertConversation(
	prompt,
	response,
	modelName string,
	temperature float32,
	inputTokens int32,
	outputTokens int32,
	convID int,
) error {
	_, err := sqlDB.db.Exec(`
		INSERT INTO `+sqlDB.dbTable+` (prompt, response, model_name, temperature, input_tokens, output_tokens, conv_id)
		VALUES (?, ?, ?, ?, ?, ?, ?);
	`, prompt, response, modelName, temperature, inputTokens, outputTokens, convID)
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

// TODO: the problem to be solved here is that the DB stores the conversations
// slightly differently from the YML file. For the YML file, the prompt and the
// response are two different entries, which is what LLMConversations
// represents. However, in the DB, the prompt and the response are stored in
// the same row. So, we need to figure out how to load the conversations from
// the DB and return them as LLMConversations.
// One options is to read the DB etnry and create two LLMConversations from it.

// func (sqlDB *ChatDB) LoadConversation(convID int) ([]LLM.LLMConversations, error) {
// 	rows, err := sqlDB.db.Query(`
// 		SELECT prompt, response, model_name, temperature, input_tokens, output_tokens, conv_id
// 		FROM `+sqlDB.dbTable+` WHERE conv_id = ?;
// 	`, convID)
// 	if err != nil {
// 		return nil, fmt.Errorf("%v", err)
// 	}
// 	defer rows.Close()
//
// 	var conversations []LLM.LLMConversations
// 	for rows.Next() {
// 		var conv LLM.LLMConversations
// 		// err := rows.Scan(&conv.Content, &conv.Response, &conv.ModelName, &conv.Temperature, &conv.InputTokens, &conv.OutputTokens, &conv.ConvID)
// 		if err != nil {
// 			return nil, fmt.Errorf("%v", err)
// 		}
// 		conversations = append(conversations, conv)
// 	}
//
// 	return conversations, nil
// }
