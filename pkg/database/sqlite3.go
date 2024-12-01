package database

// Use this module like this:
// db := NewDB("path/to/database.db")
// InsertConversation(db, "prompt", "response", "model_name", model.temp)

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/duluk/ask-ai/pkg/LLM"
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

// Return LLMConversations for a given conv_id. This will require generating
// the LLMConversations from a single row in the DB as the LLMConversations
// structure has one etnry for the user role with prompt and another with the
// assistant role and response.
func (sqlDB *ChatDB) LoadConversationFromDB(convID int) ([]LLM.LLMConversations, error) {
	rows, err := sqlDB.db.Query(`
		SELECT prompt, response, model_name, timestamp, temperature, input_tokens, output_tokens, conv_id
		FROM `+sqlDB.dbTable+` WHERE conv_id = ?;
	`, convID)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	defer rows.Close()

	var row struct {
		prompt       string
		response     string
		modelName    string
		timestamp    string
		temperature  float32
		inputTokens  int32
		outputTokens int32
		convID       int
	}
	var conversations []LLM.LLMConversations
	for rows.Next() {
		err := rows.Scan(&row.prompt, &row.response, &row.modelName, &row.timestamp, &row.temperature, &row.inputTokens, &row.outputTokens, &row.convID)
		if err != nil {
			return nil, fmt.Errorf("%v", err)
		}

		userTurn := LLM.LLMConversations{
			Role:         "user",
			Content:      row.prompt,
			Model:        row.modelName,
			Timestamp:    row.timestamp,
			InputTokens:  row.inputTokens,
			OutputTokens: 0,
			ConvID:       row.convID,
		}
		conversations = append(conversations, userTurn)

		assistantTurn := LLM.LLMConversations{
			Role:         "assistant",
			Content:      row.response,
			Model:        row.modelName,
			Timestamp:    row.timestamp,
			InputTokens:  row.inputTokens,
			OutputTokens: row.outputTokens,
			ConvID:       row.convID,
		}
		conversations = append(conversations, assistantTurn)
	}

	return conversations, nil
}

// TODO: probably want a different return structure, so that the ID and
// response at the minimum can be returned. But may want prompt too. May want
// everything.
func (sqlDB *ChatDB) SearchResponse(keyword string) ([]string, error) {
	rows, err := sqlDB.db.Query(`
		SELECT response FROM `+sqlDB.dbTable+` WHERE response LIKE ?;
	`, "%"+keyword+"%")
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	defer rows.Close()

	var responses []string
	for rows.Next() {
		var response string
		err := rows.Scan(&response)
		if err != nil {
			return nil, fmt.Errorf("%v", err)
		}
		responses = append(responses, response)
	}

	return responses, nil
}
