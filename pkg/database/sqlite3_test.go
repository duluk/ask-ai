package database

import (
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

func TestNewDB(t *testing.T) {
	db, err := NewDB(dbPath, dbTable)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	db.Close()
}

func TestInsertConversation(t *testing.T) {
	RemoveDB()
	db, err := NewDB(dbPath, dbTable)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	err = db.InsertConversation("prompt", "response", "model_name", 0.5, 10, 20, 1)
	assert.Nil(t, err)

	db.Close()
}

func TestClose(t *testing.T) {
	db, err := NewDB(dbPath, dbTable)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	db.Close()
}

func RemoveDB() {
	os.Remove(dbPath)
}

func TestInsertConversationWithError(t *testing.T) {
	RemoveDB()
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
}
