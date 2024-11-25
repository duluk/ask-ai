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

	// Clean up the test database
	os.Remove(dbPath)

	os.Exit(code)
}

func TestNewDB(t *testing.T) {
	// Create a new DB and assert that it's not nil
	db, err := NewDB(dbPath, dbTable)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	// Close the DB to clean up
	db.Close()
}

func TestCreateTable(t *testing.T) {
	// Create a new DB and assert that it's not nil
	db, err := NewDB(dbPath, dbTable)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	// Call createTable on the existing DB to assert it doesn't error
	err = createTable(db.db, dbTable)
	assert.Nil(t, err)

	// Close the DB to clean up
	db.Close()
}

func TestInsertConversation(t *testing.T) {
	// Create a new DB and assert that it's not nil
	db, err := NewDB(dbPath, dbTable)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	// Insert a conversation with valid data and assert there are no errors
	err = db.InsertConversation("prompt", "response", "model_name", 0.5)
	assert.Nil(t, err)

	// Close the DB to clean up
	db.Close()
}

func TestClose(t *testing.T) {
	// Create a new DB and assert that it's not nil
	db, err := NewDB(dbPath, dbTable)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	// Close the DB to clean up
	db.Close()
}

func TestInsertConversationWithError(t *testing.T) {
	// Create a new DB and assert that it's not nil
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

	// Close the DB to clean up
	db.Close()
}
