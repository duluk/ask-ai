package database

import (
	"io"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

const dbPath = "./test.db"
const dbTable = "conversations_test"

func TestMain(m *testing.M) {
	code := m.Run()

	os.Remove(dbPath)

	os.Exit(code)
}

// TestGetLastConversationID verifies that GetLastConversationID returns the highest conv_id
func TestGetLastConversationID(t *testing.T) {
	db, err := NewDB(dbPath, dbTable)
	assert.Nil(t, err)
	defer func() { db.Close(); RemoveDB() }()

	// No conversations yet
	id, err := db.GetLastConversationID()
	assert.Nil(t, err)
	assert.Equal(t, 0, id)

	// Insert conversations with increasing conv_id
	err = db.InsertConversation("p1", "r1", "m1", 0.1, 1, 1, 3)
	assert.Nil(t, err)
	err = db.InsertConversation("p2", "r2", "m2", 0.2, 2, 2, 5)
	assert.Nil(t, err)

	id, err = db.GetLastConversationID()
	assert.Nil(t, err)
	assert.Equal(t, 5, id)
}

// TestGetModel verifies that GetModel returns the model name for a given conv_id
func TestGetModel(t *testing.T) {
	db, err := NewDB(dbPath, dbTable)
	assert.Nil(t, err)
	defer func() { db.Close(); RemoveDB() }()

	// Insert a conversation
	err = db.InsertConversation("prompt", "response", "modelX", 0.5, 10, 20, 42)
	assert.Nil(t, err)

	// Existing conv_id
	model, err := db.GetModel(42)
	assert.Nil(t, err)
	assert.Equal(t, "modelX", model)

	// Non-existent conv_id
	model, err = db.GetModel(99)
	assert.Nil(t, err)
	assert.Equal(t, "", model)
}

// TestShowConversation verifies that ShowConversation prints the conversation details
func TestShowConversation(t *testing.T) {
	db, err := NewDB(dbPath, dbTable)
	assert.Nil(t, err)
	defer func() { db.Close(); RemoveDB() }()

	// Insert a conversation
	err = db.InsertConversation("prompt", "response", "model_name", 0.5, 10, 20, 7)
	assert.Nil(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	assert.Nil(t, err)
	os.Stdout = w

	db.ShowConversation(7)

	w.Close()
	os.Stdout = oldStdout

	output, err := io.ReadAll(r)
	assert.Nil(t, err)
	outStr := string(output)
	assert.Contains(t, outStr, "Prompt: prompt")
	assert.Contains(t, outStr, "Response: response")
	assert.Contains(t, outStr, "Model: model_name")
	assert.Contains(t, outStr, "Temperature: 0.500000")
	assert.Contains(t, outStr, "Input tokens: 10")
	assert.Contains(t, outStr, "Output tokens: 20")
	assert.Contains(t, outStr, "Conversation ID: 7")
}

func TestNewDB(t *testing.T) {
	db, err := NewDB(dbPath, dbTable)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	db.Close()
}

func TestInsertConversation(t *testing.T) {
	db, err := NewDB(dbPath, dbTable)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	err = db.InsertConversation("prompt", "response", "model_name", 0.5, 10, 20, 1)
	assert.Nil(t, err)

	db.Close()
	RemoveDB()
}

func TestClose(t *testing.T) {
	db, err := NewDB(dbPath, dbTable)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	db.Close()
	RemoveDB()
}

func TestInsertConversationWithError(t *testing.T) {
	db, err := NewDB(dbPath, dbTable)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	// Insert a conversation with invalid data and assert there are errors
	// TODO: This won't do much at this point because InsertConversation doesn't do
	// much validation of iput, and Go's type system won't let me enter an
	// invalid argument. InsertConversation should, however, do some
	// validation. For instance, there are restrictions about temperature - eg,
	// 0.123 is technically invalid.
	// err = db.InsertConversation("prompt", "response", "", 0.0)
	// assert.NotNil(t, err)

	db.Close()
	RemoveDB()
}

func TestLoadConversations(t *testing.T) {
	db, err := NewDB(dbPath, dbTable)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	err = db.InsertConversation("prompt", "response", "model_name", 0.5, 10, 20, 1)
	assert.Nil(t, err)
	err = db.InsertConversation("prompt2", "response2", "model_name2", 0.5, 10, 20, 2)
	assert.Nil(t, err)

	conversations, err := db.LoadConversationFromDB(1)
	assert.Nil(t, err)
	// Len 2 because one prompt/response turn in the DB is one row; however,
	// it's two LLMConversations, which LoadConversationFromDB does.
	assert.Len(t, conversations, 2)

	db.Close()
	RemoveDB()
}

func TestSearchForConversation(t *testing.T) {
	db, err := NewDB(dbPath, dbTable)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	err = db.InsertConversation("prompt", "response", "model_name", 0.5, 10, 20, 1)
	assert.Nil(t, err)
	err = db.InsertConversation("prompt2", "response2", "model_name2", 0.5, 10, 20, 2)
	assert.Nil(t, err)

	ids, err := db.SearchForConversation("response")
	assert.Nil(t, err)
	assert.Len(t, ids, 2)
	assert.Equal(t, 1, ids[0])

	ids, err = db.SearchForConversation("marklar")
	assert.Nil(t, err)
	assert.Len(t, ids, 0)
	// assert.Equal(t, 1, ids[0])

	db.Close()
	RemoveDB()
}

func RemoveDB() {
	os.Remove(dbPath)
}
