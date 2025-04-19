package tui

import (
   "testing"

   tea "github.com/charmbracelet/bubbletea"
   "github.com/stretchr/testify/assert"

   "github.com/duluk/ask-ai/pkg/config"
   "github.com/duluk/ask-ai/pkg/database"
   "github.com/duluk/ask-ai/pkg/LLM"
)

// Test that Initialize returns a Model in a non-ready state and that View
// shows the initializing message.
func TestInitializeView(t *testing.T) {
   // Prepare minimal options
   opts := &config.Options{
       ScreenWidth:     100,
       ScreenTextWidth: 80,
       ScreenHeight:    40,
       TabWidth:        4,
   }
   // Minimal client args with non-nil pointers for Model and ConvID
   modelName := "testModel"
   convID := 99
   clientArgs := LLM.ClientArgs{
       Model:  &modelName,
       ConvID: &convID,
   }
   // In-memory database for testing
   db, err := database.InitializeDB(":memory:", "tui_test")
   assert.NoError(t, err)
   defer db.Close()

   m := Initialize(opts, clientArgs, db)
   // Should not be ready before any WindowSizeMsg
   assert.False(t, m.ready)
   // Initial View should show initializing text
   assert.Equal(t, "Initializing...", m.View())
}

// Test that sending a WindowSizeMsg sets the model to ready and updates dimensions
func TestUpdateWindowSize(t *testing.T) {
   opts := &config.Options{
       ScreenWidth:     120,
       ScreenTextWidth: 100,
       ScreenHeight:    50,
       TabWidth:        4,
   }
   modelName := "m"
   convID := 1
   clientArgs := LLM.ClientArgs{
       Model:  &modelName,
       ConvID: &convID,
   }
   db, err := database.InitializeDB(":memory:", "tui_test2")
   assert.NoError(t, err)
   defer db.Close()

   m := Initialize(opts, clientArgs, db)
   // Send a resize message
   width := 80
   height := 30
   m2i, cmd := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
   // No command expected on resize
   assert.Nil(t, cmd)
   m2 := m2i.(Model)
   // Model should now be ready
   assert.True(t, m2.ready)
   // Expected viewport dimensions
   totalFixed := inputHeight + statusHeight + borderHeight + contentPadding + testPadding
   expHeight := height - totalFixed
   expWidth := width - contentMargin
   // Verify viewport size and input width
   assert.Equal(t, expWidth, m2.viewport.Width)
   assert.Equal(t, expHeight, m2.viewport.Height)
   assert.Equal(t, expWidth, m2.textInput.Width)
}